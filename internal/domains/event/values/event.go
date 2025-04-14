package event

import (
	"time"

	"github.com/google/uuid"
)

type Details struct {
	StartAt    time.Time
	EndAt      time.Time
	ProgramID  uuid.UUID
	LocationID uuid.UUID
	TeamID     uuid.UUID
	Capacity   int32
}

type CreateEventsRecurrenceValues struct {
	CreatedBy         uuid.UUID
	Day               *time.Weekday
	RecurrenceStartAt time.Time
	RecurrenceEndAt   time.Time
	EventStartTime    string
	EventEndTime      string
	ProgramID         uuid.UUID
	LocationID        uuid.UUID
	TeamID            uuid.UUID
	Capacity          int32
}

type CreateEventsSpecificValues struct {
	CreatedBy  uuid.UUID
	StartAt    time.Time
	EndAt      time.Time
	ProgramID  uuid.UUID
	LocationID uuid.UUID
	TeamID     uuid.UUID
	Capacity   int32
}

type UpdateEventValues struct {
	ID        uuid.UUID
	UpdatedBy uuid.UUID
	Details
}

type UpdateEventsValues struct {
	UpdatedBy                 uuid.UUID
	OriginalRecurrenceStartAt time.Time
	OriginalRecurrenceEndAt   time.Time
	OriginalRecurrenceDay     time.Weekday
	OriginalEventStartTime    string
	OriginalEventEndTime      string
	OriginalProgramID         uuid.UUID
	OriginalLocationID        uuid.UUID
	OriginalTeamID            uuid.UUID
	OriginalCapacity          int32

	NewRecurrenceStartAt time.Time
	NewRecurrenceEndAt   time.Time
	NewRecurrenceDay     time.Weekday
	NewEventStartTime    string
	NewEventEndTime      string
	NewProgramID         uuid.UUID
	NewLocationID        uuid.UUID
	NewTeamID            uuid.UUID
	NewCapacity          int32
}

type ReadPersonValues struct {
	ID        uuid.UUID
	FirstName string
	LastName  string
}

type ReadEventValues struct {
	ID uuid.UUID

	CreatedAt time.Time
	UpdatedAt time.Time

	CreatedBy ReadPersonValues
	UpdatedBy ReadPersonValues

	StartAt time.Time
	EndAt   time.Time

	Capacity int32

	Location struct {
		ID      uuid.UUID
		Name    string
		Address string
	}

	Program *struct {
		ID          uuid.UUID
		Name        string
		Description string
		Type        string
	}

	Team *struct {
		ID   uuid.UUID
		Name string
	}

	Customers []Customer

	Staffs []Staff
}

type Customer struct {
	ReadPersonValues
	Email                  *string
	Phone                  *string
	Gender                 *string
	HasCancelledEnrollment bool
}

type Staff struct {
	ReadPersonValues
	Email    string
	Phone    string
	Gender   *string
	RoleName string
}

type GetEventsFilter struct {
	Ids           []uuid.UUID
	ProgramType   string
	ProgramID     uuid.UUID
	LocationID    uuid.UUID
	ParticipantID uuid.UUID
	TeamID        uuid.UUID
	CreatedBy     uuid.UUID
	UpdatedBy     uuid.UUID
	Before        time.Time
	After         time.Time
}
