package factory

import (
	"github.com/abikandiah/task-worker/config"
	"github.com/abikandiah/task-worker/internal/domain"
	"github.com/abikandiah/task-worker/internal/platform/logging"
)

func NewGlobalDependencies() (*domain.GlobalDependencies, error) {
	cfg := config.MustLoad()
	logger := logging.SetupLogger(cfg.Logger)

	return &domain.GlobalDependencies{
		Config: cfg,
		Logger: logger,
	}, nil
}
