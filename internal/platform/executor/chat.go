package executor

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/abikandiah/task-worker/internal/domain"
	"github.com/abikandiah/task-worker/internal/domain/task"
)

type ChatTaskDependencies struct {
	Logger *slog.Logger
}

type ChatTask struct {
	Params *task.ChatParams
	Deps   *ChatTaskDependencies
}

func ChatConstructor(params any, deps *domain.GlobalDependencies) (task.Task, error) {
	taskParams, ok := params.(task.ChatParams)
	if !ok {
		return nil, fmt.Errorf("invalid params passed to ChatTask factory: %T", params)
	}

	taskDeps := &ChatTaskDependencies{
		Logger: deps.Logger,
	}

	return &ChatTask{
		Params: &taskParams,
		Deps:   taskDeps,
	}, nil
}

func (task *ChatTask) Execute(ctx context.Context) (any, error) {
	<-time.After(20 * time.Second)
	return nil, nil
}
