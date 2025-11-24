package mock

import (
	"context"
	"errors"
	"sync"

	"github.com/abikandiah/task-worker/internal/domain"
	"github.com/google/uuid"
)

type MockRepo struct {
	jobs     map[uuid.UUID]*domain.Job
	configs  map[uuid.UUID]*domain.JobConfig
	taskRuns map[uuid.UUID]*domain.TaskRun

	// Add a Mutex for concurrent access safety
	mu sync.RWMutex

	// Error simulation flags
	// These allow us to test failure paths like "DB is down"
	FailSaveJob      error
	FailSaveTaskRuns error
	FailGetJob       error
}

func NewMockRepo() *MockRepo {
	return &MockRepo{
		jobs:     make(map[uuid.UUID]*domain.Job),
		configs:  make(map[uuid.UUID]*domain.JobConfig),
		taskRuns: make(map[uuid.UUID]*domain.TaskRun),
	}
}

func (repo *MockRepo) GetJob(ctx context.Context, jobID uuid.UUID) (*domain.Job, error) {
	repo.mu.RLock()
	defer repo.mu.RUnlock()

	// Test method failure
	if repo.FailGetJob != nil {
		return nil, repo.FailGetJob
	}

	job, ok := repo.jobs[jobID]
	if ok {
		return job, nil
	}

	return nil, errors.New("job not found")
}

func (repo *MockRepo) SaveJob(ctx context.Context, job domain.Job) (*domain.Job, error) {
	repo.mu.Lock()
	defer repo.mu.Unlock()

	// Test method failure
	if repo.FailSaveJob != nil {
		return nil, repo.FailSaveJob
	}

	jobCopy := job
	if jobCopy.ID == uuid.Nil {
		jobCopy.ID = uuid.New()
	}

	repo.jobs[jobCopy.ID] = &jobCopy
	return &jobCopy, nil
}

func (repo *MockRepo) GetJobConfig(ctx context.Context, configID uuid.UUID) (*domain.JobConfig, error) {
	repo.mu.RLock()
	defer repo.mu.RUnlock()

	config, ok := repo.configs[configID]
	if ok {
		return config, nil
	}

	return nil, errors.New("config not found")
}

func (repo *MockRepo) SaveJobConfig(ctx context.Context, config domain.JobConfig) (*domain.JobConfig, error) {
	repo.mu.Lock()
	defer repo.mu.Unlock()

	copyConfig := config
	if copyConfig.ID == uuid.Nil {
		copyConfig.ID = uuid.New()
	}

	repo.configs[copyConfig.ID] = &copyConfig
	return &copyConfig, nil
}

func (repo *MockRepo) GetTaskRun(ctx context.Context, taskID uuid.UUID) (*domain.TaskRun, error) {
	repo.mu.RLock()
	defer repo.mu.RUnlock()

	taskRun, ok := repo.taskRuns[taskID]
	if ok {
		return taskRun, nil
	}

	return nil, errors.New("taskRun not found")
}

func (repo *MockRepo) SaveTaskRun(ctx context.Context, taskRun domain.TaskRun) (*domain.TaskRun, error) {
	repo.mu.Lock()
	defer repo.mu.Unlock()

	copyTaskRun := taskRun
	if copyTaskRun.ID == uuid.Nil {
		copyTaskRun.ID = uuid.New()
	}

	repo.taskRuns[copyTaskRun.ID] = &copyTaskRun
	return &copyTaskRun, nil
}

func (repo *MockRepo) SaveTaskRuns(ctx context.Context, taskRuns []domain.TaskRun) ([]domain.TaskRun, error) {
	repo.mu.Lock()
	defer repo.mu.Unlock()

	// Test method failure
	if repo.FailSaveTaskRuns != nil {
		return nil, repo.FailSaveTaskRuns
	}

	savedTasks := make([]domain.TaskRun, 0, len(taskRuns))

	for _, taskRun := range taskRuns {
		copyTaskRun := taskRun
		if copyTaskRun.ID == uuid.Nil {
			copyTaskRun.ID = uuid.New()
		}
		repo.taskRuns[copyTaskRun.ID] = &copyTaskRun
		savedTasks = append(savedTasks, copyTaskRun)
	}

	return savedTasks, nil
}

func (repo *MockRepo) GetTaskRuns(ctx context.Context, jobID uuid.UUID) ([]domain.TaskRun, error) {
	repo.mu.RLock()
	defer repo.mu.RUnlock()

	taskRuns := []domain.TaskRun{}
	for _, taskRun := range repo.taskRuns {
		if taskRun.JobID == jobID {
			taskRuns = append(taskRuns, *taskRun)
		}
	}

	if len(taskRuns) > 0 {
		return taskRuns, nil
	}
	return nil, errors.New("taskRuns not found")
}
