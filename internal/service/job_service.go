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
	"github.com/google/uuid"
)

type JobService struct {
	*jobServiceDependencies
	jobCh   chan uuid.UUID
	taskCh  chan TaskRunRequest
	cancel  context.CancelFunc
	wg      *sync.WaitGroup
	started bool
}

type jobServiceDependencies struct {
	config      *config.WorkerConfig
	repository  domain.ServiceRepository
	taskFactory *factory.TaskFactory
	logger      *slog.Logger
}

type JobServiceParams struct {
	Config      *config.WorkerConfig
	Repository  domain.ServiceRepository
	TaskFactory *factory.TaskFactory
	Logger      *slog.Logger
}

func NewJobService(params *JobServiceParams) *JobService {
	jobServiceDeps := &jobServiceDependencies{
		config:      params.Config,
		logger:      params.Logger,
		taskFactory: params.TaskFactory,
		repository:  params.Repository,
	}

	service := &JobService{
		jobServiceDependencies: jobServiceDeps,
		jobCh:                  make(chan uuid.UUID, params.Config.JobBufferCapacity),
		taskCh:                 make(chan TaskRunRequest),
		wg:                     new(sync.WaitGroup),
	}
	return service
}

func (service *JobService) StartWorkers(ctx context.Context) {
	ctx, cancel := context.WithCancel(context.Background())
	service.cancel = cancel
	service.started = true

	// Start Task Workers
	for i := 0; i < service.config.TaskWorkerCount; i++ {
		worker := &TaskWorker{
			jobServiceDependencies: service.jobServiceDependencies,
			taskCh:                 service.taskCh,
		}

		service.wg.Add(1)
		go func() {
			defer service.wg.Done()
			worker.Run(ctx)
		}()
	}

	// Start Job Workers
	for i := 0; i < service.config.JobWorkerCount; i++ {
		worker := &JobWorker{
			jobServiceDependencies: service.jobServiceDependencies,
			jobCh:                  service.jobCh,
			taskCh:                 service.taskCh,
		}

		service.wg.Add(1)
		go func() {
			defer service.wg.Done()
			worker.Run(ctx)
		}()
	}
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
	_, err = service.repository.SaveTaskRuns(ctx, submission.TaskRuns)
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

func (service *JobService) GetAllJobs(ctx context.Context, input *domain.CursorInput) (*domain.CursorOutput[domain.Job], error) {
	output, err := service.repository.GetAllJobs(ctx, input)
	return output, err
}

func (service *JobService) Close(ctx context.Context) {
	service.logger.InfoContext(ctx, "Closing job service")
	close(service.jobCh)
	close(service.taskCh)

	if service.started {
		service.cancel()
		service.wg.Wait()
	}
}
