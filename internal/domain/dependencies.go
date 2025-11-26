package domain

import (
	"github.com/abikandiah/task-worker/config"
)

type GlobalDependencies struct {
	Config      *config.Config
	Repository  ServiceRepository
	TaskFactory TaskFactory
}
