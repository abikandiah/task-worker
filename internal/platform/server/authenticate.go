package server

import (
	"net/http"
	"strings"
)

func (server *Server) authenticate(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Extract API key from Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			server.respondError(w, http.StatusUnauthorized, "Missing authorization header")
			return
		}

		// Expected format: "Bearer <api-key>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			server.respondError(w, http.StatusUnauthorized, "Invalid authorization format")
			return
		}

		apiKey := parts[1]
		_, ok := server.apiKeys[apiKey]

		if !ok {
			server.respondError(w, http.StatusUnauthorized, "Invalid API key")
			return
		}

		next(w, r)
	}
}
