package domain

import (
	"strings"

	"github.com/google/uuid"
)

type SortDirection string

const (
	SortASC          SortDirection = "ASC"
	SortDESC         SortDirection = "DESC"
	DefaultSortLimit               = 20
	MaxSortLimit                   = 100
)

// CursorInput defines the data needed to request the next/previous page.
type CursorInput struct {
	AfterID   uuid.UUID     `json:"afterId,omitempty"`
	BeforeID  uuid.UUID     `json:"beforeId,omitempty"`
	Limit     int           `json:"limit,omitempty"`
	SortField string        `json:"sortField,omitempty"`
	SortDir   SortDirection `json:"sortDir,omitempty"`
}

func (c *CursorInput) SetDefaults() {
	// 1. Limit Validation
	if c.Limit <= 0 || c.Limit > MaxSortLimit {
		c.Limit = DefaultSortLimit
	}

	// 2. SortField Default
	if c.SortField == "" {
		c.SortField = "id"
	}

	// 3. SortDir Validation and Default
	upperDir := SortDirection(strings.ToUpper(string(c.SortDir)))
	if upperDir != SortASC && upperDir != SortDESC {
		c.SortDir = SortASC // Default to ascending
	} else {
		c.SortDir = upperDir
	}
}

func (c *CursorInput) HasAfterCursor() bool {
	return c.AfterID != uuid.Nil
}

func (c *CursorInput) HasBeforeCursor() bool {
	return c.BeforeID != uuid.Nil
}

// CursorOutput includes cursors for the next/previous request.
type CursorOutput[T any] struct {
	// NextCursor will be uuid.Nil if no next page exists.
	NextCursor *uuid.UUID `json:"nextCursor,omitempty"`
	// PrevCursor will be uuid.Nil if this is the first page.
	PrevCursor *uuid.UUID `json:"prevCursor,omitempty"`
	Limit      int        `json:"limit"`
	Data       []T        `json:"data"`
}
