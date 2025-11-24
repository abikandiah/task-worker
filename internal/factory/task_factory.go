package factory

import (
	"fmt"
	"reflect"

	"github.com/abikandiah/task-worker/internal/domain"
	"github.com/abikandiah/task-worker/internal/task"
)

// Constructor used by the CreateTask method, to create a task without needed the deps
type TaskFactoryConstructor[P any] func(params P) (task.Task, error)

// DependencyTaskConstructor is the specific signature used when registering tasks
type DependencyTaskConstructor[P any] func(params P, deps *domain.GlobalDependencies) (task.Task, error)

type TaskFactory struct {
	Constructors map[string]any // Stores TaskFactoryConstructor[P] wrappers
	deps         *domain.GlobalDependencies
}

func NewTaskFactory(deps *domain.GlobalDependencies) *TaskFactory {
	taskFactory := &TaskFactory{
		Constructors: make(map[string]any),
		deps:         deps,
	}

	Register(taskFactory, "email_send", task.EmailSendConstructor)
	Register(taskFactory, "chat", task.ChatConstructor)

	return taskFactory
}

// Register registers a DependencyTaskConstructor. It uses generics (P) to capture
// the parameter type and internally wraps the constructor to inject dependencies.
func Register[P any](factory *TaskFactory, name string, constructor DependencyTaskConstructor[P]) {
	if _, exists := factory.Constructors[name]; exists {
		panic(fmt.Sprintf("Task '%s' already registered", name))
	}

	// Encapsulate deps
	wrapper := func(params P) (task.Task, error) {
		return constructor(params, factory.deps)
	}

	factory.Constructors[name] = wrapper
}

func CreateTask[P any](factory *TaskFactory, name string, params P) (task.Task, error) {
	constructorAny, ok := factory.Constructors[name]
	if !ok {
		return nil, fmt.Errorf("task with name '%s' is not registered", name)
	}

	// Assert the stored wrapper back to its specific generic type.
	constructorFunc, ok := constructorAny.(TaskFactoryConstructor[P])
	if !ok {
		// This robust check returns an error if the caller uses the wrong parameter type.
		expectedType := reflect.TypeOf(constructorAny).In(0)
		receivedType := reflect.TypeOf(params)
		return nil, fmt.Errorf("task '%s' expects params of type %v, but received %v", name, expectedType, receivedType)
	}

	// Execute the type-safe wrapper, which handles the DI internally.
	return constructorFunc(params)
}

// GetTaskNames returns a list of all registered task names.
func (factory *TaskFactory) GetTaskNames() []string {
	keys := make([]string, 0, len(factory.Constructors))
	for k := range factory.Constructors {
		keys = append(keys, k)
	}
	return keys
}
