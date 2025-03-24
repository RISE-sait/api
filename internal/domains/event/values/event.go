package event

import (
	"api/internal/custom_types"
	"time"

	"github.com/google/uuid"
)

type Details struct {
	Day            string
	ProgramStartAt time.Time
	ProgramEndAt   time.Time
	EventStartTime custom_types.TimeWithTimeZone
	EventEndTime   custom_types.TimeWithTimeZone
	ProgramID      uuid.UUID
	LocationID     uuid.UUID
	Capacity       *int32
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
	LocationName    string
	LocationAddress string
	ProgramName     string
	ProgramType     string
}
