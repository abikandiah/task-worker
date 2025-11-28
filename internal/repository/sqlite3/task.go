package sqlite3

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/abikandiah/task-worker/internal/domain"
	"github.com/abikandiah/task-worker/internal/platform/db"
	"github.com/abikandiah/task-worker/internal/repository/models"
	"github.com/google/uuid"
)

// --- SQL Constants for task_runs table ---
const selectTaskRunFields = "id, job_id, name, description, task_name, state, start_date, end_date, details"

const selectTaskRunByIDSQL = `
    SELECT 
        ` + selectTaskRunFields + `
    FROM 
        task_runs
    WHERE 
        id = ?
`

const insertTaskRunSQL = `
    INSERT INTO task_runs (
        ` + selectTaskRunFields + `
    ) VALUES (
		:id, :job_id, :name, :description, :task_name, :state, :start_date, :end_date, :details
    )
`

const upsertTaskRunSQL = insertTaskRunSQL + `
	ON CONFLICT (id) DO UPDATE SET
		job_id = EXCLUDED.job_id,
		name = EXCLUDED.name,
		description = EXCLUDED.description,
		task_name = EXCLUDED.task_name,
		state = EXCLUDED.state,
		start_date = EXCLUDED.start_date,
		end_date = EXCLUDED.end_date,
		details = EXCLUDED.details
`

const selectAllTaskRunsSQL = `
	SELECT 
		` + selectTaskRunFields + `
	FROM 
		task_runs
	WHERE 
		job_id = ?
	ORDER BY 
		start_date ASC, id ASC
`

const selectPaginationTaskRunSQL = `
    SELECT 
        ` + selectTaskRunFields + `
    FROM 
        task_runs
`

type TaskRunDB struct {
	models.CommonTaskRunDB
	StartDate db.NullTextTime `db:"start_date"`
	EndDate   db.NullTextTime `db:"end_date"`
}

func (taskRunDB *TaskRunDB) ToDomainTaskRun() (*domain.TaskRun, error) {
	taskRun, err := taskRunDB.ToDomainTaskRunBase()
	if err != nil {
		return taskRun, err
	}

	// Extract time.Time from TextTime
	if taskRunDB.StartDate.Valid {
		taskRun.StartDate = &taskRunDB.StartDate.Time
	}
	if taskRunDB.EndDate.Valid {
		taskRun.EndDate = &taskRunDB.EndDate.Time
	}

	return taskRun, nil
}

func FromDomainTaskRun(taskRun domain.TaskRun) (*TaskRunDB, error) {
	commonTaskRunDb, err := models.NewCommonTaskRunDB(taskRun)
	if err != nil {
		return nil, err
	}

	return &TaskRunDB{
		CommonTaskRunDB: commonTaskRunDb,
		StartDate:       db.NewNullTextTime(taskRun.StartDate),
		EndDate:         db.NewNullTextTime(taskRun.EndDate),
	}, nil
}

func (repo *SQLiteServiceRepository) SaveTaskRun(ctx context.Context, taskRun domain.TaskRun) (*domain.TaskRun, error) {
	// Convert domain model to database model, handling JSON marshaling errors
	taskRunDB, err := FromDomainTaskRun(taskRun)
	if err != nil {
		return nil, err
	}

	// Execute the query using NamedExecContext
	_, err = repo.DB.NamedExecContext(ctx, upsertTaskRunSQL, taskRunDB)
	if err != nil {
		return nil, fmt.Errorf("failed to upsert task run %s: %w", taskRunDB.ID, err)
	}

	return taskRunDB.ToDomainTaskRun()
}

