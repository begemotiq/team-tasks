package main

import (
	"context"
	"flag"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	httpadapter "task-service/internal/adapter/http"
	"task-service/internal/config"
	mysqlinfra "task-service/internal/infrastructure/mysql"
	redisinfra "task-service/internal/infrastructure/redis"
	applogger "task-service/internal/logger"
	mysqlrepo "task-service/internal/repository/mysql"
	redisrepo "task-service/internal/repository/redis"
	"task-service/internal/usecase"
	authloginusecase "task-service/internal/usecase/auth_login"
	authregisterusecase "task-service/internal/usecase/auth_register"
	reportinvalidassigneesusecase "task-service/internal/usecase/report_invalid_assignees"
	reportteamsummaryusecase "task-service/internal/usecase/report_team_summary"
	reporttopcreatorsusecase "task-service/internal/usecase/report_top_creators"
	taskcreateusecase "task-service/internal/usecase/task_create"
	taskhistoryusecase "task-service/internal/usecase/task_history"
	tasklistusecase "task-service/internal/usecase/task_list"
	taskupdateusecase "task-service/internal/usecase/task_update"
	teamcreateusecase "task-service/internal/usecase/team_create"
	teamdeleteusecase "task-service/internal/usecase/team_delete"
	teaminviteusecase "task-service/internal/usecase/team_invite"
	teamlistusecase "task-service/internal/usecase/team_list"
)

func main() {
	if err := run(); err != nil {
		slog.Default().Error("api failed", "error", err)
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
	logger, err := applogger.New(os.Stdout, applogger.Config{
		Level:  cfg.Logging.Level,
		Format: cfg.Logging.Format,
	})
	if err != nil {
		return err
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	db, err := mysqlinfra.NewDB(ctx, cfg.Database)
	if err != nil {
		return err
	}
	defer db.Close()

	redisClient := redisinfra.NewClient(cfg.Redis)
	if err := redisClient.Ping(ctx).Err(); err != nil {
		return err
	}
	defer redisClient.Close()

	usersRepo := mysqlrepo.NewUserRepository(db)
	teamsRepo := mysqlrepo.NewTeamRepository(db)
	tasksRepo := mysqlrepo.NewTaskRepository(db)
	cache := redisrepo.NewCache(redisClient)
	tokens := usecase.JWTManager{Secret: []byte(cfg.JWT.Secret), TTL: cfg.JWT.TTL}
	hasher := usecase.BcryptHasher{}

	authRegisterUseCase := authregisterusecase.New(usersRepo, hasher, tokens)
	authLoginUseCase := authloginusecase.New(usersRepo, hasher, tokens)
	teamCreateUseCase := teamcreateusecase.New(teamsRepo)
	teamDeleteUseCase := teamdeleteusecase.New(teamsRepo, cache)
	teamListUseCase := teamlistusecase.New(teamsRepo)
	teamInviteUseCase := teaminviteusecase.New(teamsRepo, usersRepo)
	taskCreateUseCase := taskcreateusecase.New(tasksRepo, teamsRepo, cache)
	taskListUseCase := tasklistusecase.New(tasksRepo, teamsRepo, cache)
	taskUpdateUseCase := taskupdateusecase.New(tasksRepo, teamsRepo, cache)
	taskHistoryUseCase := taskhistoryusecase.New(tasksRepo, teamsRepo)
	reportTeamSummaryUseCase := reportteamsummaryusecase.New(tasksRepo, teamsRepo)
	reportTopCreatorsUseCase := reporttopcreatorsusecase.New(tasksRepo, teamsRepo)
	reportInvalidAssigneesUseCase := reportinvalidassigneesusecase.New(tasksRepo, teamsRepo)

	router := httpadapter.NewRouter(httpadapter.Dependencies{
		AuthRegister:           authRegisterUseCase,
		AuthLogin:              authLoginUseCase,
		TeamCreate:             teamCreateUseCase,
		TeamDelete:             teamDeleteUseCase,
		TeamList:               teamListUseCase,
		TeamInvite:             teamInviteUseCase,
		TaskCreate:             taskCreateUseCase,
		TaskList:               taskListUseCase,
		TaskUpdate:             taskUpdateUseCase,
		TaskHistory:            taskHistoryUseCase,
		ReportTeamSummary:      reportTeamSummaryUseCase,
		ReportTopCreators:      reportTopCreatorsUseCase,
		ReportInvalidAssignees: reportInvalidAssigneesUseCase,
		Tokens:                 tokens,
		RequestsPerMinute:      cfg.RateLimit.RequestsPerMinute,
		Logger:                 logger,
	})

	server := &http.Server{
		Addr:         cfg.HTTP.Address,
		Handler:      router,
		ReadTimeout:  cfg.HTTP.ReadTimeout,
		WriteTimeout: cfg.HTTP.WriteTimeout,
	}

	errCh := make(chan error, 1)
	go func() {
		logger.Info("task service listening", "address", cfg.HTTP.Address)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.HTTP.ShutdownTimeout)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		return err
	}
	logger.Info("task service stopped")
	return nil
}
