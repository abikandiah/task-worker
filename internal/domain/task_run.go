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
	DependOn  []int
	State    string
	Progress  float32
	Result    any
	StartDate time.Time
	EndDate   time.Time
}
