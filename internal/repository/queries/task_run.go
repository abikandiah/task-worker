package queries

// SelectTaskRunFields contains all column names for the task_runs table
const SelectTaskRunFields = "id, job_id, name, description, task_name, state, start_date, end_date, details"

// SelectAllTaskRunsSQL retrieves all task runs for a specific job, ordered by start date
// Database-specific implementations add the appropriate parameter placeholder
const SelectAllTaskRunsBaseSQL = `
	SELECT 
		` + SelectTaskRunFields + `
	FROM 
		task_runs
	WHERE 
		job_id = `

// SelectAllTaskRunsOrderSQL is the ORDER BY clause for task runs
const SelectAllTaskRunsOrderSQL = `
	ORDER BY 
		start_date ASC, id ASC
`

// SelectPaginationTaskRunSQL is the base query for paginated task run retrieval
// Database-specific implementations add WHERE, ORDER BY, and LIMIT clauses
const SelectPaginationTaskRunSQL = `
    SELECT 
        ` + SelectTaskRunFields + `
    FROM 
        task_runs
`

// TaskRunPaginationAllowedFields defines which fields can be used for sorting/filtering
var TaskRunPaginationAllowedFields = []string{"id", "job_id", "task_name", "state", "start_date", "end_date"}

// UpsertTaskRunConflictClause contains the common ON CONFLICT UPDATE logic
// Database-specific implementations prepend their INSERT statement
const UpsertTaskRunConflictClause = `
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
