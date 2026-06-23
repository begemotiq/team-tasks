package main

import (
	"context"
	"database/sql"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"time"

	"task-service/internal/config"
	"task-service/internal/domain"
	"task-service/internal/domain/models"
	mysqlinfra "task-service/internal/infrastructure/mysql"
	redisinfra "task-service/internal/infrastructure/redis"
	mysqlrepo "task-service/internal/repository/mysql"
	redisrepo "task-service/internal/repository/redis"
	"task-service/internal/usecase"
)

const seedPassword = "password123"

type seedUser struct {
	Email string
	Name  string
}

type seedTask struct {
	Title       string
	Description string
	Status      models.TaskStatus
	Creator     *models.User
	Assignee    *models.User
	DueDate     *time.Time
}

func main() {
	if err := run(); err != nil {
		slog.Default().Error("seed failed", "error", err)
		os.Exit(1)
	}
}

func run() error {
	configPath := flag.String("config", os.Getenv("CONFIG_PATH"), "path to YAML config")
	flag.Parse()

	cfg, err := config.Load(*configPath)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	db, err := mysqlinfra.NewDB(ctx, cfg.Database)
	if err != nil {
		return err
	}
	defer func() {
		if err := db.Close(); err != nil {
			slog.Default().Warn("close mysql connection failed", "error", err)
		}
	}()

	usersRepo := mysqlrepo.NewUserRepository(db)
	teamsRepo := mysqlrepo.NewTeamRepository(db)
	tasksRepo := mysqlrepo.NewTaskRepository(db)
	hasher := usecase.BcryptHasher{}

	owner, err := ensureUser(ctx, db, usersRepo, hasher, seedUser{Email: "owner@example.com", Name: "Owner"})
	if err != nil {
		return err
	}
	admin, err := ensureUser(ctx, db, usersRepo, hasher, seedUser{Email: "admin@example.com", Name: "Admin"})
	if err != nil {
		return err
	}
	member, err := ensureUser(ctx, db, usersRepo, hasher, seedUser{Email: "member@example.com", Name: "Member"})
	if err != nil {
		return err
	}
	outsider, err := ensureUser(ctx, db, usersRepo, hasher, seedUser{Email: "outsider@example.com", Name: "Outsider"})
	if err != nil {
		return err
	}

	backend, err := ensureTeam(ctx, teamsRepo, owner, "Backend")
	if err != nil {
		return err
	}
	if err := ensureMember(ctx, db, teamsRepo, backend.ID, admin.ID, models.RoleAdmin); err != nil {
		return err
	}
	if err := ensureMember(ctx, db, teamsRepo, backend.ID, member.ID, models.RoleMember); err != nil {
		return err
	}

	product, err := ensureTeam(ctx, teamsRepo, admin, "Product")
	if err != nil {
		return err
	}
	if err := ensureMember(ctx, db, teamsRepo, product.ID, owner.ID, models.RoleAdmin); err != nil {
		return err
	}

	tomorrow := time.Now().UTC().Add(24 * time.Hour).Truncate(time.Second)
	nextWeek := time.Now().UTC().Add(7 * 24 * time.Hour).Truncate(time.Second)
	tasks := []struct {
		Team models.Team
		Task seedTask
	}{
		{Team: *backend, Task: seedTask{Title: "Implement API", Description: "Check create/list/update endpoints", Status: models.TaskStatusTodo, Creator: owner, Assignee: admin, DueDate: &tomorrow}},
		{Team: *backend, Task: seedTask{Title: "Wire Redis cache", Description: "Validate cache hit/miss metrics", Status: models.TaskStatusInProgress, Creator: admin, Assignee: member, DueDate: &nextWeek}},
		{Team: *backend, Task: seedTask{Title: "Review RBAC", Description: "Owner/admin/member permissions smoke data", Status: models.TaskStatusDone, Creator: member, Assignee: owner}},
		{Team: *product, Task: seedTask{Title: "Prepare roadmap", Description: "Product team task for reports", Status: models.TaskStatusTodo, Creator: admin, Assignee: owner, DueDate: &nextWeek}},
	}
	for _, item := range tasks {
		if _, err := ensureTask(ctx, tasksRepo, item.Team.ID, item.Task); err != nil {
			return err
		}
	}

	invalidateCache(ctx, cfg, backend.ID, product.ID)
	printSeedSummary(owner, admin, member, outsider, backend, product)
	return nil
}

