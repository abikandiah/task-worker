package sqlite3

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/abikandiah/task-worker/internal/domain"
	"github.com/abikandiah/task-worker/internal/platform/db"
	"github.com/google/uuid"
)

func (repo *SQLiteServiceRepository) SaveTaskRun(ctx context.Context, taskRun domain.TaskRun) (*domain.TaskRun, error) {
	// Convert domain model to database model, handling JSON marshaling errors
	taskRunDB, err := FromDomainTaskRun(taskRun)
	if err != nil {
		return nil, err
	}

	// Define the UPSERT query
	query := `
        INSERT INTO task_runs (
            id, job_id, name, description, task_name, state, start_date, end_date, details
        ) VALUES (
            :id, :job_id, :name, :description, :task_name, :state, :start_date, :end_date, :details
        )
        ON CONFLICT (id) DO UPDATE SET
            job_id = EXCLUDED.job_id,
            name = EXCLUDED.name,
            description = EXCLUDED.description,
            task_name = EXCLUDED.task_name,
            state = EXCLUDED.state,
            start_date = EXCLUDED.start_date,
            end_date = EXCLUDED.end_date,
            details = EXCLUDED.details
        WHERE 
            id = :id;
    `

	// Execute the query using NamedExecContext
	_, err = repo.DB.NamedExecContext(ctx, query, taskRunDB)
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

	// Prepare UPSERT query (same as SaveTaskRun)
	query := `
        INSERT INTO task_runs (
            id, job_id, name, description, task_name, state, start_date, end_date, details
        ) VALUES (
            :id, :job_id, :name, :description, :task_name, :state, :start_date, :end_date, :details
        )
        ON CONFLICT (id) DO UPDATE SET
            job_id = EXCLUDED.job_id,
            name = EXCLUDED.name,
            description = EXCLUDED.description,
            task_name = EXCLUDED.task_name,
            state = EXCLUDED.state,
            start_date = EXCLUDED.start_date,
            end_date = EXCLUDED.end_date,
            details = EXCLUDED.details
        WHERE 
            id = :id;
    `

	savedTaskRuns := make([]domain.TaskRun, 0, len(taskRuns))

	// Iterate and execute the UPSERT for each TaskRun within the transaction
	for _, taskRun := range taskRuns {
		// Convert domain model to database model
		taskRunDB, convErr := FromDomainTaskRun(taskRun)
		if convErr != nil {
			tx.Rollback()
			return nil, fmt.Errorf("conversion failed for task run %s: %w", taskRun.ID, convErr)
		}

		_, execErr := tx.NamedExecContext(ctx, query, taskRunDB)
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
	query := `
        SELECT 
            id, job_id, name, description, task_name, state, start_date, end_date, details
        FROM 
            task_runs
        WHERE 
            id = ?
    `

	var taskRunDB TaskRunDB
	err := repo.DB.GetContext(ctx, &taskRunDB, query, taskRunID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get task run with ID %s: %w", taskRunID, err)
	}

	return taskRunDB.ToDomainTaskRun()
}

func (repo *SQLiteServiceRepository) GetTaskRuns(ctx context.Context, jobID uuid.UUID) ([]domain.TaskRun, error) {
	query := `
        SELECT 
            id, job_id, name, description, task_name, state, start_date, end_date, details
        FROM 
            task_runs
        WHERE 
            job_id = ?
        ORDER BY 
            start_date ASC, id ASC -- Order by start time and then ID for deterministic results
    `

	// Fetch data into a slice of database models
	var taskRunDBs []TaskRunDB
	err := repo.DB.SelectContext(ctx, &taskRunDBs, query, jobID)
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
		BaseQuery:     "SELECT id, job_id, name, description, task_name, state, start_date, end_date, details FROM task_runs",
		AllowedFields: []string{"id", "job_id", "task_name", "state", "start_date", "end_date"}, // Fields allowed for sorting
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
