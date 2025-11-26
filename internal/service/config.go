package service

import (
	"fmt"

	"github.com/spf13/viper"
)

type Config struct {
	JobBufferCapacity int `mapstructure:"job_buffer_capacity"`
	JobWorkerCount    int `mapstructure:"job_worker_count"`
	TaskWorkerCount   int `mapstructure:"task_worker_count"`
}

func SetConfigDefaults(v *viper.Viper) {
	v.SetDefault("worker.job_buffer_capacity", 128)
	v.SetDefault("worker.job_worker_count", 2)
	v.SetDefault("worker.task_worker_count", 4)
}

func BindEnvironmentVariables(v *viper.Viper) {
	v.BindEnv("worker.job_buffer_capacity", "JOB_BUFFER_CAPACITY")
	v.BindEnv("worker.job_worker_count", "JOB_WORKER_COUNT")
	v.BindEnv("worker.task_worker_count", "TASK_WORKER_COUNT")
}

func (config *Config) Validate() error {
	if config.JobBufferCapacity < 0 {
		return fmt.Errorf("job buffer capacity cannot be negative")
	}
	if config.JobWorkerCount < 1 {
		return fmt.Errorf("job worker count must be at least 1")
	}
	if config.TaskWorkerCount < 1 {
		return fmt.Errorf("task worker count must be at least 1")
	}
	return nil
}