func ensureUser(ctx context.Context, db *sql.DB, repo *mysqlrepo.UserRepository, hasher usecase.PasswordHasher, seed seedUser) (*models.User, error) {
	hash, err := hasher.Hash(seedPassword)
	if err != nil {
		return nil, err
	}
	user, err := repo.FindByEmail(ctx, seed.Email)
	if err == nil {
		if _, err := db.ExecContext(ctx, "UPDATE users SET password_hash = ?, name = ? WHERE id = ?", hash, seed.Name, user.ID); err != nil {
			return nil, err
		}
		return repo.FindByID(ctx, user.ID)
	}
	if !errors.Is(err, domain.ErrNotFound) {
		return nil, err
	}
	user = &models.User{
		Email:        seed.Email,
		PasswordHash: hash,
		Name:         seed.Name,
	}
	if err := repo.Create(ctx, user); err != nil {
		return nil, err
	}
	return user, nil
}

func ensureTeam(ctx context.Context, repo *mysqlrepo.TeamRepository, owner *models.User, name string) (*models.Team, error) {
	teams, err := repo.ListByUser(ctx, owner.ID)
	if err != nil {
		return nil, err
	}
	for i := range teams {
		if teams[i].Name == name && teams[i].CreatedBy == owner.ID {
			return &teams[i], nil
		}
	}
	team := &models.Team{Name: name}
	if err := repo.CreateWithOwner(ctx, team, owner.ID); err != nil {
		return nil, err
	}
	return team, nil
}

func ensureMember(ctx context.Context, db *sql.DB, repo *mysqlrepo.TeamRepository, teamID, userID int64, role models.Role) error {
	currentRole, err := repo.GetMemberRole(ctx, teamID, userID)
	if err == nil {
		if currentRole == role {
			return nil
		}
		_, err := db.ExecContext(ctx, "UPDATE team_members SET role = ? WHERE team_id = ? AND user_id = ?", role, teamID, userID)
		return err
	}
	if !errors.Is(err, domain.ErrNotFound) {
		return err
	}
	return repo.AddMember(ctx, teamID, userID, role)
}

func ensureTask(ctx context.Context, repo *mysqlrepo.TaskRepository, teamID int64, seed seedTask) (*models.Task, error) {
	filter := models.TaskFilter{TeamID: &teamID, PageSize: 100}
	list, err := repo.List(ctx, filter, seed.Creator.ID)
	if err != nil {
		return nil, err
	}
	for i := range list.Items {
		if list.Items[i].Title == seed.Title {
			return &list.Items[i], nil
		}
	}
	task := &models.Task{
		Title:       seed.Title,
		Description: seed.Description,
		Status:      seed.Status,
		TeamID:      teamID,
		CreatedBy:   seed.Creator.ID,
		DueDate:     seed.DueDate,
	}
	if seed.Assignee != nil {
		task.AssigneeID = &seed.Assignee.ID
	}
	history := &models.TaskHistory{
		ChangedBy: seed.Creator.ID,
		Field:     "created",
		OldValue:  "",
		NewValue:  string(seed.Status),
	}
	if err := repo.CreateWithHistory(ctx, task, history); err != nil {
		return nil, err
	}
	return task, nil
}

func invalidateCache(ctx context.Context, cfg *config.Config, teamIDs ...int64) {
	client := redisinfra.NewClient(cfg.Redis)
	defer func() {
		if err := client.Close(); err != nil {
			slog.Default().Warn("close redis connection failed", "error", err)
		}
	}()
	if err := client.Ping(ctx).Err(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "warning: redis cache was not invalidated: %v\n", err)
		return
	}
	cache := redisrepo.NewCache(client)
	for _, teamID := range teamIDs {
		if err := cache.DeleteTeamTasks(ctx, teamID); err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "warning: team %d cache was not invalidated: %v\n", teamID, err)
		}
	}
}

func printSeedSummary(owner, admin, member, outsider *models.User, backend, product *models.Team) {
	_, _ = fmt.Fprintln(os.Stdout, "Seed data is ready.")
	_, _ = fmt.Fprintln(os.Stdout)
	_, _ = fmt.Fprintln(os.Stdout, "Users:")
	for _, user := range []*models.User{owner, admin, member, outsider} {
		_, _ = fmt.Fprintf(os.Stdout, "  %s / %s (%s, id=%d)\n", user.Email, seedPassword, user.Name, user.ID)
	}
	_, _ = fmt.Fprintln(os.Stdout)
	_, _ = fmt.Fprintln(os.Stdout, "Teams:")
	_, _ = fmt.Fprintf(os.Stdout, "  Backend id=%d, owner=%s, admin=%s, member=%s\n", backend.ID, owner.Email, admin.Email, member.Email)
	_, _ = fmt.Fprintf(os.Stdout, "  Product id=%d, owner=%s, admin=%s\n", product.ID, admin.Email, owner.Email)
}
