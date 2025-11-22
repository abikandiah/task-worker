package domain

import (
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
