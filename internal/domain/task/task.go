package task

import (
	"context"
)

type Task interface {
	Execute(ctx context.Context) error
}
