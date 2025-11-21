package domain

import (
	"context"
	"time"
)

type Job struct {
	Identity
	Status        string
	Progress      float32
	SubmittedDate time.Time
	StartDate     time.Time
	EndDate       time.Time
}

type JobRepository interface {
	GetJob(ctx context.Context, jobID string) (*Job, error)
	SaveJob(ctx context.Context, job Job) (*Job, error)
}
