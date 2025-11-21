package executor

import (
	"context"
	"fmt"
	"log/slog"
)

type EmailParams struct {
	Recipient string
}

type EmailDependencies struct {
	Logger *slog.Logger
}

type EmailSendTask struct {
	Params *EmailParams
	Deps   *EmailDependencies
}

func (task *EmailSendTask) Execute(ctx context.Context) error {
	return nil
}

func NewEmailSendTask(params any, deps *EmailDependencies) (*EmailSendTask, error) {
	emailParams, ok := params.(EmailParams)
	if !ok {
		return nil, fmt.Errorf("invalid params passed to EmailTask factory: %T", params)
	}
	return &EmailSendTask{
		Params: &emailParams,
		Deps:   deps,
	}, nil
}
