package domain

type TaskFactory interface {
	CreateTask(name string, params any) (Task, error)
	CreateTaskFromJSON(name string, jsonData []byte) (Task, error)
	GetTaskNames() []string
	IsRegistered(name string) bool
	MustCreateTask(name string, params any) Task
}
