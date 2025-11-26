package sqlite3

import (
	"context"

	"github.com/abikandiah/task-worker/internal/domain"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type SQLiteServiceRepository struct {
	DB *sqlx.DB
}

func NewSQLiteServiceRepository(db *sqlx.DB) *SQLiteServiceRepository {
	return &SQLiteServiceRepository{
		DB: db,
	}
}

func (repo *SQLiteServiceRepository) GetAllJobs(ctx context.Context, input *domain.CursorInput) (*domain.CursorOutput[domain.Job], error) {
	return nil, nil
}

func (repo *SQLiteServiceRepository) GetJob(ctx context.Context, jobID uuid.UUID) (*domain.Job, error) {
	return nil, nil
}

func (repo *SQLiteServiceRepository) SaveJob(ctx context.Context, job domain.Job) (*domain.Job, error) {
	return nil, nil
}

func (repo *SQLiteServiceRepository) GetAllJobConfigs(ctx context.Context, input *domain.CursorInput) (*domain.CursorOutput[domain.JobConfig], error) {
	return nil, nil
}

func (repo *SQLiteServiceRepository) GetJobConfig(ctx context.Context, configID uuid.UUID) (*domain.JobConfig, error) {
	return nil, nil
}

func (repo *SQLiteServiceRepository) SaveJobConfig(ctx context.Context, config domain.JobConfig) (*domain.JobConfig, error) {
	return nil, nil
}

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
