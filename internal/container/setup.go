package container

import (
	"github.com/abikandiah/task-worker/internal/core/port"
	"github.com/abikandiah/task-worker/internal/task"
)

func InitTaskFactory(deps *GlobalDependencies) *task.TaskFactory {
	taskFactory := task.NewTaskFactory()
	taskFactory.Register("email_send_task", func(params any) (port.Task, error) {
		return task.NewEmailSendTask(params, &task.EmailDependencies{
			Logger: deps.Logger,
		})
	})

	return taskFactory
}
