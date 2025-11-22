package service

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/abikandiah/task-worker/internal/domain"
	"github.com/abikandiah/task-worker/internal/domain/task"
	"github.com/abikandiah/task-worker/internal/repository"
)

type TaskExecutor struct {
	TaskFactory task.TaskFactory
	Repository  ExecutorRepository
	Logger      *slog.Logger
}

type ExecutorRepository struct {
	repository.JobRepository
	repository.ExecutorConfigRepository
	repository.TaskRunRepository
}

func (exec *TaskExecutor) ExecuteJob(ctx context.Context, configID string, jobID string) error {
	ctx = context.WithValue(ctx, jobIDKey, jobID)

	// Get Config
	config, err := exec.Repository.GetExecutorConfig(ctx, configID)
	if err != nil {
		exec.Logger.ErrorContext(ctx, "Failed to fetch config", slog.String(string(configIDKey), configID), slog.Any("error", err))
		return fmt.Errorf("failed to fetch config %s: %w", configID, err)
	}

	// Get Job and taskRuns
	job, err := exec.Repository.GetJob(ctx, jobID)
	if err != nil {
		exec.Logger.ErrorContext(ctx, "Failed to fetch job", slog.String(string(jobIDKey), jobID), slog.Any("error", err))
		return fmt.Errorf("failed to fetch job %s: %w", jobID, err)
	}

	taskRuns, err := exec.Repository.GetTaskRuns(ctx, jobID)
	if err != nil {
		exec.Logger.ErrorContext(ctx, "Failed to fetch taskRuns for job", slog.String(string(jobIDKey), jobID), slog.Any("error", err))
		return fmt.Errorf("failed to fetch taskRuns for job %s: %w", jobID, err)
	}

	// Execute job
	exec.Logger.InfoContext(ctx, "Starting job execution")
	job.StartDate = time.Now()
	job.State = "RUNNING"

	// Start a seperate thread for periodically saving the job and taskRun - task will update job and taskRun objects

	_, err = exec.Repository.SaveJob(ctx, *job)
	if err != nil {
		exec.Logger.ErrorContext(ctx, "Failed to save job", slog.String(string(jobIDKey), jobID), slog.Any("error", err))
		return fmt.Errorf("failed to save job %s: %w", jobID, err)
	}

	// Execut job tasks
	for index, taskRun := range taskRuns {
		err := exec.ExecuteTaskRun(ctx, taskRun)
		if err != nil {
			exec.Logger.ErrorContext(ctx, "Failed to execute taskRun", slog.Int("index", index), slog.String(string(taskIDKey), taskRun.ID), slog.Any("error", err))
			return fmt.Errorf("failed to execute taskRun (%d) %s: %w", index, taskRun.ID, err)
		}
	}

	return nil
}

func (exec *TaskExecutor) ExecuteTaskRun(ctx context.Context, taskRun domain.TaskRun) error {
	ctx = context.WithValue(ctx, taskIDKey, taskRun.ID)

	// Get core task object with execution logic
	task, err := exec.TaskFactory.CreateTask(taskRun.TaskName, taskRun.Params)
	if err != nil {
		exec.Logger.ErrorContext(ctx, "Failed to fetch task", slog.String(string(taskNameKey), taskRun.TaskName), slog.Any("error", err))
		return fmt.Errorf("failed to fetch task %s: %w", taskRun.TaskName, err)
	}

	task.Execute(ctx)

	return nil
}
