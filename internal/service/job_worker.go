package service

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/abikandiah/task-worker/internal/domain"
	"github.com/google/uuid"
)

type JobWorker struct {
	JobServiceDependencies
	workerCh chan uuid.UUID
}

type JobTask struct {
	ID    uuid.UUID
	JobID uuid.UUID
}

func (worker *JobWorker) Run() error {
	for jobID := range worker.workerCh {

		ctx := context.WithValue(context.Background(), jobIDKey, jobID)

		// Get job
		job, err := worker.repository.GetJob(ctx, jobID)
		if err != nil {
			worker.logger.ErrorContext(ctx, "Failed to fetch job", slog.Any("error", err))
			return fmt.Errorf("failed to fetch job %s: %w", jobID, err)
		}

		ctx = context.WithValue(ctx, jobNameKey, job.Name)
		ctx = context.WithValue(ctx, configIDKey, job.ConfigID)

		// Get config
		config, err := worker.repository.GetJobConfig(ctx, job.ConfigID)
		if err != nil {
			worker.logger.ErrorContext(ctx, "Failed to fetch config", slog.Any("error", err))
			return fmt.Errorf("failed to fetch config %s: %w", job.ConfigID, err)
		}

		worker.logger.InfoContext(ctx, "Job Started")
		job.StartDate = time.Now()
		job.State = domain.StateRunning

		ctxTimeout, cancel := context.WithTimeout(ctx, (time.Duration(config.JobTimeout) * time.Second))
		defer cancel() // Always schedule cleanup to release resources

		jobCh := make(chan struct{})
		go worker.runJob(ctx, job, jobCh)

		select {
		case <-jobCh:
			// Job finished
			worker.logger.InfoContext(ctx, "Job Finished")
			job.State = domain.StateFinished

		case <-ctxTimeout.Done():
			// Context cancelled / timed out
			worker.logger.InfoContext(ctx, "Job Stopped", jobIDKey, job.ID)
			job.State = domain.StateStopped
		}

		job.EndDate = time.Now()
	}

	return nil
}

func (worker *JobWorker) runJob(ctx context.Context, job *domain.Job, resultCh chan struct{}) {
	<-time.After(10 * time.Second)

	resultCh <- struct{}{}
}
