package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/abikandiah/task-worker/internal/domain"
	"github.com/abikandiah/task-worker/internal/platform/db"
	"github.com/abikandiah/task-worker/internal/repository/models"
	"github.com/abikandiah/task-worker/internal/repository/queries"
	"github.com/google/uuid"
)

// --- SQL Constants for job_configs table ---
const selectJobConfigByIDSQL = `
    SELECT 
        ` + queries.SelectConfigFields + `
    FROM 
        job_configs
    WHERE 
        id = $1
`

const insertJobConfigSQL = `
    INSERT INTO job_configs (
        ` + queries.SelectConfigFields + `
    ) VALUES (
        $1, $2, $3, $4, $5, $6
    )
`

const upsertJobConfigSQL = insertJobConfigSQL + queries.UpsertJobConfigConflictClause

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

func (repo *PostgresServiceRepository) GetOrCreateDefaultJobConfig(ctx context.Context) (*domain.JobConfig, error) {
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
		if isPostgresUniqueConstraintError(err) {
			// Wait and try again
			time.Sleep(100 * time.Millisecond)
			continue
		}

		// Must be a different error (DB connection, permissions, etc.)
		return nil, err
	}

	return nil, fmt.Errorf("failed to get or create default config after multiple retries")
}

func (repo *PostgresServiceRepository) GetDefaultJobConfig(ctx context.Context) (*domain.JobConfig, error) {
	// Execute query
	var configDB JobConfigDB
	err := repo.DB.GetContext(ctx, &configDB, queries.SelectDefaultJobConfigSQL)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get default job config: %w", err)
	}

	return configDB.ToDomainJobConfig()
}

func (repo *PostgresServiceRepository) SaveJobConfig(ctx context.Context, config domain.JobConfig) (*domain.JobConfig, error) {
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

	// Execute the query using positional parameters
	_, err = repo.DB.ExecContext(ctx, query,
		configDB.ID,
		configDB.Name,
		configDB.Description,
		configDB.IsDefault,
		configDB.Version,
		configDB.DetailsJSON,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to upsert job config %s: %w", configDB.ID, err)
	}

	return configDB.ToDomainJobConfig()
}

func (repo *PostgresServiceRepository) GetJobConfig(ctx context.Context, configID uuid.UUID) (*domain.JobConfig, error) {
	// Execute query
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

func (repo *PostgresServiceRepository) GetAllJobConfigs(ctx context.Context, cursor *domain.CursorInput) (*domain.CursorOutput[domain.JobConfig], error) {
	pq := &db.PaginationQuery{
		BaseQuery:     queries.SelectPaginationConfigSQL,
		AllowedFields: queries.JobConfigPaginationAllowedFields,
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
