package container

import (
	"github.com/abikandiah/task-worker/internal/domain/task"
	"github.com/abikandiah/task-worker/internal/platform/executor"
)

func InitTaskFactory(deps *GlobalDependencies) *task.TaskFactory {
	taskFactory := task.NewTaskFactory()
	taskFactory.Register("email_send_task", func(params any) (task.Task, error) {
		return executor.NewEmailSendTask(params, &executor.EmailDependencies{
			Logger: deps.Logger,
		})
	})

	return taskFactory
}
