package task

import (
	"fmt"
)

type (
	TaskFactoryConstructor func(params any) (Task, error)
	TaskFactory            struct {
		Constructors map[string]TaskFactoryConstructor
	}
)

func NewTaskFactory() *TaskFactory {
	return &TaskFactory{
		Constructors: make(map[string]TaskFactoryConstructor),
	}
}

// Register a factoryFn for a Task
func (factory *TaskFactory) Register(name string, constructor TaskFactoryConstructor) {
	if _, exists := factory.Constructors[name]; exists {
		panic(fmt.Sprintf("Task '%s' already registered", name))
	}
	factory.Constructors[name] = constructor
}

func (factory *TaskFactory) CreateTask(name string, params any) (Task, error) {
	constructor, ok := factory.Constructors[name]
	if !ok {
		return nil, fmt.Errorf("task with name '%s' is not registered", name)
	}

	return constructor(params)
}

func (factory *TaskFactory) GetTaskNames() []string {
	keys := make([]string, 0, len(factory.Constructors))
	for k := range factory.Constructors {
		keys = append(keys, k)
	}
	return keys
}
