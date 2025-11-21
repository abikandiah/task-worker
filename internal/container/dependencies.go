package container

import "log/slog"

type GlobalDependencies struct {
	Logger *slog.Logger
}

func NewGlobalDependencies() (*GlobalDependencies, error) {
	return &GlobalDependencies{}, nil
}
