package event

import (
	"time"

	"github.com/google/uuid"
)

// EventChanges tracks what changed between an existing event and an update
type EventChanges struct {
	StartTimeChanged bool
	EndTimeChanged   bool
	LocationChanged  bool

	// Store old/new values for message building
	OldStartAt      time.Time
	NewStartAt      time.Time
	OldEndAt        time.Time
	NewEndAt        time.Time
	OldLocationID   uuid.UUID
	NewLocationID   uuid.UUID
	OldLocationName string
	NewLocationName string
}

// HasSignificantChanges returns true if any key details changed that warrant notification
func (c EventChanges) HasSignificantChanges() bool {
	return c.StartTimeChanged || c.EndTimeChanged || c.LocationChanged
}

// DetectEventChanges compares an existing event with new update values to identify changes
func DetectEventChanges(existing ReadEventValues, new UpdateEventValues) EventChanges {
	return EventChanges{
		StartTimeChanged: !existing.StartAt.Equal(new.StartAt),
		EndTimeChanged:   !existing.EndAt.Equal(new.EndAt),
		LocationChanged:  existing.Location.ID != new.LocationID,
		OldStartAt:       existing.StartAt,
		NewStartAt:       new.StartAt,
		OldEndAt:         existing.EndAt,
		NewEndAt:         new.EndAt,
		OldLocationID:    existing.Location.ID,
		NewLocationID:    new.LocationID,
		OldLocationName:  existing.Location.Name,
	}
}
