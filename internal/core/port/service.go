package port

import (
	"context"

	"github.com/abikandiah/task-worker/internal/core/domain"
)

type Task interface {
	Execute(ctx context.Context) error
}

type ExecutorService interface {
	ExecuteJob(ctx context.Context, jobID string) error
	ExecuteTaskRun(ctx context.Context, taskRun domain.TaskRun) error
}
