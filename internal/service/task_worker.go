package service

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/abikandiah/task-worker/internal/domain"
)

type TaskWorker struct {
	*JobServiceDependencies
}

func (worker *TaskWorker) Run(ctx context.Context, taskRun *domain.TaskRun, errCh chan<- error) {
	ctx = context.WithValue(ctx, taskIDKey, taskRun.ID)

	taskRun.StartDate = time.Now()
	worker.repository.SaveTaskRun(ctx, *taskRun)

	task, err := worker.taskFactory.CreateTask(taskRun.TaskName, taskRun.Params)
	if err != nil {
		worker.logger.ErrorContext(ctx, "Failed to create task", slog.Any("error", err))
		errCh <- fmt.Errorf("failed to create task %s: %w", taskRun.TaskName, err)
	}

	res, err := task.Execute(ctx)
	if err != nil {
		worker.logger.ErrorContext(ctx, "Failed to create task", slog.Any("error", err))
		errCh <- fmt.Errorf("failed to create task %s: %w", taskRun.TaskName, err)
	}

	taskRun.Result = res
	taskRun.EndDate = time.Now()
	worker.repository.SaveTaskRun(ctx, *taskRun)

	errCh <- nil
}
