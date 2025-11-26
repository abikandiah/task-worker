package server

import (
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/abikandiah/task-worker/internal/service"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type Server struct {
	*serverDepedencies
	router  *chi.Mux
	limiter *rateLimiter
	apiKeys map[string]struct{}
}

type serverDepedencies struct {
	serverConfig *Config
	jobService   *service.JobService
}

type ServerParams struct {
	ServerConfig *Config
	JobService   *service.JobService
}

func NewServer(deps *ServerParams) *http.Server {
	server := &Server{
		serverDepedencies: &serverDepedencies{
			serverConfig: deps.ServerConfig,
			jobService:   deps.JobService,
		},
		router:  chi.NewRouter(),
		limiter: newRateLimiter(deps.ServerConfig.RateLimit),
		apiKeys: make(map[string]struct{}),
	}

	// Load API key with validation
	workerSecret := os.Getenv("WORKER_SECRET")
	if workerSecret == "" {
		slog.Warn("WORKER_SECRET not set - API authentication may not work")
	} else {
		server.apiKeys[workerSecret] = struct{}{}
	}

	server.setupMiddleware()
	server.setupRoutes()

	config := server.serverConfig
	errorLog := log.New(
		&logWriter{logger: slog.Default()},
		"",
		0,
	)

	httpServer := &http.Server{
		Addr:              fmt.Sprintf("%s:%d", config.Host, config.Port),
		Handler:           server,
		ReadTimeout:       config.ReadTimeout * time.Second,
		WriteTimeout:      config.WriteTimeout * time.Second,
		IdleTimeout:       config.IdleTimeout * time.Second,
		MaxHeaderBytes:    1 << 20,
		ReadHeaderTimeout: 5 * time.Second,
		ErrorLog:          errorLog,
	}

	slog.Info("server initialized", slog.Any("server_config", config))

	server.printRoutes("")

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
		// Authenticated routes
		r.Use(server.authenticateMiddleware)

		r.Route("/jobs", server.setupJobRoutes())
		r.Route("/jobs/configs/", server.setupJobConfigRoutes())
	})
}

func (server *Server) printRoutes(prefix string) {
	// Walk through the router's handler structure
	chi.Walk(server.router, func(method string, routePath string, handler http.Handler, middlewares ...func(http.Handler) http.Handler) error {
		fullPath := prefix + routePath
		if fullPath == "" {
			return nil
		}
		fmt.Printf("Method: %-7s Path: %s\n", method, fullPath)
		return nil
	})
}

// Health check
func (server *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	server.respondJSON(w, http.StatusOK, map[string]string{
		"status": "ok",
		"time":   time.Now().Format(time.RFC3339),
	})
}

type logWriter struct {
	logger *slog.Logger
}

func (lw *logWriter) Write(p []byte) (n int, err error) {
	lw.logger.Error("http server error", slog.String("error", string(p)))
	return len(p), nil
}
