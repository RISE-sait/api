package event

import (
	"api/internal/custom_types"
	"time"

	"github.com/google/uuid"
)

type Details struct {
	Day            string
	ProgramStartAt time.Time
	ProgramEndAt   *time.Time
	EventStartTime custom_types.TimeWithTimeZone
	EventEndTime   custom_types.TimeWithTimeZone
	ProgramID      uuid.UUID
	LocationID     uuid.UUID
	TeamID         uuid.UUID
	Capacity       *int32
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

type ReadEventValues struct {
	ID        uuid.UUID
	CreatedAt time.Time
	CreatedBy uuid.UUID
	UpdatedBy uuid.UUID
	UpdatedAt time.Time
	Details
	LocationName    string
	LocationAddress string
	ProgramName     string
	ProgramType     string
	TeamName        string

	Customers []Customer
	Staffs    []Staff
}

type Customer struct {
	ID                    uuid.UUID
	Email                 *string
	FirstName             string
	LastName              string
	Phone                 *string
	Gender                *string
	IsEnrollmentCancelled bool
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
