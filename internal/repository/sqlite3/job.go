package sqlite3

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/abikandiah/task-worker/internal/domain"
	"github.com/abikandiah/task-worker/internal/util"
	"github.com/google/uuid"
)

type JobDB struct {
	ID          uuid.UUID      `db:"id"`
	Name        string         `db:"name"`
	Description sql.NullString `db:"description"`
	ConfigID    uuid.UUID      `db:"config_id"`
	State       string         `db:"state"`
	Progress    float32        `db:"progress"`
	SubmitDate  time.Time      `db:"submit_date"`
	StartDate   sql.NullTime   `db:"start_date"`
	EndDate     sql.NullTime   `db:"end_date"`
}

type JobConfigDB struct {
	ID          uuid.UUID      `db:"id"`
	Name        string         `db:"name"`
	Description sql.NullString `db:"description"`
	Version     string         `db:"version"`
	DetailsJSON string         `db:"details"`
}

func (jobDB *JobDB) ToDomainJob() *domain.Job {
	identity := domain.Identity{
		ID: jobDB.ID,
		IdentitySubmission: domain.IdentitySubmission{
			Name:        jobDB.Name,
			Description: jobDB.Description.String,
		},
	}
	return &domain.Job{
		Identity: identity,
		ConfigID: jobDB.ConfigID,
		Status: domain.Status{
			State:    domain.ExecutionState(jobDB.State),
			Progress: jobDB.Progress,
		},
		SubmitDate: jobDB.SubmitDate,
		StartDate:  util.NullTimePtr(jobDB.StartDate),
		EndDate:    util.NullTimePtr(jobDB.EndDate),
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
		ID:          jobID,
		Name:        job.Name,
		Description: sql.NullString{String: job.Description, Valid: job.Description != ""},
		ConfigID:    job.ConfigID,
		State:       string(job.State),
		Progress:    job.Progress,
		SubmitDate:  submitDate,
		StartDate:   util.TimePtrToNull(job.StartDate),
		EndDate:     util.TimePtrToNull(job.EndDate),
	}
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
		DetailsJSON: string(detailsBytes),
	}, nil
}
