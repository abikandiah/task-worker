package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/abikandiah/task-worker/internal/domain"
)

type TaskRunRequest struct {
	data    *domain.TaskRun
	timeout int
	errCh   chan error
}

type TaskWorker struct {
	*JobServiceDependencies
	taskCh <-chan TaskRunRequest
}

var ErrTaskTimedOut = errors.New("task timed out")

func (worker *TaskWorker) Run(ctx context.Context) {
	for request := range worker.taskCh {

		ctx := context.WithValue(ctx, LKeys.TaskID, request.data.ID)
		request.errCh <- worker.runTask(ctx, request.data, request.timeout)
	}
}

func (worker *TaskWorker) runTask(ctx context.Context, taskRun *domain.TaskRun, timeout int) error {
	// Execute task with timeout
	ctxTimeout, cancel := context.WithTimeoutCause(ctx, (time.Duration(timeout) * time.Second), ErrTaskTimedOut)
	defer cancel()

	errCh := make(chan error, 1)
	go func() {
		errCh <- worker.ExecuteTask(ctxTimeout, taskRun)
	}()

	select {
	case err := <-errCh:
		if err != nil {
			worker.logger.ErrorContext(ctx, "Task failed", slog.Any("error", err))
			worker.updateTaskState(ctx, taskRun, domain.StateError)
			worker.repository.SaveTaskRun(ctx, *taskRun)
		}
		return err

	case <-ctxTimeout.Done():
		// Error that cancelled the context
		cause := context.Cause(ctxTimeout)

		if errors.Is(cause, ErrTaskTimedOut) {
			return cause
		}
		return fmt.Errorf("task interrupted by upstream cancellation: %w", cause)
	}
}

func (worker *TaskWorker) ExecuteTask(ctx context.Context, taskRun *domain.TaskRun) error {
	taskRun.StartDate = time.Now()
	worker.updateTaskState(ctx, taskRun, domain.StateRunning)
	worker.repository.SaveTaskRun(ctx, *taskRun)

	// Finalize task in defer block
	defer func() {
		taskRun.EndDate = time.Now()
		worker.repository.SaveTaskRun(ctx, *taskRun)
	}()

	task, err := worker.taskFactory.CreateTask(taskRun.TaskName, taskRun.Params)
	if err != nil {
		worker.logger.ErrorContext(ctx, "Failed to create task", slog.Any("error", err))
		return fmt.Errorf("failed to create task %s: %w", taskRun.TaskName, err)
	}

	res, err := task.Execute(ctx)
	if err != nil {
		worker.logger.ErrorContext(ctx, "Task failed", slog.Any("error", err))
		return fmt.Errorf("task failed %s: %w", taskRun.TaskName, err)
	}

	taskRun.Result = res

	return nil
}

func (worker *TaskWorker) updateTaskState(ctx context.Context, taskRun *domain.TaskRun, state domain.ExecutionState) {
	taskRun.State = state
	worker.logger.InfoContext(ctx, "TaskRun "+taskRun.State.String())
}
