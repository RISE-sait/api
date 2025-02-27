package event

import (
	"github.com/google/uuid"
	"time"
)

type Details struct {
	BeginDateTime time.Time
	EndDateTime   time.Time
	PracticeID    uuid.UUID
	CourseID      uuid.UUID
	LocationID    uuid.UUID
}

type CreateEventValues struct {
	Details
}

type UpdateEventValues struct {
	ID uuid.UUID
	Details
}
