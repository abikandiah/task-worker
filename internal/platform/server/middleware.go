package server

import (
	"context"
	"log/slog"
	"mime"
	"net/http"
	"time"

	"github.com/abikandiah/task-worker/config"
	"github.com/abikandiah/task-worker/internal/domain"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/cors"
)

const requestLoggerKey domain.LogKey = "requestLogger"

func (server *Server) loggerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Create request-specific logger based on logger
		requestID := middleware.GetReqID(r.Context())
		requestLogger := server.logger.With(
			slog.String("request_ID", requestID),
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
		)

		// Store requestLogger in context and update request context
		ctx := context.WithValue(r.Context(), requestLoggerKey, requestLogger)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// Get logger for use in handlers/response functions
func GetRequestLogger(ctx context.Context) *slog.Logger {
	if logger, ok := ctx.Value(requestLoggerKey).(*slog.Logger); ok {
		return logger
	}
	return slog.Default()
}

func (server *Server) contentTypeMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Use standard net/http constants for clarity
		if r.Method == http.MethodPost || r.Method == http.MethodPut {
			contentType := r.Header.Get("Content-Type")

			// Be more robust: check if Content-Type is missing entirely (often a bad request)
			if contentType == "" {
				respondError(w, r, http.StatusBadRequest, "Content-Type header is required")
				return
			}

			// The standard library provides a function for robust media type parsing
			mediaType, _, err := mime.ParseMediaType(contentType)
			if err != nil || mediaType != "application/json" {
				respondError(w, r, http.StatusUnsupportedMediaType, "Content-Type must be application/json")
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}

func configureCORSMiddleware(cfg *config.CORSConfig) func(http.Handler) http.Handler {
	if !cfg.Enabled {
		return func(next http.Handler) http.Handler {
			return next // Simply pass the request along
		}
	}

	options := cors.Options{
		AllowedOrigins:   cfg.AllowedOrigins,
		AllowedMethods:   cfg.AllowedMethods,
		AllowedHeaders:   cfg.AllowedHeaders,
		AllowCredentials: cfg.AllowCredentials,
		MaxAge:           int(time.Second * cfg.MaxAge),
	}
	c := cors.New(options)
	return c.Handler
}
