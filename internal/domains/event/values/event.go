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

type CreateEventValues struct {
	CreatedBy uuid.UUID
	Details
}

type UpdateEventValues struct {
	ID        uuid.UUID
	UpdatedBy uuid.UUID
	Details
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
