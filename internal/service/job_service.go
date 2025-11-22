package service

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/abikandiah/task-worker/internal/domain"
	"github.com/abikandiah/task-worker/internal/domain/task"
	"github.com/abikandiah/task-worker/internal/repository"
	"github.com/google/uuid"
)

type JobService struct {
	JobServiceDependencies
	workerCh chan uuid.UUID
}

type JobRepository interface {
	repository.JobRepository
	repository.TaskRunRepository
}

type JobServiceDependencies struct {
	taskFactory task.TaskFactory
	repository  JobRepository
	logger      *slog.Logger
}

func NewJobScheduler(maxWorkers int, deps *JobServiceDependencies) *JobService {
	// Init Job Worker

	return &JobService{
		JobServiceDependencies: *deps,
		workerCh:               make(chan uuid.UUID, maxWorkers),
	}
}

func (scheduler *JobService) SubmitJob(ctx context.Context, submission *domain.JobSubmission) (*domain.Job, error) {
	// Translate submission into Job (validate and populate IDs etc.)
	job := &domain.Job{
		Identity: domain.Identity{
			IdentitySubmission: domain.IdentitySubmission{
				Name:        submission.Name,
				Description: submission.Description,
			},
		},
		SubmitDate: time.Now(),
	}

	// Write to DB
	job, err := scheduler.repository.SaveJob(ctx, *job)
	if err != nil {
		scheduler.logger.ErrorContext(ctx, "Failed to save job", slog.Any("error", err))
		return job, fmt.Errorf("failed to save job: %w", err)
	}

	ctx = context.WithValue(ctx, jobIDKey, job.ID)

	// Populate JobID
	for _, taskRun := range submission.TaskRuns {
		taskRun.JobID = job.ID
	}
	scheduler.repository.SaveTaskRuns(ctx, submission.TaskRuns)
	if err != nil {
		scheduler.logger.ErrorContext(ctx, "Failed to save job taskRuns", slog.Any("error", err))
		return job, fmt.Errorf("failed to save job taskRuns: %w", err)
	}

	// Send JobID to Job Worker
	scheduler.workerCh <- job.ID

	return job, nil
}

func (scheduler *JobService) GetJob(ctx context.Context, jobID uuid.UUID) (*domain.Job, error) {
	job, err := scheduler.repository.GetJob(ctx, jobID)
	return job, err
}
