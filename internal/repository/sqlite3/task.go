package sqlite3

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/abikandiah/task-worker/internal/domain"
	"github.com/abikandiah/task-worker/internal/platform/db"
	"github.com/google/uuid"
)

type TaskRunDB struct {
	ID          uuid.UUID       `db:"id"`
	JobID       uuid.UUID       `db:"job_id"`
	Name        string          `db:"name"`
	Description sql.NullString  `db:"description"`
	TaskName    string          `db:"task_name"`
	State       string          `db:"state"`
	StartDate   db.NullTextTime `db:"start_date"`
	EndDate     db.NullTextTime `db:"end_date"`
	DetailsJSON string          `db:"details"`
}

func (taskRunDB *TaskRunDB) ToDomainTaskRun() (*domain.TaskRun, error) {
	identity := domain.Identity{
		ID: taskRunDB.ID,
		IdentitySubmission: domain.IdentitySubmission{
			Name:        taskRunDB.Name,
			Description: taskRunDB.Description.String,
		},
	}

	// Extract time.Time from TextTime
	var startDate, endDate *time.Time
	if taskRunDB.StartDate.Valid {
		startDate = &taskRunDB.StartDate.Time
	}
	if taskRunDB.EndDate.Valid {
		endDate = &taskRunDB.EndDate.Time
	}

	taskRun := &domain.TaskRun{
		Identity:  identity,
		JobID:     taskRunDB.JobID,
		TaskName:  taskRunDB.TaskName,
		State:     domain.ExecutionState(taskRunDB.State),
		StartDate: startDate,
		EndDate:   endDate,
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

func FromDomainTaskRun(taskRun domain.TaskRun) (*TaskRunDB, error) {
	// Marshal the Details struct into a JSON string
	detailsBytes, err := json.Marshal(taskRun.TaskRunDetails)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal taskRun details: %w", err)
	}

	// Set ID if new config
	taskRunID := taskRun.ID
	if taskRun.ID == uuid.Nil {
		taskRunID = uuid.New()
	}

	return &TaskRunDB{
		ID:          taskRunID,
		JobID:       taskRun.JobID,
		Name:        taskRun.Name,
		Description: sql.NullString{String: taskRun.Description, Valid: taskRun.Description != ""},
		TaskName:    taskRun.TaskName,
		State:       string(taskRun.State),
		StartDate:   db.NewNullTextTime(taskRun.StartDate),
		EndDate:     db.NewNullTextTime(taskRun.EndDate),
		DetailsJSON: string(detailsBytes),
	}, nil
}
