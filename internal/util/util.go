package util

import (
	"fmt"
	"time"
)

// TimePtr safely converts a time.Time value to a *time.Time pointer.
// This is the idiomatic way to handle assignment to *time.Time fields
// and avoids the compiler error when trying to take the address of a temporary value.
func TimePtr(t time.Time) *time.Time {
	return &t
}

func ParseTime(timeStr string) (time.Time, error) {
	if timeStr == "" {
		return time.Time{}, fmt.Errorf("empty time string")
	}

	formats := []string{
		"2006-01-02 15:04:05.999",
		"2006-01-02 15:04:05",
		time.RFC3339,
		"2006-01-02T15:04:05.999Z07:00",
		"2006-01-02",
	}

	for _, format := range formats {
		t, err := time.Parse(format, timeStr)
		if err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("unable to parse time: %s", timeStr)
}

// parseTimePtr is like parseTime but returns nil for empty strings
func ParseTimePtr(timeStr string) (*time.Time, error) {
	if timeStr == "" {
		return nil, nil
	}

	t, err := ParseTime(timeStr)
	if err != nil {
		return nil, err
	}
	return &t, nil
}
