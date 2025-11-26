package task

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/abikandiah/task-worker/internal/domain"
)

type EmailParams struct {
	Subject string
	Body    string
	To      string
}

type EmailDependencies struct{}

type SendEmailTask struct {
	*EmailParams
	*EmailDependencies
}

func SendEmailConstructor(params *EmailParams, deps *EmailDependencies) (domain.Task, error) {
	if params == nil {
		params = &EmailParams{
			Subject: "default-email",
			To:      "example@home",
		}
	}
	if params.To == "" {
		return nil, fmt.Errorf("to is required")
	}

	return &SendEmailTask{
		EmailParams:       params,
		EmailDependencies: deps,
	}, nil
}

func (task *SendEmailTask) Execute(ctx context.Context) (any, error) {
	<-time.After(10 * time.Second)
	slog.InfoContext(ctx, "email", "subject", task.Subject, "to", task.To)
	return nil, nil
}
