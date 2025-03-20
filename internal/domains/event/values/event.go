package event

import (
	"api/internal/custom_types"
	"github.com/google/uuid"
	"time"
)

type MutationDetails struct {
	Day            string
	ProgramStartAt time.Time
	ProgramEndAt   time.Time
	EventStartTime custom_types.TimeWithTimeZone
	EventEndTime   custom_types.TimeWithTimeZone
	PracticeID     uuid.UUID
	CourseID       uuid.UUID
	GameID         uuid.UUID
	LocationID     uuid.UUID
}

type ReadDetails struct {
	Day             string
	ProgramStartAt  time.Time
	ProgramEndAt    time.Time
	EventStartTime  custom_types.TimeWithTimeZone
	EventEndTime    custom_types.TimeWithTimeZone
	PracticeID      uuid.UUID
	CourseID        uuid.UUID
	GameID          uuid.UUID
	LocationID      uuid.UUID
	LocationAddress string
	CourseName      string
	PracticeName    string
	GameName        string
	LocationName    string
}

type CreateEventValues struct {
	MutationDetails
}

type UpdateEventValues struct {
	ID uuid.UUID
	MutationDetails
}

type ReadEventValues struct {
	ID        uuid.UUID
	CreatedAt time.Time
	UpdatedAt time.Time
	ReadDetails
}
