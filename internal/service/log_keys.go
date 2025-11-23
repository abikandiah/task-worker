package service

import "github.com/abikandiah/task-worker/internal/domain"

type LogKeys struct {
	JobID      domain.LogKey
	JobName    domain.LogKey
	TaskID     domain.LogKey
	TaskName   domain.LogKey
	ConfigID   domain.LogKey
	ConfigName domain.LogKey
}

var LKeys = LogKeys{
	JobID:      "job_id",
	JobName:    "job_name",
	TaskID:     "task_id",
	TaskName:   "task_name",
	ConfigID:   "config_id",
	ConfigName: "config_name",
}
