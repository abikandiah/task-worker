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

type JobDB struct {
	ID            uuid.UUID       `db:"id"`
	Name          string          `db:"name"`
	Description   sql.NullString  `db:"description"`
	ConfigID      uuid.UUID       `db:"config_id"`
	ConfigVersion uuid.UUID       `db:"config_version"`
	State         string          `db:"state"`
	Progress      float32         `db:"progress"`
	SubmitDate    db.TextTime     `db:"submit_date"`
	StartDate     db.NullTextTime `db:"start_date"`
	EndDate       db.NullTextTime `db:"end_date"`
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

	// Extract time.Time from TextTime
	var startDate, endDate *time.Time
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
		SubmitDate: jobDB.SubmitDate.Time,
		StartDate:  startDate,
		EndDate:    endDate,
	}
}

func FromDomainJob(job *domain.Job) *JobDB {
	// Check if new
	isNew := job.ID == uuid.Nil

	jobID := job.ID
	submitDate := job.SubmitDate
	if isNew {
		jobID = uuid.New()
		submitDate = time.Now().UTC()
	}

	return &JobDB{
		ID:            jobID,
		Name:          job.Name,
		Description:   sql.NullString{String: job.Description, Valid: job.Description != ""},
		ConfigID:      job.ConfigID,
		ConfigVersion: job.ConfigVersion,
		State:         string(job.State),
		Progress:      job.Progress,
		SubmitDate:    db.TextTime{Time: submitDate},
		StartDate:     db.NewNullTextTime(job.StartDate),
		EndDate:       db.NewNullTextTime(job.EndDate),
	}
}

type JobConfigDB struct {
	ID          uuid.UUID      `db:"id"`
	Version     uuid.UUID      `db:"version"`
	Name        string         `db:"name"`
	Description sql.NullString `db:"description"`
	IsDefault   bool           `db:"is_default"`
	DetailsJSON string         `db:"details"`
}

// GetID implements the required method for cursor pagination.
func (configDB JobConfigDB) GetID() uuid.UUID {
	return configDB.ID
}

func (configDB *JobConfigDB) ToDomainJobConfig() (*domain.JobConfig, error) {
	identity := domain.Identity{
		ID: configDB.ID,
		IdentitySubmission: domain.IdentitySubmission{
			Name:        configDB.Name,
			Description: configDB.Description.String,
		},
	}
	config := &domain.JobConfig{
		IdentityVersion: domain.IdentityVersion{
			Identity: identity,
			Version:  configDB.Version,
		},
		IsDefault: configDB.IsDefault,
	}

	// Unmarshal the DetailsJSON string back into the JobConfigDetails struct
	if configDB.DetailsJSON != "" {
		err := json.Unmarshal([]byte(configDB.DetailsJSON), &config.JobConfigDetails)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal JobConfig details JSON: %w", err)
		}
	}

	return config, nil
}

func FromDomainJobConfig(config domain.JobConfig) (JobConfigDB, error) {
	// Marshal the Details struct into a JSON string
	detailsBytes, err := json.Marshal(config.JobConfigDetails)
	if err != nil {
		return JobConfigDB{}, fmt.Errorf("failed to marshal JobConfig details: %w", err)
	}

	// Set ID if new config
	jobConfigID := config.ID
	if config.ID == uuid.Nil {
		jobConfigID = uuid.New()
	}

	return JobConfigDB{
		ID:          jobConfigID,
		Name:        config.Name,
		Description: sql.NullString{String: config.Description, Valid: config.Description != ""},
		Version:     config.Version,
		IsDefault:   config.IsDefault,
		DetailsJSON: string(detailsBytes),
	}, nil
}
