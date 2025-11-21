package worker

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/abikandiah/task-worker/internal/common"
)

type DataRepository interface {
	GetJob(ctx context.Context, jobID string) (*Job, error)
	GetTaskRuns(ctx context.Context, taskRunID string) (*TaskRun, error)
	GetTask(ctx context.Context, taskID string) (*Task, error)
}

type ExecutorConfig struct {
	common.IdentityVersion
	EnableParallelTasks bool
	MaxParallelTasks    int
}

func NewExecutorConfig() *ExecutorConfig {
	return &ExecutorConfig{
		MaxParallelTasks: 4,
	}
}

type Executor struct {
	Config   ExecutorConfig
	DataRepo DataRepository
	Logger   *slog.Logger
}

func NewExecutor(repo DataRepository, log *slog.Logger) *Executor {
	return &Executor{
		DataRepo: repo,
		Logger:   log,
	}
}

type internalKey string

const (
	jobIDKey  internalKey = "job_id"
	taskIDKey internalKey = "task_id"
)

func (exec Executor) ExecuteJob(ctx context.Context, jobID string) error {
	ctx = context.WithValue(ctx, jobIDKey, jobID)
	exec.Logger.InfoContext(ctx, "Starting job execution")

	// Get Job
	job, err := exec.DataRepo.GetJob(ctx, jobID)
	if err != nil {
		exec.Logger.ErrorContext(ctx, "Failed to fetch job", err)
		return err
	}

	taskRuns, err := exec.DataRepo.GetTaskRuns(ctx, jobID)
	if err != nil {
		exec.Logger.ErrorContext(ctx, fmt.Sprintf("Failed to fetch taskRuns for job %s", jobID), err)
		return err
	}

	// Get tasks from DB via jobId and TaskRun table

	// Execute task, determine collection pool behaviours, etc.

	return nil
}

func (taskRun TaskRun) ExecuteTask(ctx context.Context) error {
	// Get core task object with execution logic

	return nil
}
