package task

import (
	"context"
	"log/slog"
	"time"
)

type ChatParams struct {
	Message string
}

type ChatTaskDependencies struct{}

type ChatTask struct {
	*ChatParams
	*ChatTaskDependencies
}

func ChatConstructor(params *ChatParams, deps *ChatTaskDependencies) (Task, error) {
	if params == nil {
		params = &ChatParams{Message: "default-chat"}
	}
	return &ChatTask{
		ChatParams:           params,
		ChatTaskDependencies: deps,
	}, nil
}

func (task *ChatTask) Execute(ctx context.Context) (any, error) {
	slog.InfoContext(ctx, task.Message)

	<-time.After(20 * time.Second)
	return nil, nil
}
