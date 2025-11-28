package sqlite3

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/abikandiah/task-worker/internal/domain"
	"github.com/abikandiah/task-worker/internal/platform/db"
	"github.com/abikandiah/task-worker/internal/repository/models"
	"github.com/google/uuid"
)

// --- SQL Constants for job_configs table ---
const selectConfigFields = "id, name, description, is_default, version, details"

const selectDefaultJobConfigSQL = `
    SELECT 
        ` + selectConfigFields + `
    FROM 
        job_configs
    WHERE 
        is_default = TRUE
`

const selectJobConfigByIDSQL = `
    SELECT 
        ` + selectConfigFields + `
    FROM 
        job_configs
    WHERE 
        id = ?
`

const insertJobConfigSQL = `
    INSERT INTO job_configs (
        ` + selectConfigFields + `
    ) VALUES (
        :id, :name, :description, :is_default, :version, :details
    )
`

const upsertJobConfigSQL = insertJobConfigSQL + `
    ON CONFLICT (id, version) DO UPDATE SET
        name = EXCLUDED.name,
        description = EXCLUDED.description,
        is_default = EXCLUDED.is_default,
        details = EXCLUDED.details;
`

const selectPaginationConfigSQL = `
    SELECT 
        ` + selectConfigFields + `
    FROM 
        job_configs
`

type JobConfigDB struct {
	models.CommonJobConfigDB
}

func (configDB *JobConfigDB) ToDomainJobConfig() (*domain.JobConfig, error) {
	return configDB.ToDomainJobConfigBase()
}

func FromDomainJobConfig(config domain.JobConfig) (JobConfigDB, error) {
	commonConfig, err := models.NewCommonJobConfigDB(config)
	return JobConfigDB{
		CommonJobConfigDB: commonConfig,
	}, err
}

func (repo *SQLiteServiceRepository) GetOrCreateDefaultJobConfig(ctx context.Context) (*domain.JobConfig, error) {
	// Loop with a few retries just in case
	for range 3 {
		config, err := repo.GetDefaultJobConfig(ctx)
		if err != nil {
			return nil, err
		}

		// Found default
		if config != nil {
			return config, nil
		}

		// Try and create default
		newConfig := domain.NewDefaultJobConfig()
		config, err = repo.SaveJobConfig(ctx, *newConfig)

		if err == nil {
			return config, nil
		}

		// Check if the error was due to the unique constraint
		if isSQLiteUniqueConstraintError(err) {
			// Wait and try again
			time.Sleep(100 * time.Millisecond)
			continue
		}

		// Must be a different error (DB connection, permissions, etc.)
		return nil, err
	}

	return nil, fmt.Errorf("failed to get or create default config after multiple retries")
}

func (repo *SQLiteServiceRepository) GetDefaultJobConfig(ctx context.Context) (*domain.JobConfig, error) {
	// Excute query
	var configDB JobConfigDB
	err := repo.DB.GetContext(ctx, &configDB, selectDefaultJobConfigSQL)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get default job config: %w", err)
	}

	return configDB.ToDomainJobConfig()
}

func (repo *SQLiteServiceRepository) SaveJobConfig(ctx context.Context, config domain.JobConfig) (*domain.JobConfig, error) {
	configDB, err := FromDomainJobConfig(config)
	if err != nil {
		return nil, err
	}

	var query string

	// --- Determine Query Type ---
	if config.IsDefault {
		// INSERT for the Default. We rely on the DB to fail if a default already exists.
		query = insertJobConfigSQL
	} else {
		// UPSERT for regular configs. Targets the full Composite PRIMARY KEY.
		query = upsertJobConfigSQL
	}

	// Execute the query using NamedExecContext
	_, err = repo.DB.NamedExecContext(ctx, query, configDB)
	if err != nil {
		return nil, fmt.Errorf("failed to upsert job config %s: %w", configDB.ID, err)
	}

	return configDB.ToDomainJobConfig()
}

func (repo *SQLiteServiceRepository) GetJobConfig(ctx context.Context, configID uuid.UUID) (*domain.JobConfig, error) {
	// Excute query
	var configDB JobConfigDB
	err := repo.DB.GetContext(ctx, &configDB, selectJobConfigByIDSQL, configID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get job config with ID %s: %w", configID, err)
	}

	return configDB.ToDomainJobConfig()
}

func (repo *SQLiteServiceRepository) GetAllJobConfigs(ctx context.Context, cursor *domain.CursorInput) (*domain.CursorOutput[domain.JobConfig], error) {
	pq := &db.PaginationQuery{
		BaseQuery:     selectPaginationConfigSQL,
		AllowedFields: []string{"id", "name", "version"},
	}

	// Paginate with DB struct for correct sqlx scanning
	dbOutput, err := db.Paginate[JobConfigDB](ctx, repo.DB, pq, cursor)
	if err != nil {
		return nil, err
	}

	// Convert JobConfigDB slice to JobConfig slice
	domainConfigs := make([]domain.JobConfig, len(dbOutput.Data))
	for i, configDB := range dbOutput.Data {
		domainConfig, err := configDB.ToDomainJobConfig()
		if err != nil {
			return nil, fmt.Errorf("failed to convert job config DB model to domain model: %w", err)
		}
		domainConfigs[i] = *domainConfig
	}

	// Create the final output structure using the domain type JobConfig
	domainOutput := &domain.CursorOutput[domain.JobConfig]{
		Limit:      dbOutput.Limit,
		Data:       domainConfigs,
		NextCursor: dbOutput.NextCursor,
		PrevCursor: dbOutput.PrevCursor,
	}

	return domainOutput, nil
}
