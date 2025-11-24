package server

import (
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"net/http"
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
	router *http.ServeMux
}

type serverDepedencies struct {
	config *config.ServerConfig
	logger *slog.Logger
}

type ServerParams struct {
	Config *config.ServerConfig
	Logger *slog.Logger
}

func NewServer(deps *ServerParams) *http.Server {
	server := &Server{
		serverDepedencies: &serverDepedencies{
			config: deps.Config,
			logger: deps.Logger,
		},
		router: http.NewServeMux(),
	}

	server.routes()
	config := server.config

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

	server.logger.Info("server initialized", "addr", addr, "timeouts", timeouts)

	httpServer := &http.Server{
		Addr:         addr,
		Handler:      server.middleware(server),
		ReadTimeout:  config.ReadTimeout * time.Second,
		WriteTimeout: config.WriteTimeout * time.Second,
		IdleTimeout:  config.IdleTimeout * time.Second,
	}

	return httpServer
}

// Set up API routes
func (s *Server) routes() {
	s.router.HandleFunc("/health", s.handleHealth())
	// s.router.HandleFunc("/api/items/", nil)
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}

// Middleware chain
func (s *Server) middleware(next http.Handler) http.Handler {
	return s.logging(s.cors(s.recovery(next)))
}

// logging middleware logs each request
func (s *Server) logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %s", r.Method, r.URL.Path, time.Since(start))
	})
}

// cors middleware adds CORS headers
func (s *Server) cors(next http.Handler) http.Handler {
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
func (s *Server) recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("panic: %v", err)
				s.respondError(w, http.StatusInternalServerError, "Internal server error")
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// handleHealth returns health check endpoint
func (s *Server) handleHealth() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s.respondJSON(w, http.StatusOK, map[string]string{
			"status": "ok",
			"time":   time.Now().Format(time.RFC3339),
		})
	}
}

// respondJSON sends a JSON response
func (s *Server) respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("error encoding response: %v", err)
	}
}

// respondError sends an error response
func (s *Server) respondError(w http.ResponseWriter, status int, message string) {
	s.respondJSON(w, status, ErrorResponse{
		Error:   http.StatusText(status),
		Message: message,
	})
}
