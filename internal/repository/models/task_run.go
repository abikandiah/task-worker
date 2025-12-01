package models

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/abikandiah/task-worker/internal/domain"
	"github.com/google/uuid"
)

type CommonTaskRunDB struct {
	ID          uuid.UUID      `db:"id"`
	JobID       uuid.UUID      `db:"job_id"`
	Name        string         `db:"name"`
	Description sql.NullString `db:"description"`
	TaskName    string         `db:"task_name"`
	State       string         `db:"state"`
	DetailsJSON string         `db:"details"`
}

func (taskRunDB *CommonTaskRunDB) ToDomainTaskRunBase() (*domain.TaskRun, error) {
	identity := domain.Identity{
		ID: taskRunDB.ID,
		IdentitySubmission: domain.IdentitySubmission{
			Name:        taskRunDB.Name,
			Description: taskRunDB.Description.String,
		},
	}

	taskRun := &domain.TaskRun{
		Identity: identity,
		JobID:    taskRunDB.JobID,
		TaskName: taskRunDB.TaskName,
		State:    domain.ExecutionState(taskRunDB.State),
	}

	// Unmarshal the DetailsJSON string back into the JobConfigDetails struct
	if taskRunDB.DetailsJSON != "" {
		err := json.Unmarshal([]byte(taskRunDB.DetailsJSON), &taskRun.TaskRunDetails)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal taskRun details JSON: %w", err)
		}
	}

	return taskRun, nil
}

func NewCommonTaskRunDB(taskRun domain.TaskRun) (CommonTaskRunDB, error) {
	// Marshal the Details struct into a JSON string
	detailsBytes, err := json.Marshal(taskRun.TaskRunDetails)
	if err != nil {
		return CommonTaskRunDB{}, fmt.Errorf("failed to marshal taskRun details: %w", err)
	}

	// Set ID if new config
	taskRunID := taskRun.ID
	if taskRun.ID == uuid.Nil {
		taskRunID = uuid.New()
	}

	return CommonTaskRunDB{
		ID:          taskRunID,
		JobID:       taskRun.JobID,
		Name:        taskRun.Name,
		Description: sql.NullString{String: taskRun.Description, Valid: taskRun.Description != ""},
		TaskName:    taskRun.TaskName,
		State:       string(taskRun.State),
		DetailsJSON: string(detailsBytes),
	}, nil
}
