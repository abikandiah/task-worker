package queries

// SelectConfigFields contains all column names for the job_configs table
const SelectConfigFields = "id, name, description, is_default, version, details"

// SelectDefaultJobConfigSQL retrieves the default job configuration
const SelectDefaultJobConfigSQL = `
    SELECT 
        ` + SelectConfigFields + `
    FROM 
        job_configs
    WHERE 
        is_default = TRUE
`

// SelectPaginationConfigSQL is the base query for paginated job config retrieval
const SelectPaginationConfigSQL = `
    SELECT 
        ` + SelectConfigFields + `
    FROM 
        job_configs
`

// JobConfigPaginationAllowedFields defines which fields can be used for sorting/filtering
var JobConfigPaginationAllowedFields = []string{"id", "name", "version"}

// UpsertJobConfigConflictClause contains the common ON CONFLICT UPDATE logic
// Database-specific implementations prepend their INSERT statement
const UpsertJobConfigConflictClause = `
    ON CONFLICT (id, version) DO UPDATE SET
        name = EXCLUDED.name,
        description = EXCLUDED.description,
        is_default = EXCLUDED.is_default,
        details = EXCLUDED.details
`
