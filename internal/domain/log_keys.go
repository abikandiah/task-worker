package domain

type LogKey string

type LogKeys struct {
	JobID      LogKey
	JobName    LogKey
	JobState   LogKey
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
	JobState:   "job_state",
	TaskID:     "task_id",
	TaskName:   "task_name",
	ConfigID:   "config_id",
	ConfigName: "config_name",
	RequestID:  "request_id",
	Method:     "method",
	Path:       "path",
}

var ContextLKeys = []LogKey{
	LKeys.JobID,
	LKeys.JobName,
	LKeys.JobState,
	LKeys.TaskID,
	LKeys.TaskName,
	LKeys.ConfigID,
	LKeys.ConfigName,
	LKeys.RequestID,
	LKeys.Method,
	LKeys.Path,
}
