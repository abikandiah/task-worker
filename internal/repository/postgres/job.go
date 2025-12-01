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

// --- SQL Constants for jobs table ---

const selectJobByIDSQL = `
    SELECT 
        ` + queries.SelectJobFields + `
    FROM 
        jobs
    WHERE 
        id = $1
`

const insertJobSQL = `
    INSERT INTO jobs (
        ` + queries.SelectJobFields + `
    ) VALUES (
        $1, $2, $3, $4, $5, $6, $7, $8, $9, $10
    )
`

const upsertJobSQL = insertJobSQL + queries.UpsertJobConflictClause

type JobDB struct {
	models.CommonJobDB
	SubmitDate time.Time  `db:"submit_date"`
	StartDate  *time.Time `db:"start_date"`
	EndDate    *time.Time `db:"end_date"`
}

func (jobDB *JobDB) ToDomainJob() *domain.Job {
	job := jobDB.ToDomainJobBase()

	// Use native time.Time types directly
	job.SubmitDate = jobDB.SubmitDate
	job.StartDate = jobDB.StartDate
	job.EndDate = jobDB.EndDate

	return job
}

func FromDomainJob(job *domain.Job) *JobDB {
	isNew := job.ID == uuid.Nil
	submitDate := job.SubmitDate

	if isNew {
		submitDate = time.Now().UTC()
	}

	commonJobDb := models.NewCommonJobDB(job)

	return &JobDB{
		CommonJobDB: commonJobDb,
		SubmitDate:  submitDate,
		StartDate:   job.StartDate,
		EndDate:     job.EndDate,
	}
}

func (repo *PostgresServiceRepository) SaveJob(ctx context.Context, job domain.Job) (*domain.Job, error) {
	jobDB := FromDomainJob(&job)

	// Execute the query using positional parameters
	_, err := repo.DB.ExecContext(ctx, upsertJobSQL,
		jobDB.ID,
		jobDB.Name,
		jobDB.Description,
		jobDB.ConfigID,
		jobDB.ConfigVersion,
		jobDB.State,
		jobDB.Progress,
		jobDB.SubmitDate,
		jobDB.StartDate,
		jobDB.EndDate,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to upsert job %s: %w", jobDB.ID, err)
	}

	return jobDB.ToDomainJob(), nil
}

func (repo *PostgresServiceRepository) GetJob(ctx context.Context, jobID uuid.UUID) (*domain.Job, error) {
	// Execute query
	var jobDB JobDB
	err := repo.DB.GetContext(ctx, &jobDB, selectJobByIDSQL, jobID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get job with ID %s: %w", jobID, err)
	}

	return jobDB.ToDomainJob(), nil
}

func (repo *PostgresServiceRepository) GetAllJobs(ctx context.Context, cursor *domain.CursorInput) (*domain.CursorOutput[domain.Job], error) {
	pq := &db.PaginationQuery{
		BaseQuery:     queries.SelectPaginationJobSQL,
		AllowedFields: queries.JobPaginationAllowedFields,
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
