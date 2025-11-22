package repository

import (
	"context"

	"github.com/abikandiah/task-worker/internal/domain"
	"github.com/google/uuid"
)

type TaskRunRepository interface {
	GetTaskRuns(ctx context.Context, jobID uuid.UUID) ([]domain.TaskRun, error)
	SaveTaskRuns(ctx context.Context, taskRuns []domain.TaskRun) ([]domain.TaskRun, error)

	GetTaskRun(ctx context.Context, taskRunID uuid.UUID) (*domain.TaskRun, error)
	SaveTaskRun(ctx context.Context, taskRun domain.TaskRun) (*domain.TaskRun, error)
}
