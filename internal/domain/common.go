package domain

import (
	"fmt"

	"github.com/google/uuid"
)

type LogKey string

type Identity struct {
	IdentitySubmission
	ID uuid.UUID `json:"id"`
}

type IdentitySubmission struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type IdentityVersion struct {
	Identity
	Version string `json:"version"`
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
