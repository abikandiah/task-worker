package service

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/abikandiah/task-worker/internal/core/domain"
)

type Executor struct {
	ConfigID   string
	Repository domain.ExecutorRepository
	Logger     *slog.Logger
}

func NewExecutor(configID string, repo domain.ExecutorRepository, log *slog.Logger) *Executor {
	return &Executor{
		ConfigID:   configID,
		Repository: repo,
		Logger:     log,
	}
}

type internalKey string

const (
	jobIDKey       internalKey = "job_id"
	taskIDKey      internalKey = "task_run_id"
	taskVersionKey internalKey = "task_version"
)

func (exec Executor) ExecuteJob(ctx context.Context, jobID string) error {
	ctx = context.WithValue(ctx, jobIDKey, jobID)
	exec.Logger.InfoContext(ctx, "Starting job execution")

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
	job.StartDate = time.Now()
	job.Status = "RUNNING"
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

func (exec Executor) ExecuteTaskRun(ctx context.Context, taskRun domain.TaskRun) error {
	ctx = context.WithValue(ctx, taskIDKey, taskRun.ID)

	// Get core task object with execution logic
	task, err := exec.Repository.GetTask(ctx, taskRun.TaskID, taskRun.TaskVersion)
	if err != nil {
		exec.Logger.ErrorContext(ctx, "Failed to fetch task", slog.String(string(taskIDKey), taskRun.TaskID), slog.String(string(taskVersionKey), taskRun.TaskVersion), slog.Any("error", err))
		return fmt.Errorf("failed to fetch task %s: %w", taskRun.TaskID, err)
	}

	task.Run()

	return nil
}
