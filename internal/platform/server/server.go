package server

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/abikandiah/task-worker/config"
	"github.com/abikandiah/task-worker/internal/service"
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
	jobService   *service.JobService
}

type ServerParams struct {
	ServerConfig    *config.ServerConfig
	RateLimitConfig *config.RateLimitConfig
	Logger          *slog.Logger
	JobService      *service.JobService
}

func NewServer(deps *ServerParams) *http.Server {
	server := &Server{
		serverDepedencies: &serverDepedencies{
			serverConfig: deps.ServerConfig,
			logger:       deps.Logger,
			jobService:   deps.JobService,
		},
		router:  chi.NewRouter(),
		limiter: newRateLimiter(deps.RateLimitConfig),
		apiKeys: make(map[string]struct{}),
	}

	server.apiKeys[os.Getenv("WORKER_SECRET")] = struct{}{}
	server.setupMiddleware()
	server.setupRoutes()

	config := server.serverConfig
	httpServer := &http.Server{
		Addr:              fmt.Sprintf("%s:%d", config.Host, config.Port),
		Handler:           server,
		ReadTimeout:       config.ReadTimeout * time.Second,
		WriteTimeout:      config.WriteTimeout * time.Second,
		IdleTimeout:       config.IdleTimeout * time.Second,
		MaxHeaderBytes:    1 << 20,
		ReadHeaderTimeout: 5 * time.Second,
	}

	server.logger.Info("server initialized", slog.Any("server_config", config))
	server.logger.Info("rate limiter initialized", slog.Any("rate_limit_config", deps.RateLimitConfig))

	return httpServer
}

func (server *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	server.router.ServeHTTP(w, r)
}

func (server *Server) setupMiddleware() {
	// Chi's built-in middleware
	server.router.Use(middleware.Recoverer)
	server.router.Use(middleware.RequestID)
	server.router.Use(middleware.RealIP)
	server.router.Use(middleware.StripSlashes)

	// Custom middleware
	server.router.Use(server.loggerMiddleware)
	server.router.Use(configureCORSMiddleware(server.serverConfig.Cors))
	server.router.Use(server.rateLimitMiddleware)
	server.router.Use(server.contentTypeMiddleware)

	// Timeout middleware
	server.router.Use(middleware.Timeout(server.serverConfig.Timeout * time.Second))
}

// Set up API routes
func (server *Server) setupRoutes() {
	server.router.Get("/health", server.handleHealth)

	server.router.Route("/api/v1", func(r chi.Router) {
		// Authenticate routes
		r.Use(server.authenticateMiddleware)

		// Job endpoints
		r.Route("/job", func(r chi.Router) {
			r.Get("/", server.handleGetJobs)
			// r.Post("/", server.createItem)

			// r.Route("/{id}", func(r chi.Router) {
			// r.Get("/", server.getItem)
			// r.Put("/", server.updateItem)
			// r.Delete("/", server.deleteItem)
		})
	})

	// Easy to add more resource groups
	// r.Route("/users", func(r chi.Router) { ... })
}
