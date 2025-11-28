package repository

import (
	"context"

	"github.com/abikandiah/task-worker/internal/domain"
	"github.com/google/uuid"
)

type ServiceRepository interface {
	JobRepository
	TaskRunRepository
	Close() error
}

type JobRepository interface {
	SaveJob(ctx context.Context, job domain.Job) (*domain.Job, error)
	GetJob(ctx context.Context, jobID uuid.UUID) (*domain.Job, error)
	GetAllJobs(ctx context.Context, cursor *domain.CursorInput) (*domain.CursorOutput[domain.Job], error)

	GetDefaultJobConfig(ctx context.Context) (*domain.JobConfig, error)
	GetOrCreateDefaultJobConfig(ctx context.Context) (*domain.JobConfig, error)

	SaveJobConfig(ctx context.Context, config domain.JobConfig) (*domain.JobConfig, error)
	GetJobConfig(ctx context.Context, configID uuid.UUID) (*domain.JobConfig, error)
	GetAllJobConfigs(ctx context.Context, cursor *domain.CursorInput) (*domain.CursorOutput[domain.JobConfig], error)
}

type TaskRunRepository interface {
	SaveTaskRun(ctx context.Context, taskRun domain.TaskRun) (*domain.TaskRun, error)
	SaveTaskRuns(ctx context.Context, taskRuns []domain.TaskRun) ([]domain.TaskRun, error)

	GetTaskRun(ctx context.Context, taskRunID uuid.UUID) (*domain.TaskRun, error)
	GetTaskRuns(ctx context.Context, jobID uuid.UUID) ([]domain.TaskRun, error)
	GetAllTaskRuns(ctx context.Context, cursor *domain.CursorInput) (*domain.CursorOutput[domain.TaskRun], error)
}
