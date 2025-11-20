package worker

import "github.com/abikandiah/task-worker/internal/common"

type Task struct {
	common.Description
}

func (t Task) Run() any {
	return "Not implemented!"
}

type TaskRun struct {
	Name        string
	TaskID      string
	TaskVersion string
	Status      string
	Options     map[string]any
}

type TaskJob struct {
	common.Description
	Status string
	Tasks  []TaskRun
}
