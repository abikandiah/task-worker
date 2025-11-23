package domain

import (
	"log/slog"

	"github.com/abikandiah/task-worker/config"
)

type GlobalDependencies struct {
	Config *config.Config
	Logger *slog.Logger
}
