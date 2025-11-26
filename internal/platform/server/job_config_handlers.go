package server

import (
	"log/slog"
	"net/http"

	"github.com/abikandiah/task-worker/internal/domain"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

func (server *Server) setupJobConfigRoutes() func(r chi.Router) {
	return func(r chi.Router) {
		r.Route("/{id}", func(r chi.Router) {
			r.Use(server.updateRequestContextWithID("id", domain.LKeys.ConfigID))

			r.Get("/", server.handleGetJobConfig)
		})
	}
}

// Get JobConfig by ID
func (server *Server) handleGetJobConfig(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	configID := parseUUIDOrDefault(chi.URLParam(r, "id"))

	if configID == uuid.Nil {
		server.respondError(w, http.StatusBadRequest, "config ID is required")
		return
	}

	config, err := server.jobService.GetJobConfig(ctx, configID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get config", slog.Any("error", err))
		server.respondError(w, http.StatusNotFound, "config not found")
		return
	}

	server.respondJSON(w, http.StatusOK, config)
}
