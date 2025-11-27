package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Job struct {
	Identity
	Status
	ConfigID      uuid.UUID  `json:"configId,omitempty"`
	ConfigVersion uuid.UUID  `json:"configVersion,omitempty"`
	SubmitDate    time.Time  `json:"submitDate"`
	StartDate     *time.Time `json:"startDate,omitempty"`
	EndDate       *time.Time `json:"endDate,omitempty"`
}

type JobSubmission struct {
	IdentitySubmission
	ConfigID      uuid.UUID `json:"configId,omitempty"`
	ConfigVersion uuid.UUID `json:"configVersion,omitempty"`
	TaskRuns      []TaskRun `json:"taskRuns"`
}

// GetID implements the required method for cursor pagination.
func (job Job) GetID() uuid.UUID {
	return job.ID
}

type JobConfig struct {
	IdentityVersion
	IsDefault        bool `json:"isDefault"`
	JobConfigDetails `json:"details"`
}

type JobConfigDetails struct {
	JobTimeout          int  `json:"jobTimeout"`
	TaskTimeout         int  `json:"taskTimeout"`
	EnableParallelTasks bool `json:"enableParallelTasks"`
	MaxParallelTasks    int  `json:"maxParallelTasks"`
}

// GetID implements the required method for cursor pagination.
func (config JobConfig) GetID() uuid.UUID {
	return config.ID
}

func NewDefaultJobConfig() *JobConfig {
	submission := IdentitySubmission{
		Name: "Default JobConfig",
	}
	identity := Identity{
		IdentitySubmission: submission,
	}
	identityVersion := IdentityVersion{
		Identity: identity,
		Version:  uuid.New(),
	}
	return &JobConfig{
		IdentityVersion: identityVersion,
		IsDefault:       true,
		JobConfigDetails: JobConfigDetails{
			JobTimeout:          600,
			TaskTimeout:         120,
			EnableParallelTasks: true,
			MaxParallelTasks:    2,
		},
	}
}

type JobRepository interface {
	SaveJob(ctx context.Context, job Job) (*Job, error)
	GetJob(ctx context.Context, jobID uuid.UUID) (*Job, error)
	GetAllJobs(ctx context.Context, cursor *CursorInput) (*CursorOutput[Job], error)

	GetDefaultJobConfig(ctx context.Context) (*JobConfig, error)
	GetOrCreateDefaultJobConfig(ctx context.Context) (*JobConfig, error)

	SaveJobConfig(ctx context.Context, config JobConfig) (*JobConfig, error)
	GetJobConfig(ctx context.Context, configID uuid.UUID) (*JobConfig, error)
	GetAllJobConfigs(ctx context.Context, cursor *CursorInput) (*CursorOutput[JobConfig], error)
}
