package event

import (
	"api/internal/custom_types"
	"github.com/google/uuid"
	"time"
)

type Details struct {
	Day              string
	ProgramStartAt   time.Time
	ProgramEndAt     time.Time
	SessionStartTime custom_types.TimeWithTimeZone
	SessionEndTime   custom_types.TimeWithTimeZone
	PracticeID       uuid.UUID
	CourseID         uuid.UUID
	GameID           uuid.UUID
	LocationID       uuid.UUID
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
