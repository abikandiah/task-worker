package domain

type ExecutorConfig struct {
	IdentityVersion
	EnableParallelTasks bool
	MaxParallelTasks    int
}

func NewExecutorConfig() *ExecutorConfig {
	return &ExecutorConfig{
		MaxParallelTasks: 4,
	}
}
