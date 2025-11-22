package executor

import (
	"fmt"

	"github.com/abikandiah/task-worker/internal/domain"
	"github.com/abikandiah/task-worker/internal/domain/task"
)

type TaskFactoryConstructor func(params any, deps *domain.GlobalDependencies) (task.Task, error)

type TaskFactory struct {
	task.TaskFactory
	deps *domain.GlobalDependencies
}

func InitTaskFactory(deps *domain.GlobalDependencies) *TaskFactory {
	taskFactory := &TaskFactory{
		TaskFactory: *task.NewTaskFactory(),
		deps:        deps,
	}

	taskFactory.Register("email_send_task", EmailSendConstructor)
	taskFactory.Register("chat_task", ChatConstructor)

	return taskFactory
}

// Register a factoryFn for a Task
func (factory *TaskFactory) Register(name string, constructor TaskFactoryConstructor) {
	if _, exists := factory.Constructors[name]; exists {
		panic(fmt.Sprintf("Task '%s' already registered", name))
	}

	factory.Constructors[name] = func(params any) (task.Task, error) {
		return constructor(params, factory.deps)
	}
}
