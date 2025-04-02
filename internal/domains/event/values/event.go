package event

import (
	"time"

	"github.com/google/uuid"
)

type Details struct {
	ScheduleID uuid.UUID
	StartAt    time.Time
	EndAt      time.Time
}

type MutationValues struct {
	ProgramID  uuid.UUID
	LocationID uuid.UUID
	TeamID     uuid.UUID
}

type CreateEventValues struct {
	CreatedBy uuid.UUID
	Capacity  int32
	Details
	MutationValues
}

type UpdateEventValues struct {
	ID        uuid.UUID
	UpdatedBy uuid.UUID
	Capacity  *int32
	Details
	MutationValues
}

type ReadPersonValues struct {
	ID        uuid.UUID
	FirstName string
	LastName  string
	Email     *string
}

type ReadEventValues struct {
	ID uuid.UUID

	CreatedAt time.Time
	UpdatedAt time.Time

	CreatedBy ReadPersonValues
	UpdatedBy ReadPersonValues

	Details

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
	ID        uuid.UUID
	Email     *string
	FirstName string
	LastName  string
	Phone     *string
	Gender    *string
}

type Staff struct {
	ID        uuid.UUID
	Email     string
	FirstName string
	LastName  string
	Phone     string
	Gender    *string
	RoleName  string
}
