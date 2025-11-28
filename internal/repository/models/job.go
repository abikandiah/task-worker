package models

import (
	"database/sql"

	"github.com/abikandiah/task-worker/internal/domain"
	"github.com/google/uuid"
)

type CommonJobDB struct {
	ID            uuid.UUID      `db:"id"`
	Name          string         `db:"name"`
	Description   sql.NullString `db:"description"`
	ConfigID      uuid.UUID      `db:"config_id"`
	ConfigVersion uuid.UUID      `db:"config_version"`
	State         string         `db:"state"`
	Progress      float32        `db:"progress"`
}

// GetID implements the required method for cursor pagination.
func (jdb CommonJobDB) GetID() uuid.UUID {
	return jdb.ID
}

func (jobDB *CommonJobDB) ToDomainJobBase() *domain.Job {
	identity := domain.Identity{
		ID: jobDB.ID,
		IdentitySubmission: domain.IdentitySubmission{
			Name:        jobDB.Name,
			Description: jobDB.Description.String,
		},
	}

	return &domain.Job{
		Identity:      identity,
		ConfigID:      jobDB.ConfigID,
		ConfigVersion: jobDB.ConfigVersion,
		Status: domain.Status{
			State:    domain.ExecutionState(jobDB.State),
			Progress: jobDB.Progress,
		},
	}
}

func NewCommonJobDB(job *domain.Job) CommonJobDB {
	isNew := job.ID == uuid.Nil
	jobID := job.ID
	if isNew {
		jobID = uuid.New()
	}

	return CommonJobDB{
		ID:            jobID,
		Name:          job.Name,
		Description:   sql.NullString{String: job.Description, Valid: job.Description != ""},
		ConfigID:      job.ConfigID,
		ConfigVersion: job.ConfigVersion,
		State:         string(job.State),
		Progress:      job.Progress,
	}
}
