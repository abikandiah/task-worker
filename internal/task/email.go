package task

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/abikandiah/task-worker/internal/domain"
)

type EmailParams struct {
	Recipient string
	Subject   string
	Message   string
}

type EmailDependencies struct {
	Logger *slog.Logger
}

type EmailSendTask struct {
	params *EmailParams
	deps   *EmailDependencies
}

func EmailSendConstructor(params *EmailParams, deps *domain.GlobalDependencies) (Task, error) {
	if params.Recipient == "" {
		return nil, fmt.Errorf("recipient is required")
	}

	taskDeps := &EmailDependencies{
		Logger: deps.Logger,
	}

	return &EmailSendTask{
		params: params,
		deps:   taskDeps,
	}, nil
}

func (task *EmailSendTask) Execute(ctx context.Context) (any, error) {
	<-time.After(10 * time.Second)
	return nil, nil
}
