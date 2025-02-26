package values

import "github.com/google/uuid"

type EnrollmentDetails struct {
	CustomerId  uuid.UUID
	EventId     uuid.UUID
	IsCancelled bool
}
