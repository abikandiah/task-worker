package port

import (
	"context"

	"github.com/abikandiah/task-worker/internal/core/domain"
)

type JobRepository interface {
	GetJob(ctx context.Context, jobID string) (*domain.Job, error)
	SaveJob(ctx context.Context, job domain.Job) (*domain.Job, error)
}

type TaskRunRepository interface {
	GetTaskRuns(ctx context.Context, jobID string) ([]domain.TaskRun, error)
	GetTaskRun(ctx context.Context, taskRunID string) (*domain.TaskRun, error)
	SaveTaskRun(ctx context.Context, taskRun domain.TaskRun) (*domain.TaskRun, error)
}

type ExecutorConfigRepository interface {
	GetExecutorConfig(ctx context.Context, configID string) (*domain.ExecutorConfig, error)
	SaveExecutorConfig(ctx context.Context, config domain.ExecutorConfig) (*domain.ExecutorConfig, error)
}

type ExecutorRepository interface {
	ExecutorConfigRepository
	JobRepository
	TaskRunRepository
}
