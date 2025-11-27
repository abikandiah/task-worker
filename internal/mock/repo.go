package mock

import (
	"context"
	"errors"
	"sort"
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
	FailGetJobs      error
}

func NewMockRepo() *MockRepo {
	return &MockRepo{
		jobs:     make(map[uuid.UUID]*domain.Job),
		configs:  make(map[uuid.UUID]*domain.JobConfig),
		taskRuns: make(map[uuid.UUID]*domain.TaskRun),
	}
}

func (repo *MockRepo) GetAllJobs(ctx context.Context, cursor *domain.CursorInput) (*domain.CursorOutput[domain.Job], error) {
	repo.mu.RLock()
	defer repo.mu.RUnlock()

	cursor.SetDefaults()

	allJobs := make([]domain.Job, 0, len(repo.jobs))
	for _, job := range repo.jobs {
		allJobs = append(allJobs, *job)
	}

	// Sort based on SortField and SortDir
	sort.Slice(allJobs, func(i, j int) bool {
		// Only handling ID for simplicity
		if cursor.SortDir == domain.SortASC {
			return allJobs[i].ID.String() < allJobs[j].ID.String()
		}
		return allJobs[i].ID.String() > allJobs[j].ID.String()
	})

	// Find the Starting Index based on the cursor
	startIndex := 0
	if cursor.HasAfterCursor() {
		// Find the index of the cursor ID
		for i, job := range allJobs {
			if job.ID == cursor.AfterID {
				// Start index is the element *after* the cursor
				startIndex = i + 1
				break
			}
		}
		// If the cursor wasn't found, treat it as the first page (startIndex = 0)
	}

	endIndex := startIndex + cursor.Limit

	// Adjust bounds
	if startIndex >= len(allJobs) {
		// No more records
		return &domain.CursorOutput[domain.Job]{Data: []domain.Job{}}, nil
	}
	if endIndex > len(allJobs) {
		endIndex = len(allJobs)
	}

	// The resulting page of jobs
	pageJobs := allJobs[startIndex:endIndex]

	// Calculate Cursors for the Output
	nextCursor := uuid.Nil
	if len(pageJobs) > 0 && endIndex < len(allJobs) {
		// The next cursor is the ID of the last element in the *current* page
		nextCursor = pageJobs[len(pageJobs)-1].ID
	}

	prevCursor := uuid.Nil
	if startIndex > 0 {
		// The previous cursor is the ID of the first element in the *current* page
		prevCursor = pageJobs[0].ID
	}

	// Return the strongly-typed generic output
	return &domain.CursorOutput[domain.Job]{
		NextCursor: &nextCursor,
		PrevCursor: &prevCursor,
		Limit:      cursor.Limit,
		Data:       pageJobs,
	}, nil
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

func (repo *MockRepo) GetAllJobConfigs(ctx context.Context, cursor *domain.CursorInput) (*domain.CursorOutput[domain.JobConfig], error) {
	repo.mu.RLock()
	defer repo.mu.RUnlock()

	configs := make([]domain.JobConfig, 0, len(repo.configs))
	for _, config := range repo.configs {
		configs = append(configs, *config)
	}

	return &domain.CursorOutput[domain.JobConfig]{
		NextCursor: nil,
		PrevCursor: nil,
		Limit:      cursor.Limit,
		Data:       configs,
	}, nil
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

func (repo *MockRepo) GetTaskRun(ctx context.Context, taskRunID uuid.UUID) (*domain.TaskRun, error) {
	repo.mu.RLock()
	defer repo.mu.RUnlock()

	taskRun, ok := repo.taskRuns[taskRunID]
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

func (repo *MockRepo) GetAllTaskRuns(ctx context.Context, cursor *domain.CursorInput) (*domain.CursorOutput[domain.TaskRun], error) {
	return nil, nil
}

func (repo *MockRepo) Close() error {
	return nil
}
