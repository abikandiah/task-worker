package domain

import (
	"fmt"

	"github.com/google/uuid"
)

type LogKey string

type Identity struct {
	IdentitySubmission
	ID uuid.UUID
}

type IdentitySubmission struct {
	Name        string
	Description string
}

type IdentityVersion struct {
	Identity
	Version string
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
