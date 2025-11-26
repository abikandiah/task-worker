package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/abikandiah/task-worker/internal/domain"
	"github.com/google/uuid"
)

type JobWorker struct {
	*jobServiceDependencies
	jobCh  <-chan uuid.UUID
	taskCh chan<- TaskRunRequest
}

var ErrJobTimedOut = errors.New("task timed out")

func (worker *JobWorker) Run(ctx context.Context) {
	for jobID := range worker.jobCh {
		ctx := context.WithValue(ctx, domain.LKeys.JobID, jobID)

		// Get and run job
		job, err := worker.repository.GetJob(ctx, jobID)
		if err != nil {
			slog.ErrorContext(ctx, "Failed to fetch job", slog.Any("error", err))
		} else {

			err = worker.runJob(ctx, job)
			if err != nil {
				slog.ErrorContext(ctx, "Job failed", slog.Any("error", err))
				worker.updateJobState(ctx, job, domain.StateError)
				worker.repository.SaveJob(ctx, *job)
			}
		}
	}
}

func (worker *JobWorker) runJob(ctx context.Context, job *domain.Job) error {
	// Get config, revert to default if none set

	var config *domain.JobConfig
	if job.ConfigID == uuid.Nil {
		config = domain.NewDefaultJobConfig()
	} else {
		cfg, err := worker.repository.GetJobConfig(ctx, job.ConfigID)
		if err != nil {
			slog.ErrorContext(ctx, "Failed to fetch config", slog.Any("error", err))
			return fmt.Errorf("failed to fetch config %s: %w", job.ConfigID, err)
		}
		config = cfg
	}

	ctx = context.WithValue(ctx, domain.LKeys.JobName, job.Name)
	ctx = context.WithValue(ctx, domain.LKeys.ConfigID, job.ConfigID)

	ctxTimeout, cancel := context.WithTimeoutCause(ctx, (time.Duration(config.JobTimeout) * time.Second), ErrJobTimedOut)
	defer cancel()

	errCh := make(chan error, 1)
	go func() {
		errCh <- worker.executeJob(ctxTimeout, job, config)
	}()

	select {
	case err := <-errCh:
		return err

	case <-ctxTimeout.Done():
		// Error that cancelled the context
		cause := context.Cause(ctxTimeout)

		if errors.Is(cause, ErrJobTimedOut) {
			return cause
		}
		return fmt.Errorf("job interrupted by upstream cancellation: %w", cause)
	}
}

func (worker *JobWorker) executeJob(ctx context.Context, job *domain.Job, config *domain.JobConfig) error {
	job.StartDate = time.Now()
	worker.updateJobState(ctx, job, domain.StateRunning)
	worker.repository.SaveJob(ctx, *job)

	// Finalize job in defer block
	defer func() {
		job.EndDate = time.Now()
		worker.repository.SaveJob(ctx, *job)
	}()

	// Get TaskRuns
	taskRuns, err := worker.repository.GetTaskRuns(ctx, job.ID)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to fetch taskRuns", slog.Any("error", err))
		return fmt.Errorf("failed to fetch taskRuns %s: %w", job.ID, err)
	}

	// Submit tasks to TaskWorkers and wait for result
	wg := new(sync.WaitGroup)

	for _, taskRun := range taskRuns {
		if !config.EnableParallelTasks || !taskRun.Parallel {
			wg.Wait()
		}

		wg.Add(1)
		go func() {
			defer wg.Done()

			errCh := make(chan error)
			taskRequest := &TaskRunRequest{
				data:  &taskRun,
				errCh: errCh,
			}

			worker.taskCh <- *taskRequest
			err := <-errCh
			if err != nil {
				// TODO: do something with the error?
			}
		}()
	}

	return nil
}

func (worker *JobWorker) updateJobState(ctx context.Context, job *domain.Job, state domain.ExecutionState) {
	job.State = state
	slog.InfoContext(ctx, "Job "+job.State.String())
}
