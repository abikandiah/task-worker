package server

import (
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/abikandiah/task-worker/config"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}

type Server struct {
	*serverDepedencies
	router  *chi.Mux
	limiter *rateLimiter
	apiKeys map[string]struct{}
}

type serverDepedencies struct {
	serverConfig *config.ServerConfig
	logger       *slog.Logger
}

type ServerParams struct {
	ServerConfig    *config.ServerConfig
	RateLimitConfig *config.RateLimitConfig
	Logger          *slog.Logger
}

func NewServer(deps *ServerParams) *http.Server {
	server := &Server{
		serverDepedencies: &serverDepedencies{
			serverConfig: deps.ServerConfig,
			logger:       deps.Logger,
		},
		router:  chi.NewRouter(),
		limiter: newRateLimiter(deps.RateLimitConfig),
		apiKeys: make(map[string]struct{}),
	}

	server.setupMiddleware()
	server.setupRoutes()

	config := server.serverConfig
	httpServer := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", config.Host, config.Port),
		Handler:      server,
		ReadTimeout:  config.ReadTimeout * time.Second,
		WriteTimeout: config.WriteTimeout * time.Second,
		IdleTimeout:  config.IdleTimeout * time.Second,
	}

	server.logger.Info("server initialized", "config", *config)
	server.logger.Info("rate limiter initialized", "requests_per_second", server.limiter.rate, "burst", server.limiter.burst)

	return httpServer
}

func (server *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	server.router.ServeHTTP(w, r)
}

// Middleware chain
func (server *Server) setupMiddleware() {
	// Chi's built-in middleware
	server.router.Use(middleware.RequestID)
	server.router.Use(middleware.RealIP)
	server.router.Use(middleware.Logger)
	server.router.Use(middleware.Recoverer)

	// Custom middleware
	server.router.Use(server.corsMiddleware)
	server.router.Use(server.rateLimitMiddleware)

	// Timeout middleware
	server.router.Use(middleware.Timeout(server.serverConfig.Timeout * time.Second))
}

func (s *Server) contentTypeMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost || r.Method == http.MethodPut {
			contentType := r.Header.Get("Content-Type")
			if !strings.HasPrefix(contentType, "application/json") {
				s.respondError(w, http.StatusUnsupportedMediaType,
					"Content-Type must be application/json")
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}

// corsMiddleware adds CORS headers
func (s *Server) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// Health check
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	s.respondJSON(w, http.StatusOK, map[string]string{
		"status": "ok",
		"time":   time.Now().Format(time.RFC3339),
	})
}

// Send a JSON response
func (server *Server) respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("error encoding response: %v", err)
	}
}

// Send an error response
func (server *Server) respondError(w http.ResponseWriter, status int, message string) {
	server.respondJSON(w, status, ErrorResponse{
		Error:   http.StatusText(status),
		Message: message,
	})
}
