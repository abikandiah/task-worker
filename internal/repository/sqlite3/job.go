package sqlite3

import (
	"time"

	"github.com/abikandiah/task-worker/internal/domain"
	"github.com/google/uuid"
)

type Job struct {
	ID          uuid.UUID `db:"id"`
	Name        string    `db:"name"`
	Description string
	JobStatus
	ConfigID   uuid.UUID `db:"configId,omitempty"`
	SubmitDate time.Time `db:"submitDate"`
	StartDate  time.Time `db:"startDate,omitempty"`
	EndDate    time.Time `db:"endDate,omitempty"`
}

type JobStatus struct {
	State    domain.ExecutionState `db:"state"`
	Progress float32               `db:"progress"`
}

type JobConfigDB struct {
	// Fixed Columns
	ID      uuid.UUID `db:"id"`
	Name    string    `db:"name"`
	Version string    `db:"version"`

	// JSON Payload Column
	DetailsJSON JobConfigDetails `db:"details"`
}

type JobConfigDetails struct {
	JobTimeout          int  `json:"jobTimeout"`
	TaskTimeout         int  `json:"taskTimeout"`
	EnableParallelTasks bool `json:"enableParallelTasks"`
	MaxParallelTasks    int  `json:"maxParallelTasks"`
}
