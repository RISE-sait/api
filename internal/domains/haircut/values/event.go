package haircut

import (
	"github.com/google/uuid"
	"time"
)

type EventValuesBase struct {
	BeginDateTime time.Time
	EndDateTime   time.Time
	BarberID      uuid.UUID
	CustomerID    uuid.UUID
	ServiceName   string
}

type CreateEventValues struct {
	EventValuesBase
}

type EventReadValues struct {
	ID        uuid.UUID
	CreatedAt time.Time
	UpdatedAt time.Time
	EventValuesBase
	BarberName   string
	CustomerName string
}
