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
	"github.com/abikandiah/task-worker/internal/platform/server"
	"github.com/abikandiah/task-worker/internal/service"
)

// Entry point for HTTP server
func main() {
	deps, err := factory.NewGlobalDependencies()
	if err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}

	taskFactory := factory.NewTaskFactory(deps)
	repo := mock.NewMockRepo()

	jobService := service.NewJobService(&service.JobServiceParams{
		Config:      deps.Config.Worker,
		Logger:      deps.Logger,
		TaskFactory: taskFactory,
		Repository:  repo,
	})

	ctx := context.Background()
	jobService.StartWorkers(ctx)

	httpServer := server.NewServer(&server.ServerParams{
		ServerConfig:    deps.Config.Server,
		RateLimitConfig: deps.Config.RateLimit,
		Logger:          deps.Logger,
		JobService:      jobService,
	})

	// Start server in a goroutine
	go func() {
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			deps.Logger.Error("server failed to start", slog.Any("error", err))
		}
	}()

	deps.Logger.Info("server is listening and ready to handle requests")

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	deps.Logger.Info("server shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(ctx); err != nil {
		deps.Logger.Error("server forced to shutdown", slog.Any("error", err))
	}

	deps.Logger.Info("server stopped")
}
