package domain

import (
	"context"

	"github.com/abikandiah/task-worker/internal/domain/task"
)

type JobScheduler interface {
	SubmitJob(ctx context.Context, submission *JobSubmission) (*Job, error)
	GetJob(ctx context.Context, jobID string) (*Job, error)
}

type JobWorker interface {
	Run(ctx context.Context, job *Job, config *JobConfig) error
}

type TaskWorker interface {
	Run(ctx context.Context, task *task.Task) error
}
