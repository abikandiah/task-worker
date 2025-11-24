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
)

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}

type Server struct {
	*serverDepedencies
	router  *http.ServeMux
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
		router:  http.NewServeMux(),
		limiter: newRateLimiter(deps.RateLimitConfig),
		apiKeys: make(map[string]struct{}),
	}

	server.routes()
	config := server.serverConfig

	addr := fmt.Sprintf("%s:%d", config.Host, config.Port)
	timeouts := struct {
		read  string
		write string
		idle  string
	}{
		read:  config.ReadTimeout.String(),
		write: config.WriteTimeout.String(),
		idle:  config.IdleTimeout.String(),
	}

	httpServer := &http.Server{
		Addr:         addr,
		Handler:      server.middleware(server),
		ReadTimeout:  config.ReadTimeout * time.Second,
		WriteTimeout: config.WriteTimeout * time.Second,
		IdleTimeout:  config.IdleTimeout * time.Second,
	}

	server.logger.Info("server initialized", "addr", addr, "timeouts", timeouts)
	server.logger.Info("rate limiter initialized", "requests_per_second", server.limiter.rate, "burst", server.limiter.burst)

	return httpServer
}

// Set up API routes
func (server *Server) routes() {
	server.router.HandleFunc("/health", server.handleHealth())

	server.router.HandleFunc("/api/v1/items/", func(w http.ResponseWriter, r *http.Request) {})
}

func (server *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	server.router.ServeHTTP(w, r)
}

// Middleware chain
func (server *Server) middleware(next http.Handler) http.Handler {
	return server.logging(server.cors(server.recovery(server.rateLimit(next))))
}

func (server *Server) validateContentType(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost || r.Method == http.MethodPut {
			contentType := r.Header.Get("Content-Type")
			if !strings.HasPrefix(contentType, "application/json") {
				server.respondError(w, http.StatusUnsupportedMediaType,
					"Content-Type must be application/json")
				return
			}
		}
		next(w, r)
	}
}

// logging middleware logs each request
func (server *Server) logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %s", r.Method, r.URL.Path, time.Since(start))
	})
}

// cors middleware adds CORS headers
func (server *Server) cors(next http.Handler) http.Handler {
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

// recovery middleware recovers from panics
func (server *Server) recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("panic: %v", err)
				server.respondError(w, http.StatusInternalServerError, "Internal server error")
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// handleHealth returns health check endpoint
func (server *Server) handleHealth() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		server.respondJSON(w, http.StatusOK, map[string]string{
			"status": "ok",
			"time":   time.Now().Format(time.RFC3339),
		})
	}
}

// respondJSON sends a JSON response
func (server *Server) respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("error encoding response: %v", err)
	}
}

// respondError sends an error response
func (server *Server) respondError(w http.ResponseWriter, status int, message string) {
	server.respondJSON(w, status, ErrorResponse{
		Error:   http.StatusText(status),
		Message: message,
	})
}
