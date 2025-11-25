package server

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/google/uuid"
)

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
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
