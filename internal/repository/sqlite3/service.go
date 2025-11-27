package sqlite3

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/abikandiah/task-worker/internal/domain"
	"github.com/abikandiah/task-worker/internal/platform/db"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type SQLiteServiceRepository struct {
	DB *sqlx.DB
}

func NewSQLiteServiceRepository(db *sqlx.DB) *SQLiteServiceRepository {
	return &SQLiteServiceRepository{
		DB: db,
	}
}

func (repo *SQLiteServiceRepository) Close() error {
	return repo.DB.Close()
}

func (repo *SQLiteServiceRepository) GetAllJobs(ctx context.Context, cursor *domain.CursorInput) (*domain.CursorOutput[domain.Job], error) {
	pq := &db.PaginationQuery{
		BaseQuery:     "SELECT id, name, description, config_id, state, progress, submit_date, start_date, end_date FROM jobs",
		AllowedFields: []string{"id", "state", "submit_date", "start_date", "end_date"},
	}

	return db.Paginate[domain.Job](ctx, repo.DB, pq, cursor)
}

func (repo *SQLiteServiceRepository) SaveJob(ctx context.Context, job domain.Job) (*domain.Job, error) {
	jobDB := FromDomainJob(&job)
	// Define the UPSERT query
	query := `
        INSERT INTO jobs (
            id, name, description, config_id, state, progress, submit_date, start_date, end_date
        ) VALUES (
            :id, :name, :description, :config_id, :state, :progress, :submit_date, :start_date, :end_date
        )
        ON CONFLICT (id) DO UPDATE SET
            name = EXCLUDED.name,
            description = EXCLUDED.description,
            config_id = EXCLUDED.config_id,
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
            id, name, description, config_id, state, progress, submit_date, start_date, end_date
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

func (repo *SQLiteServiceRepository) GetAllJobConfigs(ctx context.Context, input *domain.CursorInput) (*domain.CursorOutput[domain.JobConfig], error) {
	return nil, nil
}

func (repo *SQLiteServiceRepository) SaveJobConfig(ctx context.Context, config domain.JobConfig) (*domain.JobConfig, error) {
	configDB, err := FromDomainJobConfig(config)
	if err != nil {
		return nil, err
	}

	// Define the UPSERT query
	query := `
        INSERT INTO job_configs (
            id, name, description, version, details
        ) VALUES (
            :id, :name, :description, :version, :detailsjson
        )
        ON CONFLICT (id) DO UPDATE SET
            name = EXCLUDED.name,
            description = EXCLUDED.description,
            version = EXCLUDED.version,
            details = EXCLUDED.detailsjson
        WHERE 
            id = :id;
    `
	// 3. Execute the query using NamedExecContext
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

func (repo *SQLiteServiceRepository) GetTaskRun(ctx context.Context, taskID uuid.UUID) (*domain.TaskRun, error) {
	return nil, nil
}

func (repo *SQLiteServiceRepository) SaveTaskRun(ctx context.Context, taskRun domain.TaskRun) (*domain.TaskRun, error) {
	return nil, nil
}

func (repo *SQLiteServiceRepository) SaveTaskRuns(ctx context.Context, taskRuns []domain.TaskRun) ([]domain.TaskRun, error) {
	return nil, nil
}

func (repo *SQLiteServiceRepository) GetTaskRuns(ctx context.Context, jobID uuid.UUID) ([]domain.TaskRun, error) {
	return nil, nil
}
