package domain

import (
	"context"
	"errors"
	"time"
)

type Task struct {
	IdentityVersion
}

func (t Task) Run() (any, error) {
	return nil, errors.New("not implemented")
}

type TaskRun struct {
	Identity
	JobID       string
	TaskID      string
	TaskVersion string
	DependentOn []string
	Options     map[string]any
	Status      string
	Progress    float32
	Result      any
	StartDate   time.Time
	EndDate     time.Time
}

type TaskRepository interface {
	GetTaskRuns(ctx context.Context, jobID string) ([]TaskRun, error)
	GetTaskRun(ctx context.Context, taskRunID string) (*TaskRun, error)
	SaveTaskRun(ctx context.Context, taskRun TaskRun) (*TaskRun, error)

	GetTask(ctx context.Context, taskID string, taskVersion string) (*Task, error)
	SaveTask(ctx context.Context, task Task) (*Task, error)
}
