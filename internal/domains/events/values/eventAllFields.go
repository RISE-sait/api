package values

import "github.com/google/uuid"

type EventAllFields struct {
	EventDetails
	ID uuid.UUID
}
