package server

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5/middleware"
)

// Send a JSON response
func respondJSON(w http.ResponseWriter, r *http.Request, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		// Retrieve RequestID from the context for better debugging
		requestContext := r.Context()
		requestID := middleware.GetReqID(requestContext)
		logger := GetRequestLogger(requestContext)
		logger.ErrorContext(requestContext, "error encoding response", "reqID", requestID, slog.Any("error", err))
	}
}

// Send an error response
func respondError(w http.ResponseWriter, r *http.Request, status int, message string) {
	respondJSON(w, r, status, ErrorResponse{
		Error:   http.StatusText(status),
		Message: message,
	})
}
