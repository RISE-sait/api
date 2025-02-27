package barber

import (
	"github.com/google/uuid"
	"time"
)

type EventValuesBase struct {
	BeginDateTime time.Time
	EndDateTime   time.Time
	BarberID      uuid.UUID
	CustomerID    uuid.UUID
}

type CreateEventValues struct {
	EventValuesBase
}

type UpdateEventValues struct {
	ID uuid.UUID
	EventValuesBase
}

type EventReadValues struct {
	ID        uuid.UUID
	CreatedAt time.Time
	UpdatedAt time.Time
	EventValuesBase
}
