package server

import (
	"context"
	"mime"
	"net/http"
	"time"

	"github.com/abikandiah/task-worker/internal/domain"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/cors"
)

func (server *Server) loggerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Update request context with values
		requestID := middleware.GetReqID(r.Context())
		ctx := context.WithValue(r.Context(), domain.LKeys.RequestID, requestID)
		ctx = context.WithValue(ctx, domain.LKeys.Method, r.Method)
		ctx = context.WithValue(ctx, domain.LKeys.Path, r.URL.Path)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (server *Server) contentTypeMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Use standard net/http constants for clarity
		if r.Method == http.MethodPost || r.Method == http.MethodPut {
			contentType := r.Header.Get("Content-Type")

			// Be more robust: check if Content-Type is missing entirely (often a bad request)
			if contentType == "" {
				server.respondError(w, http.StatusBadRequest, "Content-Type header is required")
				return
			}

			// The standard library provides a function for robust media type parsing
			mediaType, _, err := mime.ParseMediaType(contentType)
			if err != nil || mediaType != "application/json" {
				server.respondError(w, http.StatusUnsupportedMediaType, "Content-Type must be application/json")
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}

func configureCORSMiddleware(cfg *CORSConfig) func(http.Handler) http.Handler {
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
