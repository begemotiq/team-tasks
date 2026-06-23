//go:build integration

package integration

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	"task-service/internal/config"
	"task-service/internal/domain/models"
	mysqlrepo "task-service/internal/repository/mysql"
)

const mysqlStartupTimeout = 30 * time.Second

var (
	integrationDB        *sql.DB
	integrationFixtureDB *sql.DB
	integrationContainer testcontainers.Container
	fixtureSequence      uint64
)

type repositorySet struct {
	users  *mysqlrepo.UserRepository
	teams  *mysqlrepo.TeamRepository
	tasks  *mysqlrepo.TaskRepository
	outbox *mysqlrepo.OutboxRepository
}

type mysqlFixture struct {
	t      *testing.T
	ctx    context.Context
	repos  repositorySet
	suffix string
}

func TestMain(m *testing.M) {
	ctx := context.Background()
	db, fixtureDB, container, skip, err := startMySQL(ctx)
	if err != nil {
		if skip {
			if requireDocker() {
				_, _ = fmt.Fprintf(os.Stderr, "Docker is required for MySQL integration tests: %v\n", err)
				os.Exit(1)
			}
			_, _ = fmt.Fprintf(os.Stderr, "skipping MySQL integration tests: %v\n", err)
			os.Exit(0)
		}
		_, _ = fmt.Fprintf(os.Stderr, "failed to start MySQL integration tests: %v\n", err)
		os.Exit(1)
	}

	integrationDB = db
	integrationFixtureDB = fixtureDB
	integrationContainer = container

	code := m.Run()

	_ = integrationDB.Close()
	_ = integrationFixtureDB.Close()
	_ = integrationContainer.Terminate(context.Background())
	os.Exit(code)
}

func startMySQL(ctx context.Context) (db *sql.DB, fixtureDB *sql.DB, container testcontainers.Container, skip bool, err error) {
	defer func() {
		if recovered := recover(); recovered != nil {
			err = fmt.Errorf("testcontainers provider is not available: %v", recovered)
			skip = true
		}
	}()

	container, err = testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        "mysql:8.4",
			ExposedPorts: []string{"3306/tcp"},
			Env: map[string]string{
				"MYSQL_ROOT_PASSWORD": "root",
				"MYSQL_DATABASE":      "tasks",
				"MYSQL_USER":          "tasks",
				"MYSQL_PASSWORD":      "tasks",
			},
			WaitingFor: wait.ForLog("ready for connections").WithOccurrence(2).WithStartupTimeout(2 * time.Minute),
		},
		Started: true,
	})
	if err != nil {
		return nil, nil, nil, true, err
	}

	host, err := container.Host(ctx)
	if err != nil {
		_ = container.Terminate(ctx)
		return nil, nil, nil, false, err
	}
	port, err := container.MappedPort(ctx, "3306/tcp")
	if err != nil {
		_ = container.Terminate(ctx)
		return nil, nil, nil, false, err
	}

	cfg := config.Default().Database
	cfg.Host = host
	cfg.Port = port.Int()

	fixtureDB, err = openTestDB(ctx, cfg, multiStatementTestDSN(cfg))
	if err != nil {
		_ = container.Terminate(ctx)
		return nil, nil, nil, false, err
	}
	if err := migrateWithRetry(ctx, fixtureDB); err != nil {
		_ = fixtureDB.Close()
		_ = container.Terminate(ctx)
		return nil, nil, nil, false, err
	}

	db, err = openTestDB(ctx, cfg, cfg.DSN())
	if err != nil {
		_ = fixtureDB.Close()
		_ = container.Terminate(ctx)
		return nil, nil, nil, false, err
	}

	return db, fixtureDB, container, false, nil
}

func requireDocker() bool {
	value := os.Getenv("INTEGRATION_REQUIRE_DOCKER")
	return value == "1" || strings.EqualFold(value, "true")
}

func openTestDB(ctx context.Context, cfg config.DatabaseConfig, dsn string) (*sql.DB, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.ConnMaxLifetime)

	if err := retryUntil(ctx, mysqlStartupTimeout, func(ctx context.Context) error {
		pingCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
		defer cancel()
		return db.PingContext(pingCtx)
	}); err != nil {
		closeErr := db.Close()
		if closeErr != nil {
			return nil, fmt.Errorf("%w; close db: %v", err, closeErr)
		}
		return nil, err
	}
	return db, nil
}

func multiStatementTestDSN(cfg config.DatabaseConfig) string {
	return cfg.DSN() + "&multiStatements=true"
}

func migrateWithRetry(ctx context.Context, db *sql.DB) error {
	return retryUntil(ctx, mysqlStartupTimeout, func(ctx context.Context) error {
		return migrate(ctx, db)
	})
}

func migrate(ctx context.Context, db *sql.DB) error {
	migration, err := readProjectFile("migrations", "001_init.sql")
	if err != nil {
		return err
	}
	_, err = db.ExecContext(ctx, string(migration))
	return err
}

func retryUntil(ctx context.Context, timeout time.Duration, fn func(context.Context) error) error {
	deadline := time.Now().Add(timeout)
	var lastErr error
	for {
		attemptCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
		lastErr = fn(attemptCtx)
		cancel()
		if lastErr == nil {
			return nil
		}
		if time.Now().After(deadline) {
			return lastErr
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(500 * time.Millisecond):
		}
	}
}

