package db

import (
	"database/sql/driver"
	"fmt"
	"time"
)

// TextTime is a custom type that scans TEXT dates into time.Time
// Works with SQLite, MySQL TEXT columns, or any DB storing dates as strings
type TextTime struct {
	time.Time
}

func (tt *TextTime) Scan(value interface{}) error {
	if value == nil {
		return fmt.Errorf("cannot scan nil into TextTime")
	}

	var timeStr string
	switch v := value.(type) {
	case string:
		timeStr = v
	case []byte:
		timeStr = string(v)
	case time.Time:
		// Some drivers might return time.Time directly
		tt.Time = v
		return nil
	default:
		return fmt.Errorf("unsupported type for TextTime: %T", value)
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
			tt.Time = t
			return nil
		}
	}

	return fmt.Errorf("unable to parse time: %s", timeStr)
}

func (tt TextTime) Value() (driver.Value, error) {
	return tt.Time.Format("2006-01-02 15:04:05.999"), nil
}

func NewNullTextTime(t *time.Time) NullTextTime {
	if t == nil {
		return NullTextTime{Valid: false}
	}
	return NullTextTime{
		TextTime: TextTime{Time: *t},
		Valid:    true,
	}
}

// NullTextTime handles nullable time fields
type NullTextTime struct {
	TextTime
	Valid bool
}

func (ntt *NullTextTime) Scan(value interface{}) error {
	if value == nil {
		ntt.Valid = false
		return nil
	}

	err := ntt.TextTime.Scan(value)
	if err != nil {
		return err
	}
	ntt.Valid = true
	return nil
}

func (ntt NullTextTime) Value() (driver.Value, error) {
	if !ntt.Valid {
		return nil, nil
	}
	return ntt.TextTime.Value()
}
