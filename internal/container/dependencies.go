package container

import "github.com/abikandiah/task-worker/internal/domain"

func NewGlobalDependencies() (*domain.GlobalDependencies, error) {
	return &domain.GlobalDependencies{}, nil
}
