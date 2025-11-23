package mock

import (
	"context"
	"errors"
	"sync"

	"github.com/abikandiah/task-worker/internal/domain"
	"github.com/google/uuid"
)

type MockRepo struct {
	data map[string]any
	// Add a Mutex for concurrent access safety
	mu sync.RWMutex
}

func (repo *MockRepo) GetJob(ctx context.Context, jobID uuid.UUID) (*domain.Job, error) {
	repo.mu.RLock()
	defer repo.mu.RUnlock()

	job, ok := repo.data[jobID.String()]
	if ok {
		return job.(*domain.Job), nil
	}

	return nil, errors.New("job not found")
}

func (repo *MockRepo) SaveJob(ctx context.Context, job domain.Job) (*domain.Job, error) {
	repo.mu.Lock()
	defer repo.mu.Unlock()

	jobCopy := job
	repo.data[job.ID.String()] = &jobCopy

	return &jobCopy, nil
}

func (repo *MockRepo) GetJobConfig(ctx context.Context, configID uuid.UUID) (*domain.JobConfig, error) {
	repo.mu.RLock()
	defer repo.mu.RUnlock()

	config, ok := repo.data[configID.String()]
	if ok {
		return config.(*domain.JobConfig), nil
	}

	return nil, errors.New("config not found")
}

func (repo *MockRepo) SaveJobConfig(ctx context.Context, config domain.JobConfig) (*domain.JobConfig, error) {
	repo.mu.Lock()
	defer repo.mu.Unlock()

	configCopy := config
	repo.data[config.ID.String()] = &configCopy

	return &configCopy, nil
}
