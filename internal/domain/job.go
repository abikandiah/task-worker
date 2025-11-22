package domain

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

type Job struct {
	Identity
	ConfigID   uuid.UUID
	State      ExecutionState
	Progress   float32
	SubmitDate time.Time
	StartDate  time.Time
	EndDate    time.Time
}

type JobSubmission struct {
	IdentitySubmission
	TaskRuns []TaskRun
}

type JobConfig struct {
	IdentityVersion
	JobTimeout          int
	TaskTimeout         int
	EnableParallelTasks bool
	MaxParallelTasks    int
}

// State
type ExecutionState int

const (
	_ ExecutionState = iota

	StatePending
	StateRunning
	StateFinished
	StateStopped
	StatePaused
	StateWarning
	StateError
	StateRejected
)

var executionStateStrings = map[ExecutionState]string{
	StatePending:  "PENDING",
	StateRunning:  "RUNNING",
	StateFinished: "FINISHED",
	StateStopped:  "STOPPED",
	StatePaused:   "PAUSED",
	StateWarning:  "WARNING",
	StateError:    "ERROR",
	StateRejected: "REJECTED",
}

func (s ExecutionState) String() string {
	if str, ok := executionStateStrings[s]; ok {
		return str
	}
	return fmt.Sprintf("ExecutionState(%d)", s)
}
