package domain

import (
	"context"
)

type Task interface {
	Execute(ctx context.Context) (any, error)
}
