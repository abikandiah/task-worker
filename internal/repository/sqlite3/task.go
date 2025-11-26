package sqlite3

import (
	"context"
	"encoding/json"
	"time"

	"github.com/abikandiah/task-worker/internal/domain"
	"github.com/google/uuid"
)

type TaskRunDB struct {
	ID          uuid.UUID             `db:"id"`
	JobID       uuid.UUID             `db:"job_id"`
	Name        string                `db:"name"`
	Description string                `db:"description"`
	TaskName    string                `db:"task_name"`
	State       domain.ExecutionState `db:"state"`
	StartDate   time.Time             `db:"start_date"`
	EndDate     time.Time             `db:"end_date"`
	DetailsJSON TaskRunDetails        `db:"details"`
}

type TaskRunDetails struct {
	Parallel bool            `json:"parallel"`
	Params   json.RawMessage `json:"params"`
	Result   any             `json:"result"`
	Progress float32         `json:"progress"`
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
