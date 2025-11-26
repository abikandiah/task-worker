package domain

import (
	"fmt"

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

var stringToExecutionState = generateStringToExecutionState()

func generateStringToExecutionState() map[string]ExecutionState {
	res := make(map[string]ExecutionState, len(executionStateStrings))
	for state, val := range executionStateStrings {
		res[val] = state
	}
	return res
}

func (s ExecutionState) String() string {
	if str, ok := executionStateStrings[s]; ok {
		return str
	}
	return fmt.Sprintf("ExecutionState(%d)", s)
}

func (s ExecutionState) MarshalJSON() ([]byte, error) {
	return []byte(`"` + s.String() + `"`), nil
}

func (s *ExecutionState) UnmarshalJSON(b []byte) error {
	// Remove the surrounding quotes from the JSON string (e.g., "ACTIVE" -> ACTIVE)
	strVal := string(b)
	if len(strVal) > 0 && strVal[0] == '"' && strVal[len(strVal)-1] == '"' {
		strVal = strVal[1 : len(strVal)-1]
	}

	state, ok := stringToExecutionState[strVal]
	if !ok {
		return fmt.Errorf("invalid Status value: %s", strVal)
	}
	*s = state
	return nil
}
