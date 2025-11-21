package domain

import (
	"context"
)

type ExecutorService interface {
	ExecuteJob(ctx context.Context, jobID string) error
	ExecuteTaskRun(ctx context.Context, taskRun TaskRun) error
}
