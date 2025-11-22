package repository

import (
	"context"

	"github.com/abikandiah/task-worker/internal/domain"
	"github.com/google/uuid"
)

type JobRepository interface {
	GetJob(ctx context.Context, jobID uuid.UUID) (*domain.Job, error)
	SaveJob(ctx context.Context, job domain.Job) (*domain.Job, error)

	GetJobConfig(ctx context.Context, configID uuid.UUID) (*domain.JobConfig, error)
	SaveJobConfig(ctx context.Context, config domain.JobConfig) (*domain.JobConfig, error)
}
