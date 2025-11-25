package task

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/abikandiah/task-worker/internal/domain"
)

type DurationParams struct {
	Length int
}

type DurationDependencies struct {
	Logger *slog.Logger
}

type DurationTask struct {
	*DurationParams
	deps *DurationDependencies
}

func DurationConstructor(params *DurationParams, deps *domain.GlobalDependencies) (Task, error) {
	if params.Length < 0 {
		return nil, fmt.Errorf("duration must be greater than 0")
	}

	taskDeps := &DurationDependencies{
		Logger: deps.Logger,
	}

	return &DurationTask{
		DurationParams: params,
		deps:           taskDeps,
	}, nil
}

func (task *DurationTask) Execute(ctx context.Context) (any, error) {
	task.deps.Logger.InfoContext(ctx, "starting duration task")
	task.deps.Logger.InfoContext(ctx, fmt.Sprintf("waiting for %d", task.Length))

	<-time.After(time.Duration(task.Length) * time.Second)
	return nil, nil
}
