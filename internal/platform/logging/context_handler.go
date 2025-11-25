package logging

import (
	"context"
	"log/slog"

	"github.com/abikandiah/task-worker/internal/domain"
)

type ContextHandler struct {
	slog.Handler
}

func NewContextHandler(h slog.Handler) *ContextHandler {
	return &ContextHandler{
		Handler: h,
	}
}

func (h *ContextHandler) Handle(ctx context.Context, r slog.Record) error {
	// Iterate over the LKeys and add them to the record
	for _, logKey := range domain.ContextLKeys {

		val := ctx.Value(logKey)
		if val != nil {
			r.AddAttrs(slog.Any(string(logKey), val))
		}
	}
	return h.Handler.Handle(ctx, r)
}
