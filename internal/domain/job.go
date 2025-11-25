package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Job struct {
	Identity
	ConfigID   uuid.UUID      `json:"configId"`
	State      ExecutionState `json:"state"`
	Progress   float32        `json:"progress"`
	SubmitDate time.Time      `json:"submitDate"`
	StartDate  time.Time      `json:"startDate,omitempty"`
	EndDate    time.Time      `json:"endDate,omitempty"`
}

type JobSubmission struct {
	IdentitySubmission
	TaskRuns []TaskRun `json:"taskRuns"`
}

type JobConfig struct {
	IdentityVersion
	JobTimeout          int  `json:"jobTimeout"`
	TaskTimeout         int  `json:"taskTimeout"`
	EnableParallelTasks bool `json:"enableParallelTasks"`
	MaxParallelTasks    int  `json:"maxParallelTasks"`
}

type JobRepository interface {
	GetAllJobs(ctx context.Context, input *CursorInput) (*CursorOutput[Job], error)

	GetJob(ctx context.Context, jobID uuid.UUID) (*Job, error)
	SaveJob(ctx context.Context, job Job) (*Job, error)

	GetJobConfig(ctx context.Context, configID uuid.UUID) (*JobConfig, error)
	SaveJobConfig(ctx context.Context, config JobConfig) (*JobConfig, error)
}
