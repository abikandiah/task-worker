package server

import (
	"context"
	"errors"
	"net/http"

	"github.com/abikandiah/task-worker/internal/domain"
)

const userLogKey domain.LogKey = "user"

type UserInfo struct {
	Name string
}

func (server *Server) authenticateMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get token from the Authorization header (e.g., Bearer <token>)
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" || len(authHeader) < 7 || authHeader[:6] != "Bearer" {
			http.Error(w, "Invalid authorization format", http.StatusUnauthorized)
			return // Stop execution
		}

		token := authHeader[7:]

		// Validate the token and retrieve user information
		user, err := server.validateToken(token)
		if err != nil {
			http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
			return // Stop execution
		}

		// Set authenticated user in context
		ctx := context.WithValue(r.Context(), userLogKey, user)
		// Update context and proceed
		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	})
}

func (server *Server) validateToken(token string) (*UserInfo, error) {
	_, ok := server.apiKeys[token]
	server.logger.Info("", "server", server.apiKeys)
	if !ok {
		return nil, errors.New("invalid API key")
	}

	return nil, nil
}
