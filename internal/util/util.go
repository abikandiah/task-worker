package util

import (
	"database/sql"
	"time"
)

// Helper to convert *time.Time to sql.NullTime
func TimePtrToNull(t *time.Time) sql.NullTime {
	if t != nil {
		return sql.NullTime{Time: *t, Valid: true}
	}
	return sql.NullTime{}
}

// Helper to convert sql.NullTime to *time.Time
func NullTimePtr(nt sql.NullTime) *time.Time {
	if nt.Valid {
		return &nt.Time
	}
	return nil
}

// TimePtr safely converts a time.Time value to a *time.Time pointer.
// This is the idiomatic way to handle assignment to *time.Time fields
// and avoids the compiler error when trying to take the address of a temporary value.
func TimePtr(t time.Time) *time.Time {
	return &t
}
