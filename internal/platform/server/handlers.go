package server

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/abikandiah/task-worker/internal/domain"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// Health check
func (server *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	server.respondJSON(w, http.StatusOK, map[string]string{
		"status": "ok",
		"time":   time.Now().Format(time.RFC3339),
	})
}

func (server *Server) setupJobRoutes() func(r chi.Router) {
	return func(r chi.Router) {
		r.Get("/", server.handleGetJobs)
		r.Post("/", server.handleSubmitJob)

		r.Route("/{id}", func(r chi.Router) {
			r.Use(server.updateJobRequestLoggerMiddleware)

			r.Get("/", server.handleGetJob)
			r.Get("/status", server.handleGetJobStatus)
		})
	}
}

func (server *Server) updateJobRequestLoggerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		jobID := parseUUIDOrDefault(chi.URLParam(r, "jobID"))

		if jobID != uuid.Nil {
			logger := GetRequestLogger(r.Context())
			jobLogger := logger.With(
				slog.String("job_id", jobID.String()),
			)

			// Store jobLogger in context and update request context
			ctx := context.WithValue(r.Context(), requestLoggerKey, jobLogger)
			next.ServeHTTP(w, r.WithContext(ctx))
		}
	})
}

// Submit Job
func (server *Server) handleSubmitJob(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := GetRequestLogger(ctx)

	var submission domain.JobSubmission
	if err := json.NewDecoder(r.Body).Decode(&submission); err != nil {
		logger.Warn("failed to decode job request",
			slog.String("error", err.Error()),
		)
		server.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Validate required fields
	if submission.Name == "" {
		logger.Warn("job name missing")
		server.respondError(w, http.StatusBadRequest, "job name is required")
		return
	}

	if len(submission.TaskRuns) == 0 {
		logger.Warn("job taskRuns missing")
		server.respondError(w, http.StatusBadRequest, "job taskRuns is required")
		return
	}

	// Submit job to service
	job, err := server.jobService.SubmitJob(ctx, &submission)
	if err != nil {
		logger.Error("failed to submit job", slog.String("error", err.Error()))
		server.respondError(w, http.StatusInternalServerError, "failed to submit job")
		return
	}

	logger.Info("job submitted successfully", slog.String("job_id", job.ID.String()))

	server.respondJSON(w, http.StatusCreated, job)
}

// Get Job by ID
func (server *Server) handleGetJob(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := GetRequestLogger(ctx)

	jobID := parseUUIDOrDefault(chi.URLParam(r, "jobID"))

	if jobID == uuid.Nil {
		server.respondError(w, http.StatusBadRequest, "job ID is required")
		return
	}

	job, err := server.jobService.GetJob(ctx, jobID)
	if err != nil {
		logger.Error("failed to get job", slog.String("error", err.Error()))
		server.respondError(w, http.StatusNotFound, "job not found")
		return
	}

	server.respondJSON(w, http.StatusOK, job)
}

// handleGetJobStatus retrieves job status by ID
func (server *Server) handleGetJobStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := GetRequestLogger(ctx)

	jobID := parseUUIDOrDefault(chi.URLParam(r, "jobID"))

	if jobID == uuid.Nil {
		server.respondError(w, http.StatusBadRequest, "job ID is required")
		return
	}

	status, err := server.jobService.GetJobStatus(r.Context(), jobID)
	if err != nil {
		logger.Error("failed to get job status", slog.String("error", err.Error()))
		server.respondError(w, http.StatusNotFound, "job not found")
		return
	}

	server.respondJSON(w, http.StatusOK, map[string]any{
		"job_id": jobID,
		"status": status,
	})
}

func (server *Server) handleGetJobs(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	query := r.URL.Query()

	afterID := parseUUIDOrDefault(query.Get("afterId"))
	beforeID := parseUUIDOrDefault(query.Get("beforeId"))
	limit, _ := strconv.Atoi(query.Get("limit"))

	sortField := query.Get("sortField")
	sortDir := domain.SortDirection(query.Get("sortDir"))

	inputReq := &domain.CursorInput{
		AfterID:   afterID,
		BeforeID:  beforeID,
		Limit:     limit,
		SortField: sortField,
		SortDir:   sortDir,
	}
	inputReq.SetDefaults()

	res, err := server.jobService.GetAllJobs(ctx, inputReq)
	if err != nil {
		http.Error(w, "Failed to retrieve jobs.", http.StatusInternalServerError)
		return
	}

	server.respondJSON(w, http.StatusOK, res)
}
