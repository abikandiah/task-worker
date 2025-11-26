package factory

import (
	"github.com/abikandiah/task-worker/config"
	"github.com/abikandiah/task-worker/internal/domain"
	"github.com/abikandiah/task-worker/internal/task"
)

func NewGlobalDependencies() (*domain.GlobalDependencies, error) {
	config := config.MustLoad()

	return &domain.GlobalDependencies{
		Config: config,
	}, nil
}

func RegisterDepdenencies(factory *TaskFactory, deps *domain.GlobalDependencies) {
	RegisterDependency(factory, deps.Config)
	RegisterDependency(factory, deps.Repository)
}

func RegisterTasks(factory *TaskFactory) {
	Register(factory, "chat", task.ChatConstructor)
	Register(factory, "send_email", task.SendEmailConstructor)
	Register(factory, "duration", task.DurationConstructor)
}
