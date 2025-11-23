package executor

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/abikandiah/task-worker/internal/domain"
	"github.com/abikandiah/task-worker/internal/domain/task"
)

type EmailDependencies struct {
	Logger *slog.Logger
}

type EmailSendTask struct {
	Params *task.EmailSendParams
	Deps   *EmailDependencies
}

func EmailSendConstructor(params any, deps *domain.GlobalDependencies) (task.Task, error) {
	taskParams, ok := params.(task.EmailSendParams)
	if !ok {
		return nil, fmt.Errorf("invalid params passed to EmailSend task factory: %T", params)
	}

	taskDeps := &EmailDependencies{
		Logger: deps.Logger,
	}

	return &EmailSendTask{
		Params: &taskParams,
		Deps:   taskDeps,
	}, nil
}

func (task *EmailSendTask) Execute(ctx context.Context) (any, error) {
	<-time.After(10 * time.Second)
	return nil, nil
}
