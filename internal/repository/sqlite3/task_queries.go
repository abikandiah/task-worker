package sqlite3

import (
	"context"

	"github.com/abikandiah/task-worker/internal/domain"
	"github.com/google/uuid"
)

func (repo *SQLiteServiceRepository) GetTaskRun(ctx context.Context, taskID uuid.UUID) (*domain.TaskRun, error) {
	return nil, nil
}

func (repo *SQLiteServiceRepository) SaveTaskRun(ctx context.Context, taskRun domain.TaskRun) (*domain.TaskRun, error) {
	return nil, nil
}

func (repo *SQLiteServiceRepository) SaveTaskRuns(ctx context.Context, taskRuns []domain.TaskRun) ([]domain.TaskRun, error) {
	return nil, nil
}

func (repo *SQLiteServiceRepository) GetTaskRuns(ctx context.Context, jobID uuid.UUID) ([]domain.TaskRun, error) {
	return nil, nil
}
