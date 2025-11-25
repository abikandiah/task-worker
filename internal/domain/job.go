package domain

import (
	"context"
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

type JobRepository interface {
	GetAllJobs(ctx context.Context, input *CursorInput) (*CursorOutput[Job], error)

	GetJob(ctx context.Context, jobID uuid.UUID) (*Job, error)
	SaveJob(ctx context.Context, job Job) (*Job, error)

	GetJobConfig(ctx context.Context, configID uuid.UUID) (*JobConfig, error)
	SaveJobConfig(ctx context.Context, config JobConfig) (*JobConfig, error)
}
