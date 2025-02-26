package entity

import (
	"github.com/google/uuid"
	"time"
)

type Enrollment struct {
	ID          uuid.UUID
	CustomerID  uuid.UUID
	EventID     uuid.UUID
	CreatedAt   time.Time
	UpdatedAt   time.Time
	CheckedInAt time.Time
	IsCancelled bool
}
