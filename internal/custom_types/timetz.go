package custom_types

import (
	"database/sql/driver"
	"fmt"
	"time"
)

// TimeWithTimeZone is a custom type for handling time values with time zones in a database.
// It ensures that the time is formatted in "HH:MM:SS-07:00" format and stored in UTC.
type TimeWithTimeZone struct {
	Time string
}

// Scan implements the sql.Scanner interface.
// It converts database values into the TimeWithTimeZone struct.
func (t *TimeWithTimeZone) Scan(value interface{}) error {
	if value == nil {
		t.Time = ""
		return nil
	}
	switch v := value.(type) {
	case time.Time:
		// If it's a time.Time, convert it to UTC and format it to string
		t.Time = v.UTC().Format("15:04:05-07:00")
	case string:
		// If it's a string, try to parse it using RFC3339 format
		parsedTime, err := time.Parse(time.RFC3339, v)
		if err != nil {
			return fmt.Errorf("failed to parse string to time: %v", err)
		}
		// Format the parsed time to only include the time and timezone in the format we want
		t.Time = parsedTime.Format("15:04:05-07:00")
	default:
		return fmt.Errorf("unable to scan into TimeWithTimeZone: %v", value)
	}
	return nil
}

// Value implements the driver.Valuer interface.
// It converts the TimeWithTimeZone struct into a format suitable for database storage.
func (t TimeWithTimeZone) Value() (driver.Value, error) {
	// Ensure the time is always stored in UTC with a timezone
	if t.Time == "" {
		return nil, fmt.Errorf("time string is empty")
	}

	// If the time string doesn't have a timezone part, assume UTC
	if len(t.Time) == 8 { // Format "15:04:05"
		t.Time = t.Time + "-00:00" // Append UTC timezone
	}

	parsedTime, err := time.Parse("15:04:05-07:00", t.Time)
	if err != nil {
		return nil, fmt.Errorf("failed to parse time for storage: %v", err)
	}

	// Store the time in UTC format
	return parsedTime.UTC().Format("15:04:05-07:00"), nil
}
