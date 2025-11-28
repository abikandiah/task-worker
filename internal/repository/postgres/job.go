package postgres

import (
	"database/sql"
	"time"

	"github.com/abikandiah/task-worker/internal/domain"
	"github.com/google/uuid"
)

type JobDB struct {
	ID            uuid.UUID      `db:"id"`
	Name          string         `db:"name"`
	Description   sql.NullString `db:"description"`
	ConfigID      uuid.UUID      `db:"config_id"`
	ConfigVersion uuid.UUID      `db:"config_version"`
	State         string         `db:"state"`
	Progress      float32        `db:"progress"`
	SubmitDate    time.Time      `db:"submit_date"`
	StartDate     sql.NullTime   `db:"start_date"`
	EndDate       sql.NullTime   `db:"end_date"`
}

// GetID implements the required method for cursor pagination.
func (jdb JobDB) GetID() uuid.UUID {
	return jdb.ID
}

func (jobDB *JobDB) ToDomainJob() *domain.Job {
	identity := domain.Identity{
		ID: jobDB.ID,
		IdentitySubmission: domain.IdentitySubmission{
			Name:        jobDB.Name,
			Description: jobDB.Description.String,
		},
	}

	var startDate, endDate *time.Time
	// Standard conversion from sql.NullTime to *time.Time
	if jobDB.StartDate.Valid {
		startDate = &jobDB.StartDate.Time
	}
	if jobDB.EndDate.Valid {
		endDate = &jobDB.EndDate.Time
	}

	return &domain.Job{
		Identity:      identity,
		ConfigID:      jobDB.ConfigID,
		ConfigVersion: jobDB.ConfigVersion,
		Status: domain.Status{
			State:    domain.ExecutionState(jobDB.State),
			Progress: jobDB.Progress,
		},
		SubmitDate: jobDB.SubmitDate, // Direct value use
		StartDate:  startDate,
		EndDate:    endDate,
	}
}

func FromDomainJob(job *domain.Job) *JobDB {
	isNew := job.ID == uuid.Nil

	jobID := job.ID
	submitDate := job.SubmitDate
	if isNew {
		jobID = uuid.New()
		submitDate = time.Now().UTC()
	}

	// Custom utility functions (like util.TimePtrToNull) would typically be used here
	// to convert *time.Time to sql.NullTime, but for simplicity, we'll inline the standard logic.

	var startDate sql.NullTime
	if job.StartDate != nil {
		startDate = sql.NullTime{Time: *job.StartDate, Valid: true}
	}
	var endDate sql.NullTime
	if job.EndDate != nil {
		endDate = sql.NullTime{Time: *job.EndDate, Valid: true}
	}

	return &JobDB{
		ID:            jobID,
		Name:          job.Name,
		Description:   sql.NullString{String: job.Description, Valid: job.Description != ""},
		ConfigID:      job.ConfigID,
		ConfigVersion: job.ConfigVersion,
		State:         string(job.State),
		Progress:      job.Progress,
		SubmitDate:    submitDate,
		StartDate:     startDate, // Uses standard sql.NullTime
		EndDate:       endDate,   // Uses standard sql.NullTime
	}
}
