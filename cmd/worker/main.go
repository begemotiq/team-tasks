package main

import (
	"context"
	"flag"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	emailadapter "task-service/internal/adapter/email"
	"task-service/internal/config"
	mysqlinfra "task-service/internal/infrastructure/mysql"
	applogger "task-service/internal/logger"
	"task-service/internal/metrics"
	mysqlrepo "task-service/internal/repository/mysql"
	outboxcleanupusecase "task-service/internal/usecase/outbox_cleanup"
	outboxdispatchusecase "task-service/internal/usecase/outbox_dispatch"
)

func main() {
	if err := run(); err != nil {
		slog.Default().Error("worker failed", "error", err)
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

	metricsServer := newMetricsServer(cfg.Metrics.Address)
	errCh := make(chan error, 1)
	if metricsServer != nil {
		go func() {
			logger.Info("worker metrics listening", "address", cfg.Metrics.Address)
			if err := metricsServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				errCh <- err
			}
		}()
	}

	db, err := mysqlinfra.NewDB(ctx, cfg.Database)
	if err != nil {
		return err
	}
	defer db.Close()

	outboxRepo := mysqlrepo.NewOutboxRepository(db)
	email := emailadapter.NewService(cfg.Email)
	dispatcher := outboxdispatchusecase.New(outboxRepo, email, cfg.Outbox.RetryDelay, cfg.Outbox.MaxAttempts)
	cleanup := outboxcleanupusecase.New(outboxRepo, cfg.Outbox.CleanupRetention)

	go runOutboxDispatcher(ctx, logger, dispatcher, cfg.Outbox)
	go runOutboxCleanup(ctx, logger, cleanup, cfg.Outbox)

	logger.Info("task worker started")
	select {
	case err := <-errCh:
		stop()
		return err
	case <-ctx.Done():
	}

	if metricsServer != nil {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.HTTP.ShutdownTimeout)
		defer cancel()
		if err := metricsServer.Shutdown(shutdownCtx); err != nil {
			return err
		}
	}
	logger.Info("task worker stopped")
	return nil
}

func newMetricsServer(address string) *http.Server {
	if address == "" {
		return nil
	}
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	return &http.Server{
		Addr:              address,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}
}

func runOutboxDispatcher(ctx context.Context, logger *slog.Logger, dispatcher *outboxdispatchusecase.UseCase, cfg config.OutboxConfig) {
	batchSize := cfg.BatchSize
	if batchSize <= 0 {
		batchSize = 10
	}
	pollInterval := cfg.PollInterval
	if pollInterval <= 0 {
		pollInterval = 5 * time.Second
	}

	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	for {
		result, err := dispatcher.Dispatch(ctx, batchSize)
		if err != nil && ctx.Err() != nil {
			return
		}
		metrics.RecordOutboxDispatch(result.Claimed, result.Processed, result.Retried, result.DeadLettered, result.ErrorStage, err)
		if err != nil {
			logger.ErrorContext(ctx, "outbox dispatch failed", "error", err)
		}

		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

func runOutboxCleanup(ctx context.Context, logger *slog.Logger, cleanup *outboxcleanupusecase.UseCase, cfg config.OutboxConfig) {
	cleanupInterval := cfg.CleanupInterval
	if cleanupInterval <= 0 {
		cleanupInterval = 21 * 24 * time.Hour
	}

	ticker := time.NewTicker(cleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}

		deleted, err := cleanup.Cleanup(ctx)
		if err != nil && ctx.Err() != nil {
			return
		}
		metrics.RecordOutboxCleanup(deleted, err)
		if err != nil {
			logger.ErrorContext(ctx, "outbox cleanup failed", "error", err)
			continue
		}
		if deleted > 0 {
			logger.InfoContext(ctx, "outbox cleanup deleted processed events", "deleted", deleted)
		}
	}
}
