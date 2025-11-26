package app

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/abikandiah/task-worker/config"
	"github.com/abikandiah/task-worker/internal/domain"
	"github.com/abikandiah/task-worker/internal/factory"
	"github.com/abikandiah/task-worker/internal/platform/logging"
	"github.com/abikandiah/task-worker/internal/platform/server"
	"github.com/abikandiah/task-worker/internal/service"
	"github.com/abikandiah/task-worker/internal/task"
)

type AppDependencies struct {
	Config      *config.Config
	Repository  domain.ServiceRepository
	TaskFactory domain.TaskFactory
	Logger      *slog.Logger
}

// internal/app/app.go
type Application struct {
	*AppDependencies
	JobService *service.JobService
}

// NewApplication constructs the entire application stack.
func NewApplication(deps *AppDependencies) *Application {
	app := &Application{
		AppDependencies: deps,
	}

	app.Logger = logging.SetupLogger(deps.Config.Logger)

	taskFactory := factory.NewTaskFactory()
	app.TaskFactory = taskFactory

	registerDepdenencies(taskFactory, app.AppDependencies)
	registerTasks(taskFactory)

	app.JobService = service.NewJobService(&service.JobServiceParams{
		Config:      app.Config.Worker,
		TaskFactory: app.TaskFactory,
		Repository:  app.Repository,
	})

	return app
}

func (app *Application) Run() {
	app.startService()
	app.startHttpServer()
}

func (app *Application) startService() {
	app.JobService.StartWorkers(context.Background())
}

// Blocking, starts HTTP server
func (app *Application) startHttpServer() {
	httpServer := server.NewServer(&server.ServerParams{
		ServerConfig: app.Config.Server,
		JobService:   app.JobService,
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

func (app *Application) Close() {
	app.JobService.Close(context.Background())
}

func registerDepdenencies(f *factory.TaskFactory, deps *AppDependencies) {
	factory.RegisterDependency(f, deps.Config)
	factory.RegisterDependency(f, deps.Repository)
}

func registerTasks(f *factory.TaskFactory) {
	factory.Register(f, "chat", task.ChatConstructor)
	factory.Register(f, "send_email", task.SendEmailConstructor)
	factory.Register(f, "duration", task.DurationConstructor)
}
