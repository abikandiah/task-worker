package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/abikandiah/task-worker/internal/factory"
	"github.com/abikandiah/task-worker/internal/mock"
	"github.com/abikandiah/task-worker/internal/platform/logging"
	"github.com/abikandiah/task-worker/internal/platform/server"
	"github.com/abikandiah/task-worker/internal/service"
)

// Entry point for HTTP server
func main() {
	deps, err := factory.NewGlobalDependencies()
	if err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}

	logging.SetupLogger(logging.LoggerParams{
		Level:       deps.Config.Logger.Level,
		Environment: deps.Config.Environment,
		ServiceName: deps.Config.ServiceName,
		Version:     deps.Config.Version,
	})

	taskFactory := factory.NewTaskFactory()
	factory.RegisterDepdenencies(taskFactory, deps)
	factory.RegisterTasks(taskFactory)

	repo := mock.NewMockRepo()

	jobService := service.NewJobService(&service.JobServiceParams{
		Config:      deps.Config.Worker,
		TaskFactory: taskFactory,
		Repository:  repo,
	})

	ctx := context.Background()
	jobService.StartWorkers(ctx)

	httpServer := server.NewServer(&server.ServerParams{
		ServerConfig:    deps.Config.Server,
		RateLimitConfig: deps.Config.RateLimit,
		JobService:      jobService,
	})

	// Start server in a goroutine
	go func() {
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server failed to start", slog.Any("error", err))
		}
	}()

	slog.Info("server is listening and ready to handle requests")

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("server shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(ctx); err != nil {
		slog.Error("server forced to shutdown", slog.Any("error", err))
	}

	slog.Info("server stopped")
}