func (repo *SQLiteServiceRepository) SaveTaskRuns(ctx context.Context, taskRuns []domain.TaskRun) ([]domain.TaskRun, error) {
	if len(taskRuns) == 0 {
		return []domain.TaskRun{}, nil
	}

	// Begin Transaction
	tx, err := repo.DB.BeginTxx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction for bulk task run save: %w", err)
	}

	savedTaskRuns := make([]domain.TaskRun, 0, len(taskRuns))

	// Iterate and execute the UPSERT for each TaskRun within the transaction
	for _, taskRun := range taskRuns {
		// Convert domain model to database model
		taskRunDB, convErr := FromDomainTaskRun(taskRun)
		if convErr != nil {
			tx.Rollback()
			return nil, fmt.Errorf("conversion failed for task run %s: %w", taskRun.ID, convErr)
		}

		_, execErr := tx.NamedExecContext(ctx, upsertTaskRunSQL, taskRunDB)
		if execErr != nil {
			tx.Rollback()
			return nil, fmt.Errorf("failed to upsert task run %s in transaction: %w", taskRunDB.ID, execErr)
		}

		// Use the converted DB model to get the final domain object
		finalTaskRun, convErr := taskRunDB.ToDomainTaskRun()
		if convErr != nil {
			tx.Rollback()
			return nil, fmt.Errorf("post-save conversion failed for task run %s: %w", taskRunDB.ID, convErr)
		}
		savedTaskRuns = append(savedTaskRuns, *finalTaskRun)
	}

	// Commit Transaction
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction for bulk task run save: %w", err)
	}

	return savedTaskRuns, nil
}

func (repo *SQLiteServiceRepository) GetTaskRun(ctx context.Context, taskRunID uuid.UUID) (*domain.TaskRun, error) {
	var taskRunDB TaskRunDB
	err := repo.DB.GetContext(ctx, &taskRunDB, selectTaskRunByIDSQL, taskRunID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get task run with ID %s: %w", taskRunID, err)
	}

	return taskRunDB.ToDomainTaskRun()
}

func (repo *SQLiteServiceRepository) GetTaskRuns(ctx context.Context, jobID uuid.UUID) ([]domain.TaskRun, error) {
	// Fetch data into a slice of database models
	var taskRunDBs []TaskRunDB
	err := repo.DB.SelectContext(ctx, &taskRunDBs, selectAllTaskRunsSQL, jobID)
	if err != nil {
		return nil, fmt.Errorf("failed to get task runs for job %s: %w", jobID, err)
	}

	// Convert database models to domain models
	domainTaskRuns := make([]domain.TaskRun, len(taskRunDBs))
	for i, taskRunDB := range taskRunDBs {
		domainTaskRun, convErr := taskRunDB.ToDomainTaskRun()
		if convErr != nil {
			return nil, fmt.Errorf("failed to convert task run DB model to domain model for ID %s: %w", taskRunDB.ID, convErr)
		}
		domainTaskRuns[i] = *domainTaskRun
	}

	return domainTaskRuns, nil
}

func (repo *SQLiteServiceRepository) GetAllTaskRuns(ctx context.Context, cursor *domain.CursorInput) (*domain.CursorOutput[domain.TaskRun], error) {
	pq := &db.PaginationQuery{
		BaseQuery:     selectPaginationTaskRunSQL,
		AllowedFields: []string{"id", "job_id", "task_name", "state", "start_date", "end_date"},
	}

	dbOutput, err := db.Paginate[TaskRunDB](ctx, repo.DB, pq, cursor)
	if err != nil {
		return nil, err
	}

	domainTaskRuns := make([]domain.TaskRun, len(dbOutput.Data))
	for i, taskRunDB := range dbOutput.Data {
		domainTaskRun, err := taskRunDB.ToDomainTaskRun()
		if err != nil {
			return nil, fmt.Errorf("failed to convert task run DB model to domain model: %w", err)
		}
		domainTaskRuns[i] = *domainTaskRun
	}

	domainOutput := &domain.CursorOutput[domain.TaskRun]{
		Limit:      dbOutput.Limit,
		Data:       domainTaskRuns,
		NextCursor: dbOutput.NextCursor,
		PrevCursor: dbOutput.PrevCursor,
	}

	return domainOutput, nil
}
