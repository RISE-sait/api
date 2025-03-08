package values

import (
	"github.com/google/uuid"
	"time"
)

type EnrollmentDetails struct {
	CustomerId uuid.UUID
	EventId    uuid.UUID
}

type EnrollmentCreateDetails struct {
	EnrollmentDetails
}

type EnrollmentUpdateDetails struct {
	ID uuid.UUID
	EnrollmentDetails
	IsCancelled bool
}

type EnrollmentReadDetails struct {
	ID          uuid.UUID
	CustomerID  uuid.UUID
	EventID     uuid.UUID
	CreatedAt   time.Time
	UpdatedAt   time.Time
	CheckedInAt time.Time
	IsCancelled bool
}
