package domain

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type Task interface {
	Execute(ctx context.Context) (any, error)
}

type TaskRun struct {
	Identity
	JobID          uuid.UUID      `json:"jobId"`
	TaskName       string         `json:"taskName"`
	State          ExecutionState `json:"state"`
	StartDate      *time.Time     `json:"startDate,omitempty"`
	EndDate        *time.Time     `json:"endDate,omitempty"`
	TaskRunDetails `json:"details"`
}

type TaskRunDetails struct {
	Parallel bool            `json:"parallel"`
	Params   json.RawMessage `json:"params"`
	Result   any             `json:"result"`
	Progress float32         `json:"progress"`
}
