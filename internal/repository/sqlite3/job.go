package sqlite3

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
        id = ?
`

const insertJobSQL = `
    INSERT INTO jobs (
        ` + queries.SelectJobFields + `
    ) VALUES (
		:id, :name, :description, :config_id, :config_version, :state, :progress, :submit_date, :start_date, :end_date
    )
`

const upsertJobSQL = insertJobSQL + queries.UpsertJobConflictClause

type JobDB struct {
	models.CommonJobDB
	SubmitDate db.TextTime     `db:"submit_date"`
	StartDate  db.NullTextTime `db:"start_date"`
	EndDate    db.NullTextTime `db:"end_date"`
}

func (jobDB *JobDB) ToDomainJob() *domain.Job {
	job := jobDB.ToDomainJobBase()

	// Extract time.Time from TextTime
	job.SubmitDate = jobDB.SubmitDate.Time
	if jobDB.StartDate.Valid {
		job.StartDate = &jobDB.StartDate.Time
	}
	if jobDB.EndDate.Valid {
		job.EndDate = &jobDB.EndDate.Time
	}

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
		SubmitDate:  db.TextTime{Time: submitDate},
		StartDate:   db.NewNullTextTime(job.StartDate),
		EndDate:     db.NewNullTextTime(job.EndDate),
	}
}

func (repo *SQLiteServiceRepository) SaveJob(ctx context.Context, job domain.Job) (*domain.Job, error) {
	jobDB := FromDomainJob(&job)
	// Execute the query using NamedExecContext
	_, err := repo.DB.NamedExecContext(ctx, upsertJobSQL, jobDB)
	if err != nil {
		return nil, fmt.Errorf("failed to upsert job %s: %w", jobDB.ID, err)
	}

	return jobDB.ToDomainJob(), nil
}

func (repo *SQLiteServiceRepository) GetJob(ctx context.Context, jobID uuid.UUID) (*domain.Job, error) {
	// Excute query
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

func (repo *SQLiteServiceRepository) GetAllJobs(ctx context.Context, cursor *domain.CursorInput) (*domain.CursorOutput[domain.Job], error) {
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