func readProjectFile(path ...string) ([]byte, error) {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		return nil, fmt.Errorf("resolve project root")
	}
	root := filepath.Clean(filepath.Join(filepath.Dir(filename), "..", ".."))
	return os.ReadFile(filepath.Join(append([]string{root}, path...)...))
}

func newFixture(t *testing.T) *mysqlFixture {
	t.Helper()

	sequence := atomic.AddUint64(&fixtureSequence, 1)
	suffix := fmt.Sprintf("%s-%d", sanitizeName(t.Name()), sequence)

	return &mysqlFixture{
		t:   t,
		ctx: context.Background(),
		repos: repositorySet{
			users:  mysqlrepo.NewUserRepository(integrationDB),
			teams:  mysqlrepo.NewTeamRepository(integrationDB),
			tasks:  mysqlrepo.NewTaskRepository(integrationDB),
			outbox: mysqlrepo.NewOutboxRepository(integrationDB),
		},
		suffix: suffix,
	}
}

func (f *mysqlFixture) user(label string) *models.User {
	f.t.Helper()

	user := &models.User{
		Email:        fmt.Sprintf("%s-%s@example.com", sanitizeName(label), f.suffix),
		PasswordHash: "hash",
		Name:         fmt.Sprintf("%s %s", label, f.suffix),
	}
	if err := f.repos.users.Create(f.ctx, user); err != nil {
		f.t.Fatalf("create user %q: %v", label, err)
	}
	return user
}

func (f *mysqlFixture) team(label string, owner *models.User) *models.Team {
	f.t.Helper()

	team := &models.Team{Name: fmt.Sprintf("%s %s", label, f.suffix)}
	if err := f.repos.teams.CreateWithOwner(f.ctx, team, owner.ID); err != nil {
		f.t.Fatalf("create team %q: %v", label, err)
	}
	return team
}

func (f *mysqlFixture) member(team *models.Team, user *models.User, role models.Role) {
	f.t.Helper()

	if err := f.repos.teams.AddMember(f.ctx, team.ID, user.ID, role); err != nil {
		f.t.Fatalf("add member %d to team %d: %v", user.ID, team.ID, err)
	}
}

func (f *mysqlFixture) task(label string, team *models.Team, creator *models.User, status models.TaskStatus, assignee *models.User) *models.Task {
	f.t.Helper()

	task := &models.Task{
		Title:       fmt.Sprintf("%s %s", label, f.suffix),
		Description: fmt.Sprintf("description %s", f.suffix),
		Status:      status,
		TeamID:      team.ID,
		CreatedBy:   creator.ID,
	}
	if assignee != nil {
		task.AssigneeID = &assignee.ID
	}
	if err := f.repos.tasks.Create(f.ctx, task); err != nil {
		f.t.Fatalf("create task %q: %v", label, err)
	}
	return task
}

func (f *mysqlFixture) loadSQLFixture(name string, replacements map[string]string) {
	f.t.Helper()

	content, err := readProjectFile("tests", "integration", "fixtures", name)
	if err != nil {
		f.t.Fatalf("read SQL fixture %q: %v", name, err)
	}
	query := string(content)
	for placeholder, value := range replacements {
		query = strings.ReplaceAll(query, placeholder, value)
	}
	if _, err := integrationFixtureDB.ExecContext(f.ctx, query); err != nil {
		f.t.Fatalf("load SQL fixture %q: %v", name, err)
	}
}

func (f *mysqlFixture) mustUserByEmail(email string) *models.User {
	f.t.Helper()

	user, err := f.repos.users.FindByEmail(f.ctx, email)
	if err != nil {
		f.t.Fatalf("find fixture user %q: %v", email, err)
	}
	return user
}

func (f *mysqlFixture) mustTeamByName(user *models.User, name string) *models.Team {
	f.t.Helper()

	teams, err := f.repos.teams.ListByUser(f.ctx, user.ID)
	if err != nil {
		f.t.Fatalf("list fixture teams for user %d: %v", user.ID, err)
	}
	for i := range teams {
		if teams[i].Name == name {
			return &teams[i]
		}
	}
	f.t.Fatalf("fixture team %q was not found", name)
	return nil
}

func (f *mysqlFixture) mustTaskByTitle(requester *models.User, team *models.Team, title string) *models.Task {
	f.t.Helper()

	filter := models.TaskFilter{TeamID: &team.ID, PageSize: 100}
	list, err := f.repos.tasks.List(f.ctx, filter, requester.ID)
	if err != nil {
		f.t.Fatalf("list fixture tasks for team %d: %v", team.ID, err)
	}
	for i := range list.Items {
		if list.Items[i].Title == title {
			return &list.Items[i]
		}
	}
	f.t.Fatalf("fixture task %q was not found", title)
	return nil
}

func sanitizeName(value string) string {
	value = strings.ToLower(value)
	var builder strings.Builder
	for _, char := range value {
		switch {
		case char >= 'a' && char <= 'z':
			builder.WriteRune(char)
		case char >= '0' && char <= '9':
			builder.WriteRune(char)
		default:
			builder.WriteByte('-')
		}
	}
	return strings.Trim(builder.String(), "-")
}
