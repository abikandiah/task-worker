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

type JobWorker struct {
	*JobServiceDependencies
	workerCh <-chan uuid.UUID
}

func (worker *JobWorker) Run(ctx context.Context) {
	for jobID := range worker.workerCh {
		ctx := context.WithValue(ctx, jobIDKey, jobID)

		// Get and run job
		job, err := worker.repository.GetJob(ctx, jobID)
		if err != nil {
			worker.logger.ErrorContext(ctx, "Failed to fetch job", slog.Any("error", err))
		} else {

			err = worker.runJob(ctx, job)
			if err != nil {
				worker.logger.ErrorContext(ctx, "Job failed", slog.Any("error", err))
				worker.updateJobState(ctx, job, domain.StateError)
				worker.repository.SaveJob(ctx, *job)
			}
		}
	}
}

func (worker *JobWorker) runJob(ctx context.Context, job *domain.Job) error {
	ctx = context.WithValue(ctx, jobNameKey, job.Name)
	ctx = context.WithValue(ctx, configIDKey, job.ConfigID)

	// Get config
	config, err := worker.repository.GetJobConfig(ctx, job.ConfigID)
	if err != nil {
		worker.logger.ErrorContext(ctx, "Failed to fetch config", slog.Any("error", err))
		return fmt.Errorf("failed to fetch config %s: %w", job.ConfigID, err)
	}

	job.StartDate = time.Now()
	worker.updateJobState(ctx, job, domain.StateRunning)
	worker.repository.SaveJob(ctx, *job)

	ctxTimeout, cancel := context.WithTimeout(ctx, (time.Duration(config.JobTimeout) * time.Second))
	defer cancel() // Always schedule cleanup to release resources

	errCh := make(chan error)
	go worker.runJobTasks(ctxTimeout, job, config, errCh)

	select {
	case err := <-errCh:
		if err != nil {
			return err
		}

		// Job finished
		worker.updateJobState(ctx, job, domain.StateFinished)

	case <-ctxTimeout.Done():
		// Context cancelled / timed out
		worker.updateJobState(ctx, job, domain.StateStopped)
	}

	job.EndDate = time.Now()
	worker.repository.SaveJob(ctx, *job)

	return nil
}

func (worker *JobWorker) runJobTasks(ctx context.Context, job *domain.Job, config *domain.JobConfig, errCh chan<- error) {
	<-time.After(10 * time.Second)

	// Get TaskRuns
	taskRuns, err := worker.repository.GetTaskRuns(ctx, job.ConfigID)
	if err != nil {
		worker.logger.ErrorContext(ctx, "Failed to fetch taskRuns", slog.Any("error", err))
		errCh <- fmt.Errorf("failed to fetch taskRuns %s: %w", job.ID, err)
	}

	// Execute tasks and update result
	wg := new(sync.WaitGroup)

	for _, taskRun := range taskRuns {
		if !config.EnableParallelTasks || !taskRun.Parallel {
			wg.Wait()
		}

		wg.Add(1)
		go func() {
			defer wg.Done()
			ctxTimeout, cancel := context.WithTimeout(ctx, (time.Duration(config.TaskTimeout) * time.Second))
			defer cancel()

			errCh := make(chan error)

			go func() {
				err := worker.runTask(ctx, &taskRun)
				if err != nil {
					worker.logger.ErrorContext(ctx, "Task failed", slog.Any("error", err))
					worker.updateTaskState(ctx, &taskRun, domain.StateError)
					worker.repository.SaveTaskRun(ctx, taskRun)
					errCh <- err
				}

				errCh <- nil
			}()

			select {
			case err := <-errCh:
				if err == nil {
					worker.updateTaskState(ctx, &taskRun, domain.StateFinished)
				}
			case <-ctxTimeout.Done():
				worker.updateTaskState(ctx, &taskRun, domain.StateStopped)
			}

			taskRun.EndDate = time.Now()
			worker.repository.SaveTaskRun(ctx, taskRun)
		}()
	}

	errCh <- nil
}

func (worker *JobWorker) runTask(ctx context.Context, taskRun *domain.TaskRun) error {
	ctx = context.WithValue(ctx, taskIDKey, taskRun.ID)

	taskRun.StartDate = time.Now()
	worker.repository.SaveTaskRun(ctx, *taskRun)

	task, err := worker.taskFactory.CreateTask(taskRun.TaskName, taskRun.Params)
	if err != nil {
		worker.logger.ErrorContext(ctx, "Failed to create task", slog.Any("error", err))
		return fmt.Errorf("failed to create task %s: %w", taskRun.TaskName, err)
	}

	res, err := task.Execute(ctx)
	if err != nil {
		worker.logger.ErrorContext(ctx, "Failed to create task", slog.Any("error", err))
		return fmt.Errorf("failed to create task %s: %w", taskRun.TaskName, err)
	}

	taskRun.Result = res
	taskRun.EndDate = time.Now()
	worker.repository.SaveTaskRun(ctx, *taskRun)

	return nil
}

func (worker *JobWorker) updateJobState(ctx context.Context, job *domain.Job, state domain.ExecutionState) {
	job.State = state
	worker.logger.InfoContext(ctx, "Job "+job.State.String())
}

func (worker *JobWorker) updateTaskState(ctx context.Context, taskRun *domain.TaskRun, state domain.ExecutionState) {
	taskRun.State = state
	worker.logger.InfoContext(ctx, "TaskRun "+taskRun.State.String())
}
