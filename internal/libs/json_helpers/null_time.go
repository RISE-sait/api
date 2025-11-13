package json_helpers

import (
	"database/sql"
	"encoding/json"
	"time"
)

// NullTime is a wrapper around sql.NullTime that properly marshals to JSON
// - If Valid is false, it marshals to JSON null
// - If Valid is true, it marshals the time in RFC3339 format
type NullTime struct {
	sql.NullTime
}

// MarshalJSON implements the json.Marshaler interface
func (nt NullTime) MarshalJSON() ([]byte, error) {
	if !nt.Valid {
		return []byte("null"), nil
	}
	return json.Marshal(nt.Time.Format(time.RFC3339))
}

// UnmarshalJSON implements the json.Unmarshaler interface
func (nt *NullTime) UnmarshalJSON(data []byte) error {
	// Handle null
	if string(data) == "null" {
		nt.Valid = false
		return nil
	}

	// Parse time string
	var timeStr string
	if err := json.Unmarshal(data, &timeStr); err != nil {
		return err
	}

	parsedTime, err := time.Parse(time.RFC3339, timeStr)
	if err != nil {
		return err
	}

	nt.Time = parsedTime
	nt.Valid = true
	return nil
}

// NullTimeToPtr converts sql.NullTime to *time.Time
// Returns nil if the NullTime is not valid
func NullTimeToPtr(nt sql.NullTime) *time.Time {
	if !nt.Valid {
		return nil
	}
	return &nt.Time
}

// PtrToNullTime converts *time.Time to sql.NullTime
func PtrToNullTime(t *time.Time) sql.NullTime {
	if t == nil {
		return sql.NullTime{Valid: false}
	}
	return sql.NullTime{Time: *t, Valid: true}
}
