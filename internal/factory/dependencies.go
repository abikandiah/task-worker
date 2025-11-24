package factory

import (
	"github.com/abikandiah/task-worker/config"
	"github.com/abikandiah/task-worker/internal/domain"
	"github.com/abikandiah/task-worker/internal/platform/logging"
)

func NewGlobalDependencies() (*domain.GlobalDependencies, error) {
	config := config.MustLoad()

	logger := logging.SetupLogger(logging.LoggerParams{
		Level:       config.Logger.Level,
		Environment: config.Environment,
		ServiceName: config.ServiceName,
		Version:     config.Version,
	})

	return &domain.GlobalDependencies{
		Config: config,
		Logger: logger,
	}, nil
}
