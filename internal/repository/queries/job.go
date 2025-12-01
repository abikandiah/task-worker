package queries

// SelectJobFields contains all column names for the jobs table
const SelectJobFields = "id, name, description, config_id, config_version, state, progress, submit_date, start_date, end_date"

// SelectPaginationJobSQL is the base query for paginated job retrieval
const SelectPaginationJobSQL = `
    SELECT 
        ` + SelectJobFields + `
    FROM 
        jobs
`

// JobPaginationAllowedFields defines which fields can be used for sorting/filtering
var JobPaginationAllowedFields = []string{"id", "state", "submit_date", "start_date", "end_date"}

// UpsertJobConflictClause contains the common ON CONFLICT UPDATE logic
// Database-specific implementations prepend their INSERT statement
const UpsertJobConflictClause = `
    ON CONFLICT (id) DO UPDATE SET
        name = EXCLUDED.name,
        description = EXCLUDED.description,
        config_id = EXCLUDED.config_id,
        config_version = EXCLUDED.config_version,
        state = EXCLUDED.state,
        progress = EXCLUDED.progress,
        start_date = EXCLUDED.start_date,
        end_date = EXCLUDED.end_date
`
