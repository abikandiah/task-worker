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

// --- SQL Constants for task_runs table ---
const selectTaskRunByIDSQL = `
    SELECT 
        ` + queries.SelectTaskRunFields + `
    FROM 
        task_runs
    WHERE 
        id = $1
`

const insertTaskRunSQL = `
    INSERT INTO task_runs (
        ` + queries.SelectTaskRunFields + `
    ) VALUES (
		$1, $2, $3, $4, $5, $6, $7, $8, $9
    )
`

const upsertTaskRunSQL = insertTaskRunSQL + queries.UpsertTaskRunConflictClause

const selectAllTaskRunsSQL = queries.SelectAllTaskRunsBaseSQL + `$1` + queries.SelectAllTaskRunsOrderSQL

type TaskRunDB struct {
	models.CommonTaskRunDB
	StartDate *time.Time `db:"start_date"`
	EndDate   *time.Time `db:"end_date"`
}

func (taskRunDB *TaskRunDB) ToDomainTaskRun() (*domain.TaskRun, error) {
	taskRun, err := taskRunDB.ToDomainTaskRunBase()
	if err != nil {
		return taskRun, err
	}

	// Use native time.Time types directly
	taskRun.StartDate = taskRunDB.StartDate
	taskRun.EndDate = taskRunDB.EndDate

	return taskRun, nil
}

func FromDomainTaskRun(taskRun domain.TaskRun) (*TaskRunDB, error) {
	commonTaskRunDb, err := models.NewCommonTaskRunDB(taskRun)
	if err != nil {
		return nil, err
	}

	return &TaskRunDB{
		CommonTaskRunDB: commonTaskRunDb,
		StartDate:       taskRun.StartDate,
		EndDate:         taskRun.EndDate,
	}, nil
}

func (repo *PostgresServiceRepository) SaveTaskRun(ctx context.Context, taskRun domain.TaskRun) (*domain.TaskRun, error) {
	// Convert domain model to database model, handling JSON marshaling errors
	taskRunDB, err := FromDomainTaskRun(taskRun)
	if err != nil {
		return nil, err
	}

	// Execute the query using positional parameters
	_, err = repo.DB.ExecContext(ctx, upsertTaskRunSQL,
		taskRunDB.ID,
		taskRunDB.JobID,
		taskRunDB.Name,
		taskRunDB.Description,
		taskRunDB.TaskName,
		taskRunDB.State,
		taskRunDB.StartDate,
		taskRunDB.EndDate,
		taskRunDB.DetailsJSON,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to upsert task run %s: %w", taskRunDB.ID, err)
	}

	return taskRunDB.ToDomainTaskRun()
}

func (repo *PostgresServiceRepository) SaveTaskRuns(ctx context.Context, taskRuns []domain.TaskRun) ([]domain.TaskRun, error) {
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

		_, execErr := tx.ExecContext(ctx, upsertTaskRunSQL,
			taskRunDB.ID,
			taskRunDB.JobID,
			taskRunDB.Name,
			taskRunDB.Description,
			taskRunDB.TaskName,
			taskRunDB.State,
			taskRunDB.StartDate,
			taskRunDB.EndDate,
			taskRunDB.DetailsJSON,
		)
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

func (repo *PostgresServiceRepository) GetTaskRun(ctx context.Context, taskRunID uuid.UUID) (*domain.TaskRun, error) {
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

func (repo *PostgresServiceRepository) GetTaskRuns(ctx context.Context, jobID uuid.UUID) ([]domain.TaskRun, error) {
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

func (repo *PostgresServiceRepository) GetAllTaskRuns(ctx context.Context, cursor *domain.CursorInput) (*domain.CursorOutput[domain.TaskRun], error) {
	pq := &db.PaginationQuery{
		BaseQuery:     queries.SelectPaginationTaskRunSQL,
		AllowedFields: queries.TaskRunPaginationAllowedFields,
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
