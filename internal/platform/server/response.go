package server

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/abikandiah/task-worker/internal/domain"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}

func (server *Server) updateRequestContextWithID(idKey string, logKey domain.LogKey) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			id := parseUUIDOrDefault(chi.URLParam(r, idKey))

			if id != uuid.Nil {
				// Store ID in context
				ctx := context.WithValue(r.Context(), logKey, id)
				next.ServeHTTP(w, r.WithContext(ctx))
			}
		})
	}
}

// Send a JSON response
func (server *Server) respondJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		server.logger.Error("failed to encode response",
			slog.String("error", err.Error()),
		)
	}
}

func (server *Server) respondError(w http.ResponseWriter, status int, message string) {
	server.respondJSON(w, status, ErrorResponse{
		Error:   http.StatusText(status),
		Message: message,
	})
}

// parseUUIDOrDefault attempts to parse a string as a UUID.
// If parsing fails (invalid format), it returns uuid.Nil, effectively defaulting the cursor.
func parseUUIDOrDefault(s string) uuid.UUID {
	if s == "" {
		return uuid.Nil
	}
	id, err := uuid.Parse(s)
	if err != nil {
		return uuid.Nil
	}
	return id
}
