package server

import (
	"net/http"
	"strconv"
	"time"

	"github.com/abikandiah/task-worker/internal/domain"
	"github.com/google/uuid"
)

// Health check
func (server *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, r, http.StatusOK, map[string]string{
		"status": "ok",
		"time":   time.Now().Format(time.RFC3339),
	})
}

// Submit Job
func (Server *Server) submitJob(w http.ResponseWriter, r *http.Request) {
}

// Get Jobs
func (server *Server) handleGetJobs(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	query := r.URL.Query()

	afterID, _ := uuid.Parse(query.Get("afterId"))
	beforeID, _ := uuid.Parse(query.Get("beforeId"))
	limit, _ := strconv.Atoi(query.Get("limit"))

	inputReq := &domain.CursorInput{
		AfterID:   afterID,
		BeforeID:  beforeID,
		Limit:     limit,
		SortField: "id",
		SortDir:   "ASC",
	}
	inputReq.SetDefaults()

	res, err := server.jobService.GetAllJobs(ctx, inputReq)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respondJSON(w, r, http.StatusOK, res)
}
