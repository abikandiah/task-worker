package repository

import (
	"context"

	"github.com/abikandiah/task-worker/internal/domain"
)

type ExecutorConfigRepository interface {
	GetExecutorConfig(ctx context.Context, configID string) (*domain.ExecutorConfig, error)
	SaveExecutorConfig(ctx context.Context, config domain.ExecutorConfig) (*domain.ExecutorConfig, error)
}

type ExecutorRepository interface {
	ExecutorConfigRepository
	JobRepository
	TaskRunRepository
}
