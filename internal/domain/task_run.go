package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type TaskRun struct {
	Identity
	JobID     uuid.UUID      `json:"jobId"`
	TaskName  string         `json:"taskName"`
	Params    any            `json:"params"`
	Parallel  bool           `json:"parallel"`
	State     ExecutionState `json:"state"`
	Progress  float32        `json:"progress"`
	Result    any            `json:"result"`
	StartDate time.Time      `json:"startDate,omitempty"`
	EndDate   time.Time      `json:"endDate,omitempty"`
}

type TaskRunRepository interface {
	GetTaskRuns(ctx context.Context, jobID uuid.UUID) ([]TaskRun, error)
	SaveTaskRuns(ctx context.Context, taskRuns []TaskRun) ([]TaskRun, error)

	GetTaskRun(ctx context.Context, taskRunID uuid.UUID) (*TaskRun, error)
	SaveTaskRun(ctx context.Context, taskRun TaskRun) (*TaskRun, error)
}
