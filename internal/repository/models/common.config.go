package models

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/abikandiah/task-worker/internal/domain"
	"github.com/google/uuid"
)

type CommonJobConfigDB struct {
	ID          uuid.UUID      `db:"id"`
	Version     uuid.UUID      `db:"version"`
	Name        string         `db:"name"`
	Description sql.NullString `db:"description"`
	IsDefault   bool           `db:"is_default"`
	DetailsJSON string         `db:"details"`
}

// GetID implements the required method for cursor pagination.
func (configDB CommonJobConfigDB) GetID() uuid.UUID {
	return configDB.ID
}

func (configDB *CommonJobConfigDB) ToDomainJobConfigBase() (*domain.JobConfig, error) {
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

func NewCommonJobConfigDB(config domain.JobConfig) (CommonJobConfigDB, error) {
	// Marshal the Details struct into a JSON string
	detailsBytes, err := json.Marshal(config.JobConfigDetails)
	if err != nil {
		return CommonJobConfigDB{}, fmt.Errorf("failed to marshal JobConfig details: %w", err)
	}

	// Set ID if new config
	jobConfigID := config.ID
	if config.ID == uuid.Nil {
		jobConfigID = uuid.New()
	}

	return CommonJobConfigDB{
		ID:          jobConfigID,
		Name:        config.Name,
		Description: sql.NullString{String: config.Description, Valid: config.Description != ""},
		Version:     config.Version,
		IsDefault:   config.IsDefault,
		DetailsJSON: string(detailsBytes),
	}, nil
}
