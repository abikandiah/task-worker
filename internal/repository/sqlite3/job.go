package sqlite3

import (
	"context"
	"time"

	"github.com/abikandiah/task-worker/internal/domain"
	"github.com/google/uuid"
)

type JobDB struct {
	ID          uuid.UUID             `db:"id"`
	Name        string                `db:"name"`
	Description string                `db:"description"`
	ConfigID    uuid.UUID             `db:"config_id"`
	State       domain.ExecutionState `db:"state"`
	Progress    float32               `db:"progress"`
	SubmitDate  time.Time             `db:"submit_date"`
	StartDate   time.Time             `db:"start_date"`
	EndDate     time.Time             `db:"end_date"`
}

type JobConfigDB struct {
	ID          uuid.UUID        `db:"id"`
	Name        string           `db:"name"`
	Version     string           `db:"version"`
	DetailsJSON JobConfigDetails `db:"details"`
}

type JobConfigDetails struct {
	JobTimeout          int  `json:"job_timeout"`
	TaskTimeout         int  `json:"task_timeout"`
	EnableParallelTasks bool `json:"enable_parallel_tasks"`
	MaxParallelTasks    int  `json:"max_parallel_tasks"`
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
