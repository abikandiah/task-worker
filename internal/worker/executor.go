package worker

import "github.com/abikandiah/task-worker/internal/common"

type ExecutorProfile struct {
	common.Description
	EnableParallelTasks bool
	MaxParallel         int
}

func NewExecutorProfile() *ExecutorProfile {
	return &ExecutorProfile{
		MaxParallel: 4,
	}
}

func ExecuteTask(task Task, profile ExecutorProfile) {
}

