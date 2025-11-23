package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type TaskRun struct {
	Identity
	JobID     uuid.UUID
	TaskName  string
	Params    any
	Parallel  bool
	State     ExecutionState
	Progress  float32
	Result    any
	StartDate time.Time
	EndDate   time.Time
}

type TaskRunRepository interface {
	GetTaskRuns(ctx context.Context, jobID uuid.UUID) ([]TaskRun, error)
	SaveTaskRuns(ctx context.Context, taskRuns []TaskRun) ([]TaskRun, error)

	GetTaskRun(ctx context.Context, taskRunID uuid.UUID) (*TaskRun, error)
	SaveTaskRun(ctx context.Context, taskRun TaskRun) (*TaskRun, error)
}
