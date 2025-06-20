package playground

import (
	"time"

	"github.com/google/uuid"
)

// CreateSessionValue represents the data required to create a new session.
type CreateSessionValue struct {
	SystemID   uuid.UUID
	CustomerID uuid.UUID
	StartTime  time.Time
	EndTime    time.Time
}

// Session represents a session in the playground domain.
type Session struct {
	ID                uuid.UUID
	SystemID          uuid.UUID
	SystemName        string
	CustomerID        uuid.UUID
	CustomerFirstName string
	CustomerLastName  string
	StartTime         time.Time
	EndTime           time.Time
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

// CreateSystemValue represents the data required to create a playground system.
type CreateSystemValue struct {
	Name string
}

// UpdateSystemValue represents the data required to update a playground system.
type UpdateSystemValue struct {
	ID   uuid.UUID
	Name string
}

// System represents a playground system entry.
type System struct {
	ID        uuid.UUID
	Name      string
	CreatedAt time.Time
	UpdatedAt time.Time
}
