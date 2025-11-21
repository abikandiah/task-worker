package worker

import (
	"errors"
	"time"

	"github.com/abikandiah/task-worker/internal/common"
)

type Task struct {
	common.IdentityVersion
}

func (t Task) Run() (any, error) {
	return nil, errors.New("not implemented")
}

type TaskRun struct {
	common.Identity
	JobID       string
	TaskID      string
	TaskVersion string
	Options     map[string]any
	Status      string
	Progress    float32
	Result      any
	StartDate   time.Time
	EndDate     time.Time
}

type Job struct {
	common.Identity
	Status        string
	Progress      float32
	SubmittedDate time.Time
	StartDate     time.Time
	EndDate       time.Time
}

type JobRequest struct {
	common.Identity
	TaskIDs []string
}
