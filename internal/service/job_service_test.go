package service

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/abikandiah/task-worker/config"
	"github.com/abikandiah/task-worker/internal/domain"
	"github.com/abikandiah/task-worker/internal/factory"
	"github.com/abikandiah/task-worker/internal/mock"
	"github.com/abikandiah/task-worker/internal/platform/logging"
	"github.com/abikandiah/task-worker/internal/task"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestJobService(mockRepo *mock.MockRepo) (*JobService, *domain.GlobalDependencies) {
	config := config.Config{
		Worker: config.WorkerConfig{
			JobWorkerCount:    2,
			TaskWorkerCount:   2,
			JobBufferCapacity: 10,
		},
		Logger: config.LoggerConfig{
			Environment: "dev",
		},
	}

	globalDeps := &domain.GlobalDependencies{
		Logger:     logging.SetupLogger(config.Logger),
		Config:     &config,
		Repository: mockRepo,
	}

	service := NewJobService(JobServiceParams{
		Config:      globalDeps.Config.Worker,
		Repository:  globalDeps.Repository,
		TaskFactory: factory.NewTaskFactory(globalDeps),
		Logger:      globalDeps.Logger,
	})

	return service, globalDeps
}

func TestNewJobService(t *testing.T) {
	ctx := context.Background()
	mockRepo := mock.NewMockRepo()
	service, deps := setupTestJobService(mockRepo)
	defer service.Close(context.Background())

	assert.NotNil(t, deps)
	assert.NotNil(t, service)
	assert.NotNil(t, service.jobCh)
	assert.NotNil(t, service.taskCh)
	assert.NotNil(t, service.wg)

	service.StartWorkers(ctx)
	assert.NotNil(t, service.cancel)
}

func TestJobService_SubmitJob_Success(t *testing.T) {
	ctx := context.Background()
	mockRepo := mock.NewMockRepo()
	service, _ := setupTestJobService(mockRepo)
	defer service.Close(ctx)

	submission := &domain.JobSubmission{
		IdentitySubmission: domain.IdentitySubmission{
			Name:        "Test Job",
			Description: "Test Description",
		},
		TaskRuns: []domain.TaskRun{
			{
				TaskName: "chat",
			},
			{
				TaskName: "email_send",
			},
			{
				TaskName: "duration",
				Params: &task.DurationParams{
					Length: 30,
				},
			},
		},
	}

	job, err := service.SubmitJob(ctx, submission)

	require.NoError(t, err)
	assert.NotNil(t, job)
	assert.NotEqual(t, uuid.Nil, job.ID)
	assert.Equal(t, "Test Job", job.Name)
	assert.Equal(t, "Test Description", job.Description)
	assert.False(t, job.SubmitDate.IsZero())

	// Verify job was sent to jobCh
	select {
	case jobID := <-service.jobCh:
		assert.Equal(t, job.ID, jobID)
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Expected job ID to be sent to jobCh")
	}

	// Verify task runs were saved with correct JobID
	taskRuns, err := mockRepo.GetTaskRuns(ctx, job.ID)
	require.NoError(t, err)
	assert.Len(t, taskRuns, 3)
	for _, tr := range taskRuns {
		assert.Equal(t, job.ID, tr.JobID)
	}
}

func TestJobService_SubmitJob_EmptySubmission(t *testing.T) {
	ctx := context.Background()
	mockRepo := mock.NewMockRepo()
	service, _ := setupTestJobService(mockRepo)
	defer service.Close(ctx)

	submission := &domain.JobSubmission{
		IdentitySubmission: domain.IdentitySubmission{
			Name: "Empty Job",
		},
		TaskRuns: []domain.TaskRun{},
	}

	job, err := service.SubmitJob(ctx, submission)

	require.NoError(t, err)
	assert.NotNil(t, job)
	assert.Equal(t, "Empty Job", job.Name)
}

func TestJobService_SubmitJob_SaveJobFailure(t *testing.T) {
	ctx := context.Background()

	mockRepo := mock.NewMockRepo()
	mockRepo.FailSaveJob = errors.New("database connection failed")

	service, _ := setupTestJobService(mockRepo)
	defer service.Close(ctx)

	submission := &domain.JobSubmission{
		IdentitySubmission: domain.IdentitySubmission{
			Name: "Test Job",
		},
		TaskRuns: []domain.TaskRun{
			{TaskName: "task1"},
		},
	}

	job, err := service.SubmitJob(ctx, submission)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to save job")
	assert.Nil(t, job)
}

func TestJobService_SubmitJob_SaveTaskRunsFailure(t *testing.T) {
	ctx := context.Background()

	mockRepo := mock.NewMockRepo()
	mockRepo.FailSaveTaskRuns = errors.New("task runs save failed")

	service, _ := setupTestJobService(mockRepo)
	defer service.Close(ctx)

	submission := &domain.JobSubmission{
		IdentitySubmission: domain.IdentitySubmission{
			Name: "Test Job",
		},
		TaskRuns: []domain.TaskRun{
			{TaskName: "task1"},
		},
	}

	job, err := service.SubmitJob(ctx, submission)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to save job taskRuns")
	assert.NotNil(t, job) // Job was saved, but taskRuns failed
}

