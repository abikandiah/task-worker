package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Job struct {
	Identity
	Status
	ConfigID   uuid.UUID `json:"configId,omitempty"`
	SubmitDate time.Time `json:"submitDate"`
	StartDate  time.Time `json:"startDate,omitempty"`
	EndDate    time.Time `json:"endDate,omitempty"`
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

func NewDefaultJobConfig() *JobConfig {
	submission := IdentitySubmission{
		Name: "Default JobConfig",
	}
	identity := Identity{
		IdentitySubmission: submission,
		ID:                 uuid.New(),
	}
	identityVersion := IdentityVersion{
		Identity: identity,
		Version:  "1.0",
	}
	return &JobConfig{
		IdentityVersion:     identityVersion,
		JobTimeout:          600,
		TaskTimeout:         120,
		EnableParallelTasks: true,
		MaxParallelTasks:    2,
	}
}

type JobRepository interface {
	GetAllJobs(ctx context.Context, input *CursorInput) (*CursorOutput[Job], error)
	GetAllJobConfigs(ctx context.Context, input *CursorInput) (*CursorOutput[JobConfig], error)

	GetJob(ctx context.Context, jobID uuid.UUID) (*Job, error)
	SaveJob(ctx context.Context, job Job) (*Job, error)

	GetJobConfig(ctx context.Context, configID uuid.UUID) (*JobConfig, error)
	SaveJobConfig(ctx context.Context, config JobConfig) (*JobConfig, error)
}
