package event

import (
	"github.com/google/uuid"
	"time"
)

type Details struct {
	EventStartAt time.Time
	EventEndAt   time.Time
	PracticeID   uuid.UUID
	CourseID     uuid.UUID
	GameID       uuid.UUID
	LocationID   uuid.UUID
}

type CreateEventValues struct {
	Details
}

type UpdateEventValues struct {
	ID uuid.UUID
	Details
}

type ReadEventValues struct {
	ID        uuid.UUID
	CreatedAt time.Time
	UpdatedAt time.Time
	Details
}