func TestJobService_GetJob_Success(t *testing.T) {
	ctx := context.Background()
	mockRepo := mock.NewMockRepo()
	service, _ := setupTestJobService(mockRepo)
	defer service.Close(ctx)

	// First, submit a job
	submission := &domain.JobSubmission{
		IdentitySubmission: domain.IdentitySubmission{
			Name:        "Test Job",
			Description: "Test Description",
		},
		TaskRuns: []domain.TaskRun{},
	}

	submittedJob, err := service.SubmitJob(ctx, submission)
	require.NoError(t, err)

	// Drain the jobCh
	<-service.jobCh

	// Now retrieve it
	retrievedJob, err := service.GetJob(ctx, submittedJob.ID)

	require.NoError(t, err)
	assert.NotNil(t, retrievedJob)
	assert.Equal(t, submittedJob.ID, retrievedJob.ID)
	assert.Equal(t, "Test Job", retrievedJob.Name)
	assert.Equal(t, "Test Description", retrievedJob.Description)
}

func TestJobService_GetJob_NotFound(t *testing.T) {
	ctx := context.Background()
	mockRepo := mock.NewMockRepo()
	service, _ := setupTestJobService(mockRepo)
	defer service.Close(ctx)

	nonExistentID := uuid.New()

	job, err := service.GetJob(ctx, nonExistentID)

	require.Error(t, err)
	assert.Nil(t, job)
	assert.Contains(t, err.Error(), "job not found")
}

func TestJobService_GetJob_RepositoryError(t *testing.T) {
	ctx := context.Background()

	mockRepo := mock.NewMockRepo()
	mockRepo.FailGetJob = errors.New("database error")

	service, _ := setupTestJobService(mockRepo)
	defer service.Close(ctx)

	jobID := uuid.New()

	job, err := service.GetJob(ctx, jobID)

	require.Error(t, err)
	assert.Nil(t, job)
	assert.Contains(t, err.Error(), "database error")
}

func TestJobService_Close(t *testing.T) {
	ctx := context.Background()
	mockRepo := mock.NewMockRepo()
	service, _ := setupTestJobService(mockRepo)
	defer service.Close(ctx)

	// Submit a job to ensure workers are running
	submission := &domain.JobSubmission{
		IdentitySubmission: domain.IdentitySubmission{
			Name: "Test Job",
		},
		TaskRuns: []domain.TaskRun{},
	}

	_, err := service.SubmitJob(ctx, submission)
	require.NoError(t, err)

	// Close should not panic and should complete
	assert.NotPanics(t, func() {
		service.Close(ctx)
	})

	// Verify channels are closed
	_, ok := <-service.jobCh
	assert.False(t, ok, "jobCh should be closed")

	_, ok = <-service.taskCh
	assert.False(t, ok, "taskCh should be closed")
}

func TestJobService_MultipleJobs(t *testing.T) {
	ctx := context.Background()
	mockRepo := mock.NewMockRepo()
	service, _ := setupTestJobService(mockRepo)
	defer service.Close(ctx)

	numJobs := 5

	jobIDs := make([]uuid.UUID, 0, numJobs)

	// Submit multiple jobs
	for i := 0; i < numJobs; i++ {
		submission := &domain.JobSubmission{
			IdentitySubmission: domain.IdentitySubmission{
				Name: fmt.Sprintf("Job %d", i),
			},
			TaskRuns: []domain.TaskRun{
				{TaskName: fmt.Sprintf("task-%d", i)},
			},
		}

		job, err := service.SubmitJob(ctx, submission)
		require.NoError(t, err)
		jobIDs = append(jobIDs, job.ID)
	}

	// Verify all jobs can be retrieved
	for i, jobID := range jobIDs {
		job, err := service.GetJob(ctx, jobID)
		require.NoError(t, err)
		assert.Equal(t, fmt.Sprintf("Job %d", i), job.Name)
	}

	// Drain the jobCh
	for i := 0; i < numJobs; i++ {
		select {
		case <-service.jobCh:
		case <-time.After(100 * time.Millisecond):
			t.Fatal("Expected job ID in channel")
		}
	}
}

func TestJobService_ConcurrentSubmissions(t *testing.T) {
	ctx := context.Background()
	mockRepo := mock.NewMockRepo()
	service, _ := setupTestJobService(mockRepo)
	defer service.Close(ctx)

	numGoroutines := 10

	results := make(chan error, numGoroutines)

	// Submit jobs concurrently
	for i := 0; i < numGoroutines; i++ {
		go func(index int) {
			submission := &domain.JobSubmission{
				IdentitySubmission: domain.IdentitySubmission{
					Name: fmt.Sprintf("Concurrent Job %d", index),
				},
				TaskRuns: []domain.TaskRun{
					{TaskName: fmt.Sprintf("task-%d", index)},
				},
			}

			_, err := service.SubmitJob(ctx, submission)
			results <- err
		}(i)
	}

	// Verify all submissions succeeded
	for i := 0; i < numGoroutines; i++ {
		err := <-results
		assert.NoError(t, err)
	}
}

func TestJobService_SubmitJob_TaskRunsPopulatedWithJobID(t *testing.T) {
	ctx := context.Background()
	mockRepo := mock.NewMockRepo()
	service, _ := setupTestJobService(mockRepo)
	defer service.Close(ctx)

	submission := &domain.JobSubmission{
		IdentitySubmission: domain.IdentitySubmission{
			Name: "Test Job",
		},
		TaskRuns: []domain.TaskRun{
			{TaskName: "task1"},
			{TaskName: "task2"},
			{TaskName: "task3"},
		},
	}

	job, err := service.SubmitJob(ctx, submission)
	require.NoError(t, err)

	// Verify all task runs have the correct JobID
	taskRuns, err := mockRepo.GetTaskRuns(ctx, job.ID)
	require.NoError(t, err)

	assert.Len(t, taskRuns, 3)
	for _, tr := range taskRuns {
		assert.Equal(t, job.ID, tr.JobID, "TaskRun should have the correct JobID")
		assert.NotEqual(t, uuid.Nil, tr.ID, "TaskRun should have been assigned an ID")
	}
}
