package factory

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"reflect"
	"sync"

	"github.com/abikandiah/task-worker/internal/task"
)

type TaskConstructor[P any, D any] func(params P, deps D) (task.Task, error)

// TaskFactory manages task registration, dependency injection, and task creation.
type TaskFactory struct {
	mu           sync.RWMutex
	constructors map[string]any          // Stores constructor wrappers
	paramTypes   map[string]reflect.Type // Expected parameter types
	depTypes     map[string]reflect.Type // Expected dependency types
	dependencies map[reflect.Type]any    // Registry of available dependencies by type
}

func NewTaskFactory() *TaskFactory {
	return &TaskFactory{
		constructors: make(map[string]any),
		paramTypes:   make(map[string]reflect.Type),
		depTypes:     make(map[string]reflect.Type),
		dependencies: make(map[reflect.Type]any),
	}
}

// RegisterDependency registers a dependency instance that can be injected into tasks.
// The dependency type is inferred from the value.
func RegisterDependency[D any](factory *TaskFactory, dep D) {
	factory.mu.Lock()
	defer factory.mu.Unlock()

	depType := reflect.TypeOf(dep)
	if depType == nil {
		slog.Error("cannot register nil dependency")
		return
	}

	if _, exists := factory.dependencies[depType]; exists {
		panic(fmt.Sprintf("dependency of type %v is already registered", depType))
	}

	factory.dependencies[depType] = dep
}

// Register registers a task constructor with automatic dependency injection.
// P is the params type, D is the dependencies type.
// The factory will automatically inject D when CreateTask is called.
func Register[P any, D any](factory *TaskFactory, name string, constructor TaskConstructor[P, D]) {
	if name == "" {
		panic("task name cannot be empty")
	}
	if constructor == nil {
		panic(fmt.Sprintf("constructor for task '%s' cannot be nil", name))
	}

	factory.mu.Lock()
	defer factory.mu.Unlock()

	if _, exists := factory.constructors[name]; exists {
		panic(fmt.Sprintf("task '%s' is already registered", name))
	}

	// Store parameter and dependency types
	var zeroP P
	var zeroD D
	paramType := reflect.TypeOf(zeroP)
	depType := reflect.TypeOf(zeroD)

	factory.paramTypes[name] = paramType
	factory.depTypes[name] = depType

	// Create a wrapper that handles dependency resolution and injection
	wrapper := func(params P) (task.Task, error) {
		// Resolve dependencies
		deps, err := factory.resolveDependencies(name, depType)
		if err != nil {
			return nil, err
		}

		// Type assert the resolved dependencies to D
		typedDeps, ok := deps.(D)
		if !ok {
			return nil, fmt.Errorf("internal error: dependency resolution failed for task '%s'", name)
		}

		// Call the constructor with params and injected dependencies
		return constructor(params, typedDeps)
	}

	factory.constructors[name] = wrapper
}

// resolveDependencies resolves dependencies for a task based on its dependency type.
// It supports structs with multiple dependencies and single dependency types.
func (f *TaskFactory) resolveDependencies(taskName string, depType reflect.Type) (any, error) {
	if depType == nil {
		return nil, nil
	}

	// If dependency type is a struct, resolve each field
	if depType.Kind() == reflect.Struct {
		return f.resolveStructDependencies(taskName, depType)
	}

	// If dependency type is a pointer to a struct, resolve fields and return pointer
	if depType.Kind() == reflect.Ptr && depType.Elem().Kind() == reflect.Struct {
		structDeps, err := f.resolveStructDependencies(taskName, depType.Elem())
		if err != nil {
			return nil, err
		}
		// Create a pointer to the resolved struct
		ptrVal := reflect.New(depType.Elem())
		ptrVal.Elem().Set(reflect.ValueOf(structDeps))
		return ptrVal.Interface(), nil
	}

	// Single dependency type - look it up directly
	dep, ok := f.dependencies[depType]
	if !ok {
		return nil, fmt.Errorf("task '%s' requires dependency of type %v, but it's not registered", taskName, depType)
	}
	return dep, nil
}

// resolveStructDependencies resolves all fields in a dependency struct.
func (f *TaskFactory) resolveStructDependencies(taskName string, structType reflect.Type) (any, error) {
	depsValue := reflect.New(structType).Elem()

	for i := 0; i < structType.NumField(); i++ {
		field := structType.Field(i)
		fieldType := field.Type

		// Look up the dependency for this field type
		dep, ok := f.dependencies[fieldType]
		if !ok {
			return nil, fmt.Errorf(
				"task '%s' requires dependency field '%s' of type %v, but it's not registered",
				taskName,
				field.Name,
				fieldType,
			)
		}

		// Set the field value
		fieldValue := depsValue.Field(i)
		if fieldValue.CanSet() {
			fieldValue.Set(reflect.ValueOf(dep))
		}
	}

	return depsValue.Interface(), nil
}

