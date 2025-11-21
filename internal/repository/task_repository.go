package repository

import (
	"context"

	"github.com/abikandiah/task-worker/internal/domain"
)

type TaskRunRepository interface {
	GetTaskRuns(ctx context.Context, jobID string) ([]domain.TaskRun, error)
	GetTaskRun(ctx context.Context, taskRunID string) (*domain.TaskRun, error)
	SaveTaskRun(ctx context.Context, taskRun domain.TaskRun) (*domain.TaskRun, error)
}
