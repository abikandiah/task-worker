package server

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/abikandiah/task-worker/internal/domain"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

func (server *Server) setupJobRoutes() func(r chi.Router) {
	return func(r chi.Router) {
		r.Get("/", server.handleGetJobs)
		r.Post("/", server.handleSubmitJob)

		r.Route("/{id}", func(r chi.Router) {
			r.Use(server.updateRequestContextWithID("id", domain.LKeys.JobID))

			r.Get("/", server.handleGetJob)
			r.Get("/status", server.handleGetJobStatus)
		})
	}
}

// Submit Job
func (server *Server) handleSubmitJob(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var submission domain.JobSubmission
	if err := json.NewDecoder(r.Body).Decode(&submission); err != nil {
		server.logger.WarnContext(ctx, "failed to decode job request", slog.Any("error", err))
		server.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Validate required fields
	if submission.Name == "" {
		server.logger.WarnContext(ctx, "job name missing")
		server.respondError(w, http.StatusBadRequest, "job name is required")
		return
	}

	if len(submission.TaskRuns) == 0 {
		server.logger.WarnContext(ctx, "job taskRuns missing")
		server.respondError(w, http.StatusBadRequest, "job taskRuns is required")
		return
	}

	// Submit job to service
	job, err := server.jobService.SubmitJob(ctx, &submission)
	if err != nil {
		server.logger.ErrorContext(ctx, "failed to submit job", slog.Any("error", err))
		server.respondError(w, http.StatusInternalServerError, "failed to submit job")
		return
	}

	ctx = context.WithValue(ctx, domain.LKeys.JobID, job.ID)
	server.logger.InfoContext(ctx, "job submitted successfully")

	server.respondJSON(w, http.StatusCreated, job)
}

// Get Job by ID
func (server *Server) handleGetJob(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	jobID := parseUUIDOrDefault(chi.URLParam(r, "id"))

	if jobID == uuid.Nil {
		server.respondError(w, http.StatusBadRequest, "job ID is required")
		return
	}

	job, err := server.jobService.GetJob(ctx, jobID)
	if err != nil {
		server.logger.ErrorContext(ctx, "failed to get job", slog.Any("error", err))
		server.respondError(w, http.StatusNotFound, "job not found")
		return
	}

	server.respondJSON(w, http.StatusOK, job)
}

// Get Job Status by ID
func (server *Server) handleGetJobStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	jobID := parseUUIDOrDefault(chi.URLParam(r, "id"))

	if jobID == uuid.Nil {
		server.respondError(w, http.StatusBadRequest, "job ID is required")
		return
	}

	status, err := server.jobService.GetJobStatus(r.Context(), jobID)
	if err != nil {
		server.logger.ErrorContext(ctx, "failed to get job status", slog.Any("error", err))
		server.respondError(w, http.StatusNotFound, "job not found")
		return
	}

	server.respondJSON(w, http.StatusOK, status)
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
