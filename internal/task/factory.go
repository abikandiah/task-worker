package task

import (
	"fmt"

	"github.com/abikandiah/task-worker/internal/core/port"
)

type (
	TaskFactoryConstructor func(params any) (port.Task, error)
	TaskFactory            struct {
		constructors map[string]TaskFactoryConstructor
	}
)

func NewTaskFactory() *TaskFactory {
	return &TaskFactory{
		constructors: make(map[string]TaskFactoryConstructor),
	}
}

// Register a factoryFn for a Task
func (factory *TaskFactory) Register(name string, constructor TaskFactoryConstructor) {
	if _, exists := factory.constructors[name]; exists {
		panic(fmt.Sprintf("Task '%s' already registered", name))
	}
	factory.constructors[name] = constructor
}

func (factory *TaskFactory) CreateTask(name string, params any) (port.Task, error) {
	constructor, ok := factory.constructors[name]
	if !ok {
		return nil, fmt.Errorf("task with name '%s' is not registered", name)
	}

	return constructor(params)
}

func (factory *TaskFactory) GetTaskNames() []string {
	keys := make([]string, 0, len(factory.constructors))
	for k := range factory.constructors {
		keys = append(keys, k)
	}
	return keys
}
