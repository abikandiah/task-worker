package db

import (
	"context"
	"fmt"
	"strings"

	"github.com/abikandiah/task-worker/internal/domain"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// The required CursorInput, CursorOutput, and SortDirection types are defined in db/job_repository.go
// for centralized placeholder definition.

// PaginationQuery builds a cursor-based pagination query
type PaginationQuery struct {
	BaseQuery     string   // The main SELECT query without ORDER BY or LIMIT
	AllowedFields []string // Whitelist of sortable fields for security
	TableAlias    string   // Optional table alias (e.g., "u" for "users u")
}

// Paginate executes the pagination query and returns structured results
func Paginate[T any](ctx context.Context, db *sqlx.DB, pq *PaginationQuery, cursor *domain.CursorInput) (*domain.CursorOutput[T], error) {
	query, args, err := pq.BuildQuery(cursor)
	if err != nil {
		return nil, fmt.Errorf("build pagination query: %w", err)
	}

	var results []T
	// Use SelectContext to fetch multiple rows
	if err := db.SelectContext(ctx, &results, query, args...); err != nil {
		return nil, fmt.Errorf("execute pagination query: %w", err)
	}

	// If we got BEFORE cursor results, reverse them back to natural order
	if cursor.HasBeforeCursor() {
		reverseSlice(results)
	}

	// Build the output with cursors
	output := &domain.CursorOutput[T]{
		Limit: cursor.Limit,
		Data:  results,
	}

	// Determine if there are more pages
	hasMore := len(results) > cursor.Limit
	if hasMore {
		// Remove the extra item we fetched
		results = results[:cursor.Limit]
		output.Data = results
	}

	// Set next/prev cursors
	if len(results) > 0 {
		// Extract IDs for cursors (assumes T has an ID field, enforced by the hasID interface)
		firstID := extractID(results[0])
		lastID := extractID(results[len(results)-1])

		if cursor.HasAfterCursor() || (!cursor.HasAfterCursor() && !cursor.HasBeforeCursor()) {
			// Forward pagination
			if cursor.HasAfterCursor() {
				output.PrevCursor = &cursor.AfterID
			}
			if hasMore {
				output.NextCursor = &lastID
			}
		} else if cursor.HasBeforeCursor() {
			// Backward pagination
			output.NextCursor = &cursor.BeforeID
			if hasMore {
				output.PrevCursor = &firstID
			}
		}
	}

	return output, nil
}

// BuildQuery constructs the full keyset pagination SQL query
func (pq *PaginationQuery) BuildQuery(cursor *domain.CursorInput) (string, []any, error) {
	cursor.SetDefaults()

	// Validate sort field against whitelist
	if !pq.isFieldAllowed(cursor.SortField) {
		return "", nil, fmt.Errorf("invalid sort field: %s", cursor.SortField)
	}

	var whereConditions []string
	var args []any

	// Build cursor condition
	if cursor.HasAfterCursor() {
		// Forward pagination (AFTER)
		condition := pq.buildCursorCondition(cursor.SortField, cursor.SortDir, ">", ">=")
		whereConditions = append(whereConditions, condition)
		args = append(args, cursor.AfterID, cursor.AfterID)
	} else if cursor.HasBeforeCursor() {
		// Backward pagination (BEFORE) - reverse the sort direction
		condition := pq.buildCursorCondition(cursor.SortField, pq.reverseSortDir(cursor.SortDir), ">", ">=")
		whereConditions = append(whereConditions, condition)
		args = append(args, cursor.BeforeID, cursor.BeforeID)
	}

	// Construct the query
	query := pq.BaseQuery

	// Add WHERE clause if cursor conditions exist
	if len(whereConditions) > 0 {
		// Ensure WHERE or AND is used correctly based on if the base query already has a WHERE
		if strings.Contains(strings.ToUpper(query), " WHERE ") {
			query += " AND " + strings.Join(whereConditions, " AND ")
		} else {
			query += " WHERE " + strings.Join(whereConditions, " AND ")
		}
	}

	// Add ORDER BY
	sortField := pq.qualifyField(cursor.SortField)
	idField := pq.qualifyField("id")

	sortDir := cursor.SortDir
	if cursor.HasBeforeCursor() {
		// Reverse sort direction for backward pagination, results will be reversed in Paginate
		sortDir = pq.reverseSortDir(cursor.SortDir)
	}

	// Keysets are always sorted by (sortField, id)
	query += fmt.Sprintf(" ORDER BY %s %s, %s %s", sortField, sortDir, idField, sortDir)

	// Add LIMIT (request one extra to determine if there's a next/prev page)
	query += fmt.Sprintf(" LIMIT %d", cursor.Limit+1)

	return query, args, nil
}

// buildCursorCondition creates the WHERE condition for cursor pagination (keyset pagination)
func (pq *PaginationQuery) buildCursorCondition(sortField string, sortDir domain.SortDirection, gtOp, eqOp string) string {
	qualifiedSortField := pq.qualifyField(sortField)
	qualifiedIDField := pq.qualifyField("id")

	// For DESC sorting, reverse the operators
	if sortDir == domain.SortDESC {
		gtOp = "<"
		eqOp = "<="
	}

	// Keyset pagination condition: (sortField > cursorValue) OR (sortField = cursorValue AND id > cursorID)
	// The inner SELECTs are used to safely retrieve the cursor value from the database based on the cursor ID.
	// NOTE: This complex query uses positional parameters (?) which are bound to cursorID and cursorValue.
	return fmt.Sprintf(
		"(%s %s (SELECT %s FROM jobs WHERE id = ?) OR (%s = (SELECT %s FROM jobs WHERE id = ?) AND %s %s ?))",
		qualifiedSortField, gtOp, sortField, qualifiedSortField, sortField, qualifiedIDField, eqOp,
	)
}

// qualifyField adds table alias to field name if alias exists
func (pq *PaginationQuery) qualifyField(field string) string {
	if pq.TableAlias != "" && !strings.Contains(field, ".") {
		return fmt.Sprintf("%s.%s", pq.TableAlias, field)
	}
	return field
}

// isFieldAllowed checks if the sort field is in the whitelist
func (pq *PaginationQuery) isFieldAllowed(field string) bool {
	for _, allowed := range pq.AllowedFields {
		if allowed == field {
			return true
		}
	}
	return false
}

// reverseSortDir reverses the sort direction
func (pq *PaginationQuery) reverseSortDir(dir domain.SortDirection) domain.SortDirection {
	if dir == domain.SortASC {
		return domain.SortDESC
	}
	return domain.SortASC
}

// --- Helper Functions ---

// hasID defines the interface required by Paginate to extract the ID for cursors.
type hasID interface {
	GetID() uuid.UUID
}

// extractID uses type assertion to extract ID field
func extractID(item interface{}) uuid.UUID {
	if v, ok := item.(hasID); ok {
		return v.GetID()
	}
	// Should not happen if T is a DB struct with GetID() defined
	return uuid.Nil
}

// reverseSlice reverses a slice in place
func reverseSlice[T any](s []T) {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
}
