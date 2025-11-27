package domain

import (
	"github.com/google/uuid"
)

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

type Status struct {
	State    ExecutionState `json:"state"`
	Progress float32        `json:"progress"`
}

// State
type ExecutionState string

const (
	StatePending  ExecutionState = "PENDING"
	StateRunning  ExecutionState = "RUNNING"
	StateFinished ExecutionState = "FINISHED"
	StateStopped  ExecutionState = "STOPPED"
	StatePaused   ExecutionState = "PAUSED"
	StateWarning  ExecutionState = "WARNING"
	StateError    ExecutionState = "ERROR"
	StateRejected ExecutionState = "REJECTED"
)

func GetStateName(state ExecutionState) string {
	return string(state)
}
