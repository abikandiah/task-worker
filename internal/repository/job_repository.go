package repository

import (
	"context"

	"github.com/abikandiah/task-worker/internal/domain"
)

type JobRepository interface {
	GetJob(ctx context.Context, jobID string) (*domain.Job, error)
	SaveJob(ctx context.Context, job domain.Job) (*domain.Job, error)
}
