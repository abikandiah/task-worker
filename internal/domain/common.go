package domain

import (
	"encoding/json"
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

type Status struct {
	State    ExecutionState `json:"state"`
	Progress float32        `json:"progress"`
}

// State
type ExecutionState int

const (
	StatePending ExecutionState = iota
	StateRunning
	StateFinished
	StateStopped
	StatePaused
	StateWarning
	StateError
	StateRejected
)

// --- Lookup Maps ---

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

// stringToExecutionState is the inverse map for fast string-to-int lookups
var stringToExecutionState = generateStringToExecutionState()

func generateStringToExecutionState() map[string]ExecutionState {
	res := make(map[string]ExecutionState, len(executionStateStrings))
	for state, val := range executionStateStrings {
		res[val] = state
	}
	return res
}

// --- Stringer Method ---

// String() is called by MarshalJSON and fmt.Print* functions.
func (s ExecutionState) String() string {
	if str, ok := executionStateStrings[s]; ok {
		return str
	}
	return fmt.Sprintf("ExecutionState(%d)", s)
}

// --- JSON Marshalling/Unmarshalling ---

// MarshalJSON converts the ExecutionState int to a quoted string.
func (s ExecutionState) MarshalJSON() ([]byte, error) {
	return []byte(`"` + s.String() + `"`), nil
}

// UnmarshalJSON converts a JSON string (e.g., "RUNNING") back to the ExecutionState int type.
func (s *ExecutionState) UnmarshalJSON(b []byte) error {
	var strVal string
	if err := json.Unmarshal(b, &strVal); err != nil {
		return err // Return if unquoting fails (e.g., input was a number, not a string)
	}

	state, ok := stringToExecutionState[strVal]
	if !ok {
		return fmt.Errorf("invalid ExecutionState value: %s", string(b))
	}

	// Assign the found integer constant back to the pointer *s
	*s = state
	return nil
}
