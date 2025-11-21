package http

import "github.com/abikandiah/task-worker/internal/core/domain"

// Defines a struct that holds dependencies (like the JobService interface)
type JobHandler struct {
	// service internal/core/service.JobService
}

type TaskRunRequest struct {
	domain.Identity
	TaskID      string
	TaskVersion string
	DependentOn []int
	Options     map[string]any
}

type JobRequest struct {
	domain.Identity
	Tasks []TaskRunRequest
}
