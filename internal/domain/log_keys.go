package domain

type LogKey string

type LogKeys struct {
	JobID      LogKey
	JobName    LogKey
	TaskID     LogKey
	TaskName   LogKey
	ConfigID   LogKey
	ConfigName LogKey
	RequestID  LogKey
	Method     LogKey
	Path       LogKey
}

var LKeys = LogKeys{
	JobID:      "job_id",
	JobName:    "job_name",
	TaskID:     "task_id",
	TaskName:   "task_name",
	ConfigID:   "config_id",
	ConfigName: "config_name",
	RequestID:  "request_id",
	Method:     "method",
	Path:       "path",
}
