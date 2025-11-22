package service

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/abikandiah/task-worker/internal/domain"
	"github.com/abikandiah/task-worker/internal/domain/task"
	"github.com/abikandiah/task-worker/internal/repository"
	"github.com/google/uuid"
)

type JobService struct {
	JobServiceDependencies
	workerCh chan uuid.UUID
	cancel   context.CancelFunc
	wg       *sync.WaitGroup
}

type JobRepository interface {
	repository.JobRepository
	repository.TaskRunRepository
}

type JobServiceDependencies struct {
	taskFactory *task.TaskFactory
	repository  JobRepository
	logger      *slog.Logger
}

func NewJobService(deps *JobServiceDependencies, capacity int, maxWorkers int) *JobService {
	ctx, cancel := context.WithCancel(context.Background())

	service := &JobService{
		JobServiceDependencies: *deps,
		workerCh:               make(chan uuid.UUID, capacity),
		wg:                     new(sync.WaitGroup),
		cancel:                 cancel,
	}

	// Start Job Workers
	for i := 0; i < maxWorkers; i++ {
		worker := &JobWorker{
			JobServiceDependencies: deps,
			workerCh:               service.workerCh,
		}

		service.wg.Add(1)
		go func() {
			defer service.wg.Done()
			worker.Run(ctx)
		}()
	}

	return service
}

func (service *JobService) SubmitJob(ctx context.Context, submission *domain.JobSubmission) (*domain.Job, error) {
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
	job, err := service.repository.SaveJob(ctx, *job)
	if err != nil {
		service.logger.ErrorContext(ctx, "Failed to save job", slog.Any("error", err))
		return job, fmt.Errorf("failed to save job: %w", err)
	}

	ctx = context.WithValue(ctx, jobIDKey, job.ID)

	// Populate JobID
	for _, taskRun := range submission.TaskRuns {
		taskRun.JobID = job.ID
	}
	service.repository.SaveTaskRuns(ctx, submission.TaskRuns)
	if err != nil {
		service.logger.ErrorContext(ctx, "Failed to save job taskRuns", slog.Any("error", err))
		return job, fmt.Errorf("failed to save job taskRuns: %w", err)
	}

	// Send JobID to Job Worker
	service.workerCh <- job.ID

	return job, nil
}

func (service *JobService) GetJob(ctx context.Context, jobID uuid.UUID) (*domain.Job, error) {
	job, err := service.repository.GetJob(ctx, jobID)
	return job, err
}

func (service *JobService) Close(ctx context.Context) {
	service.logger.InfoContext(ctx, "Closing job service")
	service.cancel()
	service.wg.Wait()
	close(service.workerCh)
}
