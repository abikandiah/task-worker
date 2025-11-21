package domain

import (
	"context"
)

type ExecutorConfig struct {
	IdentityVersion
	EnableParallelTasks bool
	MaxParallelTasks    int
}

func NewExecutorConfig() *ExecutorConfig {
	return &ExecutorConfig{
		MaxParallelTasks: 4,
	}
}

type ExecutorConfigRepository interface {
	GetExecutorConfig(ctx context.Context, configID string) (*ExecutorConfig, error)
	SaveExecutorConfig(ctx context.Context, config ExecutorConfig) (*ExecutorConfig, error)
}

type ExecutorRepository interface {
	ExecutorConfigRepository
	JobRepository
	TaskRepository
}

type ExecutorService interface {
	ExecuteJob(ctx context.Context, jobID string) error
	ExecuteTaskRun(ctx context.Context, taskRun TaskRun) error
}
