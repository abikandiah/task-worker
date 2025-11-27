package sqlite3

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/abikandiah/task-worker/internal/domain"
	"github.com/abikandiah/task-worker/internal/platform/db"
	"github.com/google/uuid"
)

func (repo *SQLiteServiceRepository) SaveJob(ctx context.Context, job domain.Job) (*domain.Job, error) {
	jobDB := FromDomainJob(&job)
	// Define the UPSERT query
	query := `
        INSERT INTO jobs (
            id, name, description, config_id, config_version, state, progress, submit_date, start_date, end_date
        ) VALUES (
            :id, :name, :description, :config_id, :config_version, :state, :progress, :submit_date, :start_date, :end_date
        )
        ON CONFLICT (id) DO UPDATE SET
            name = EXCLUDED.name,
            description = EXCLUDED.description,
            config_id = EXCLUDED.config_id,
			config_version = EXCLUDED.config_version,
            state = EXCLUDED.state,
            progress = EXCLUDED.progress,
            start_date = EXCLUDED.start_date,
            end_date = EXCLUDED.end_date
        WHERE 
            id = :id;
    `

	// Execute the query using NamedExecContext
	_, err := repo.DB.NamedExecContext(ctx, query, jobDB)
	if err != nil {
		return nil, fmt.Errorf("failed to upsert job %s: %w", jobDB.ID, err)
	}

	return jobDB.ToDomainJob(), nil
}

func (repo *SQLiteServiceRepository) GetJob(ctx context.Context, jobID uuid.UUID) (*domain.Job, error) {
	query := `
        SELECT 
            id, name, description, config_id, config_version, state, progress, submit_date, start_date, end_date
        FROM 
            jobs
        WHERE 
            id = ?
    `

	// Excute query
	var jobDB JobDB
	err := repo.DB.GetContext(ctx, &jobDB, query, jobID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get job with ID %s: %w", jobID, err)
	}

	return jobDB.ToDomainJob(), nil
}

func (repo *SQLiteServiceRepository) GetAllJobs(ctx context.Context, cursor *domain.CursorInput) (*domain.CursorOutput[domain.Job], error) {
	pq := &db.PaginationQuery{
		BaseQuery:     "SELECT id, name, description, config_id, config_version, state, progress, submit_date, start_date, end_date FROM jobs",
		AllowedFields: []string{"id", "state", "submit_date", "start_date", "end_date"},
	}

	// Paginate with DB struct for correct sqlx scanning
	dbOutput, err := db.Paginate[JobDB](ctx, repo.DB, pq, cursor)
	if err != nil {
		return nil, err
	}

	// Convert JobDB slice to Job slice
	domainJobs := make([]domain.Job, len(dbOutput.Data))
	for i, jobDB := range dbOutput.Data {
		domainJobs[i] = *jobDB.ToDomainJob()
	}

	domainOutput := &domain.CursorOutput[domain.Job]{
		Limit:      dbOutput.Limit,
		Data:       domainJobs,
		NextCursor: dbOutput.NextCursor,
		PrevCursor: dbOutput.PrevCursor,
	}
	return domainOutput, nil
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
	query := `
        SELECT 
            id, name, description, version, details
        FROM 
            job_configs
        WHERE 
            is_default = TRUE
    `
	// Excute query
	var configDB JobConfigDB
	err := repo.DB.GetContext(ctx, &configDB, query)
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
		query = `
            INSERT INTO job_configs (
                id, name, description, is_default, version, details
            ) VALUES (
                :id, :name, :description, :is_default, :version, :details
            );
        `
	} else {
		// UPSERT for regular configs. Targets the full Composite PRIMARY KEY.
		query = `
            INSERT INTO job_configs (
                id, name, description, is_default, version, details
            ) VALUES (
                :id, :name, :description, :is_default, :version, :details
            )
            -- Target the full Composite Primary Key (id, version)
            ON CONFLICT (id, version) DO UPDATE SET
                name = EXCLUDED.name,
                description = EXCLUDED.description,
                is_default = EXCLUDED.is_default,
                details = EXCLUDED.details;
        `
	}

	// Execute the query using NamedExecContext
	_, err = repo.DB.NamedExecContext(ctx, query, configDB)
	if err != nil {
		return nil, fmt.Errorf("failed to upsert job config %s: %w", configDB.ID, err)
	}

	return configDB.ToDomainJobConfig()
}

func (repo *SQLiteServiceRepository) GetJobConfig(ctx context.Context, configID uuid.UUID) (*domain.JobConfig, error) {
	query := `
        SELECT 
            id, name, description, version, details
        FROM 
            job_configs
        WHERE 
            id = ?
    `

	// Excute query
	var configDB JobConfigDB
	err := repo.DB.GetContext(ctx, &configDB, query, configID)
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
		BaseQuery:     "SELECT id, name, description, version, details FROM job_configs",
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

func isSQLiteUniqueConstraintError(err error) bool {
	if err == nil {
		return false
	}
	// SQLite unique constraint errors often contain this specific phrase.
	// The table name and index/column name might also be present in the full message.
	return strings.Contains(err.Error(), "UNIQUE constraint failed")
}
