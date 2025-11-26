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

type DurationDependencies struct{}

type DurationTask struct {
	*DurationParams
	*DurationDependencies
}

func DurationConstructor(params *DurationParams, deps *DurationDependencies) (domain.Task, error) {
	if params == nil {
		params = &DurationParams{Length: 10}
	}
	if params.Length < 0 {
		return nil, fmt.Errorf("duration must be greater than 0")
	}

	return &DurationTask{
		DurationParams:       params,
		DurationDependencies: deps,
	}, nil
}

func (task *DurationTask) Execute(ctx context.Context) (any, error) {
	slog.InfoContext(ctx, "starting duration task")
	slog.InfoContext(ctx, fmt.Sprintf("waiting for %d", task.Length))

	<-time.After(time.Duration(task.Length) * time.Second)
	return nil, nil
}
