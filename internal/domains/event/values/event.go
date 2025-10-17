package event

import (
	"time"

	"github.com/google/uuid"
)

//goland:noinspection GoNameStartsWithPackageName
type EventDetails struct {
	CreatedBy                 uuid.UUID
	StartAt                   time.Time
	EndAt                     time.Time
	ProgramID                 uuid.UUID
	LocationID                uuid.UUID
	CourtID                   uuid.UUID
	TeamID                    uuid.UUID
	RequiredMembershipPlanIDs []uuid.UUID
	PriceID                   string
	CreditCost                *int32
}

type CreateEventValues struct {
	CreatedBy uuid.UUID
	EventDetails
}

type UpdateEventValues struct {
	ID        uuid.UUID
	UpdatedBy uuid.UUID
	EventDetails
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

	Program struct {
		ID          uuid.UUID
		Name        string
		Description string
		Type        string
		PhotoURL    *string
	}

	Team *struct {
		ID   uuid.UUID
		Name string
	}

	Court *struct {
		ID   uuid.UUID
		Name string
	}

	RequiredMembershipPlanIDs []uuid.UUID
	PriceID                   *string
	CreditCost                *int32

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
	CourtID       uuid.UUID
	ParticipantID uuid.UUID
	TeamID        uuid.UUID
	CreatedBy     uuid.UUID
	UpdatedBy     uuid.UUID
	Before        time.Time
	After         time.Time
	Limit         int
	Offset        int
}
