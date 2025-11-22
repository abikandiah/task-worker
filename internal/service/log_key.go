package service

type internalKey string

const (
	jobIDKey      internalKey = "job_id"
	jobNameKey    internalKey = "job_name"
	taskIDKey     internalKey = "task_run_id"
	taskNameKey   internalKey = "task_name"
	configIDKey   internalKey = "config_id"
	configNameKey internalKey = "config_name"
)
