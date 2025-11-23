package service

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/abikandiah/task-worker/config"
	"github.com/abikandiah/task-worker/internal/domain"
	"github.com/abikandiah/task-worker/internal/factory"
	"github.com/abikandiah/task-worker/internal/repository"
	"github.com/google/uuid"
)

type JobService struct {
	*JobServiceDependencies
	jobCh  chan uuid.UUID
	taskCh chan TaskRunRequest
	cancel context.CancelFunc
	wg     *sync.WaitGroup
}

type JobRepository interface {
	repository.JobRepository
	repository.TaskRunRepository
}

type JobServiceDependencies struct {
	cfg         config.WorkerConfig
	taskFactory *factory.TaskFactory
	repository  JobRepository
	logger      *slog.Logger
}

func NewJobService(deps *domain.GlobalDependencies, taskFactory *factory.TaskFactory) *JobService {
	ctx, cancel := context.WithCancel(context.Background())

	workerConfig := deps.Config.Worker
	jobServiceDeps := &JobServiceDependencies{
		cfg:         workerConfig,
		logger:      deps.Logger,
		taskFactory: taskFactory,
	}

	service := &JobService{
		JobServiceDependencies: jobServiceDeps,
		jobCh:                  make(chan uuid.UUID, workerConfig.JobBufferCapacity),
		taskCh:                 make(chan TaskRunRequest),
		wg:                     new(sync.WaitGroup),
		cancel:                 cancel,
	}

	// Start Task Workers
	for i := 0; i < workerConfig.TaskWorkerCount; i++ {
		worker := &TaskWorker{
			JobServiceDependencies: jobServiceDeps,
			taskCh:                 service.taskCh,
		}

		service.wg.Add(1)
		go func() {
			defer service.wg.Done()
			worker.Run(ctx)
		}()
	}

	// Start Job Workers
	for i := 0; i < workerConfig.JobWorkerCount; i++ {
		worker := &JobWorker{
			JobServiceDependencies: jobServiceDeps,
			jobCh:                  service.jobCh,
			taskCh:                 service.taskCh,
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

	ctx = context.WithValue(ctx, LKeys.JobID, job.ID)

	// Populate JobID
	for i := range submission.TaskRuns {
		submission.TaskRuns[i].JobID = job.ID
	}
	service.repository.SaveTaskRuns(ctx, submission.TaskRuns)
	if err != nil {
		service.logger.ErrorContext(ctx, "Failed to save job taskRuns", slog.Any("error", err))
		return job, fmt.Errorf("failed to save job taskRuns: %w", err)
	}

	// Send JobID to Job Worker
	service.jobCh <- job.ID

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
	close(service.jobCh)
	close(service.taskCh)
}
