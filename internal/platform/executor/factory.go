package executor

import (
	"github.com/abikandiah/task-worker/internal/domain"
	"github.com/abikandiah/task-worker/internal/domain/task"
)

func InitTaskFactory(deps *domain.GlobalDependencies) *task.TaskFactory {
	taskFactory := task.NewTaskFactory(deps)

	taskFactory.Register("email_send_task", EmailSendConstructor)
	taskFactory.Register("chat_task", ChatConstructor)

	return taskFactory
}
