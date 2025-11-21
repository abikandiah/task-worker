package domain

import "time"

type TaskRun struct {
	Identity
	JobID     string
	TaskName  string
	Params    any
	DependOn  []int
	Status    string
	Progress  float32
	Result    any
	StartDate time.Time
	EndDate   time.Time
}