// CreateTask creates a task instance by name with the provided parameters.
// Dependencies are automatically injected based on the task's registered dependency type.
// Params can be:
// - A concrete struct/pointer that matches the expected type
// - Raw JSON bytes that will be deserialized to the expected type
// - nil for tasks that don't require parameters
func (f *TaskFactory) CreateTask(name string, params any) (task.Task, error) {
	if name == "" {
		return nil, fmt.Errorf("task name cannot be empty")
	}

	f.mu.RLock()
	constructorAny, ok := f.constructors[name]
	expectedType, typeExists := f.paramTypes[name]
	f.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("task '%s' is not registered", name)
	}

	// If params is raw JSON bytes, deserialize to the expected type
	if jsonBytes, ok := params.([]byte); ok && typeExists {
		deserializedParams, err := f.deserializeParams(name, expectedType, jsonBytes)
		if err != nil {
			return nil, err
		}
		params = deserializedParams
	}

	// If params is a json.RawMessage, treat it as JSON bytes
	if rawMsg, ok := params.(json.RawMessage); ok && typeExists {
		deserializedParams, err := f.deserializeParams(name, expectedType, []byte(rawMsg))
		if err != nil {
			return nil, err
		}
		params = deserializedParams
	}

	return f.invokeConstructor(name, constructorAny, params)
}

// CreateTaskFromJSON creates a task from raw JSON data.
func (f *TaskFactory) CreateTaskFromJSON(name string, jsonData []byte) (task.Task, error) {
	return f.CreateTask(name, jsonData)
}

// deserializeParams deserializes JSON bytes to the expected parameter type.
func (f *TaskFactory) deserializeParams(taskName string, expectedType reflect.Type, jsonData []byte) (any, error) {
	// Handle nil/empty JSON
	if len(jsonData) == 0 || string(jsonData) == "null" {
		if expectedType == nil ||
			expectedType.Kind() == reflect.Ptr ||
			expectedType.Kind() == reflect.Interface ||
			expectedType.Kind() == reflect.Map ||
			expectedType.Kind() == reflect.Slice {
			return nil, nil
		}
		return nil, fmt.Errorf("task '%s' expects non-nil params", taskName)
	}

	// Create a new instance of the expected type
	var paramsPtr reflect.Value
	if expectedType.Kind() == reflect.Ptr {
		paramsPtr = reflect.New(expectedType.Elem())
	} else {
		paramsPtr = reflect.New(expectedType)
	}

	// Unmarshal JSON into the new instance
	if err := json.Unmarshal(jsonData, paramsPtr.Interface()); err != nil {
		return nil, fmt.Errorf("failed to deserialize params for task '%s': %w", taskName, err)
	}

	// Return the appropriate value (pointer or value)
	if expectedType.Kind() == reflect.Ptr {
		return paramsPtr.Interface(), nil
	}
	return paramsPtr.Elem().Interface(), nil
}

// invokeConstructor uses reflection to call a constructor with dynamic params.
func (f *TaskFactory) invokeConstructor(taskName string, constructorAny any, params any) (task.Task, error) {
	constructorVal := reflect.ValueOf(constructorAny)
	constructorType := constructorVal.Type()

	if constructorType.Kind() != reflect.Func {
		return nil, fmt.Errorf("internal error: constructor for task '%s' is not a function", taskName)
	}

	if constructorType.NumIn() != 1 || constructorType.NumOut() != 2 {
		return nil, fmt.Errorf("internal error: constructor for task '%s' has invalid signature", taskName)
	}

	expectedParamType := constructorType.In(0)
	paramsVal := reflect.ValueOf(params)

	// Handle nil params - create zero value for any type
	if params == nil {
		paramsVal = reflect.Zero(expectedParamType)
	} else {
		// Verify type compatibility
		actualParamType := paramsVal.Type()
		if !actualParamType.AssignableTo(expectedParamType) {
			return nil, fmt.Errorf(
				"task '%s' expects params of type %v, but received %v",
				taskName,
				expectedParamType,
				actualParamType,
			)
		}
	}

	// Call the constructor (dependencies already injected in wrapper)
	results := constructorVal.Call([]reflect.Value{paramsVal})

	// Extract results: (task.Task, error)
	taskVal := results[0].Interface()
	errVal := results[1].Interface()

	var resultTask task.Task
	var resultErr error

	if taskVal != nil {
		resultTask = taskVal.(task.Task)
	}
	if errVal != nil {
		resultErr = errVal.(error)
	}

	return resultTask, resultErr
}

// IsRegistered checks if a task with the given name is registered.
func (f *TaskFactory) IsRegistered(name string) bool {
	f.mu.RLock()
	defer f.mu.RUnlock()
	_, exists := f.constructors[name]
	return exists
}

// GetTaskNames returns a list of all registered task names.
func (f *TaskFactory) GetTaskNames() []string {
	f.mu.RLock()
	defer f.mu.RUnlock()

	names := make([]string, 0, len(f.constructors))
	for name := range f.constructors {
		names = append(names, name)
	}
	return names
}

// Count returns the number of registered tasks.
func (f *TaskFactory) Count() int {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return len(f.constructors)
}

// MustCreateTask is like CreateTask but panics on error.
func (f *TaskFactory) MustCreateTask(name string, params any) task.Task {
	t, err := f.CreateTask(name, params)
	if err != nil {
		panic(fmt.Sprintf("failed to create task '%s': %v", name, err))
	}
	return t
}
