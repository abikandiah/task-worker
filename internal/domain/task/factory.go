package task

import (
	"fmt"

	"github.com/abikandiah/task-worker/internal/domain"
)

type (
	TaskFactoryConstructor func(params any, deps *domain.GlobalDependencies) (Task, error)
	TaskFactory            struct {
		constructors map[string]TaskFactoryConstructor
		deps         *domain.GlobalDependencies
	}
)

func NewTaskFactory(deps *domain.GlobalDependencies) *TaskFactory {
	return &TaskFactory{
		constructors: make(map[string]TaskFactoryConstructor),
		deps:         deps,
	}
}

// Register a factoryFn for a Task
func (factory *TaskFactory) Register(name string, constructor TaskFactoryConstructor) {
	if _, exists := factory.constructors[name]; exists {
		panic(fmt.Sprintf("Task '%s' already registered", name))
	}
	factory.constructors[name] = constructor
}

func (factory *TaskFactory) CreateTask(name string, params any) (Task, error) {
	constructor, ok := factory.constructors[name]
	if !ok {
		return nil, fmt.Errorf("task with name '%s' is not registered", name)
	}

	return constructor(params, factory.deps)
}

func (factory *TaskFactory) GetTaskNames() []string {
	keys := make([]string, 0, len(factory.constructors))
	for k := range factory.constructors {
		keys = append(keys, k)
	}
	return keys
}
