package domain

import "log/slog"

type GlobalDependencies struct {
	Logger *slog.Logger
}
