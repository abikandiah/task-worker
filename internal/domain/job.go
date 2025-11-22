package domain

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

type Job struct {
	Identity
	ConfigID   uuid.UUID
	State      JobState
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
type JobState int

const (
	_ JobState = iota

	StatePending
	StateRunning
	StateFinished
	StateStopped
	StatePaused
	StateWarning
	StateError
	StateRejected
)

var jobStateStrings = map[JobState]string{
	StatePending:  "PENDING",
	StateRunning:  "RUNNING",
	StateFinished: "FINISHED",
	StateStopped:  "STOPPED",
	StatePaused:   "PAUSED",
	StateWarning:  "WARNING",
	StateError:    "ERROR",
	StateRejected: "REJECTED",
}

func (s JobState) String() string {
	if str, ok := jobStateStrings[s]; ok {
		return str
	}
	return fmt.Sprintf("JobState(%d)", s)
}
