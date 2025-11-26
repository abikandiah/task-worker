package service

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/abikandiah/task-worker/internal/domain"
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
	config      *Config
	repository  domain.ServiceRepository
	taskFactory domain.TaskFactory
}

type JobServiceParams struct {
	Config      *Config
	Repository  domain.ServiceRepository
	TaskFactory domain.TaskFactory
}

func NewJobService(params *JobServiceParams) *JobService {
	jobServiceDeps := &jobServiceDependencies{
		config:      params.Config,
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

	slog.InfoContext(ctx, "started workers", "jobWorkerCount",
		service.config.JobWorkerCount, "taskWorkerCounter", service.config.TaskWorkerCount)
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
		slog.ErrorContext(ctx, "Failed to save job", slog.Any("error", err))
		return job, fmt.Errorf("failed to save job: %w", err)
	}

	ctx = context.WithValue(ctx, domain.LKeys.JobID, job.ID)

	// Populate JobID
	for i := range submission.TaskRuns {
		submission.TaskRuns[i].JobID = job.ID
	}
	_, err = service.repository.SaveTaskRuns(ctx, submission.TaskRuns)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to save job taskRuns", slog.Any("error", err))
		return job, fmt.Errorf("failed to save job taskRuns: %w", err)
	}

	// Send JobID to Job Worker
	slog.InfoContext(ctx, "submitted job to queue")
	service.jobCh <- job.ID

	return job, nil
}

func (service *JobService) GetJob(ctx context.Context, jobID uuid.UUID) (*domain.Job, error) {
	job, err := service.repository.GetJob(ctx, jobID)
	return job, err
}

func (service *JobService) GetJobStatus(ctx context.Context, jobID uuid.UUID) (*domain.JobStatus, error) {
	if job, err := service.GetJob(ctx, jobID); err != nil {
		return nil, err
	} else {
		return &job.JobStatus, nil
	}
}

func (service *JobService) GetAllJobs(ctx context.Context, input *domain.CursorInput) (*domain.CursorOutput[domain.Job], error) {
	output, err := service.repository.GetAllJobs(ctx, input)
	return output, err
}

func (service *JobService) GetJobConfig(ctx context.Context, configID uuid.UUID) (*domain.JobConfig, error) {
	job, err := service.repository.GetJobConfig(ctx, configID)
	return job, err
}

func (service *JobService) Close(ctx context.Context) {
	slog.InfoContext(ctx, "Closing job service")
	close(service.jobCh)
	close(service.taskCh)

	if service.started {
		service.cancel()
		service.wg.Wait()
	}
}
