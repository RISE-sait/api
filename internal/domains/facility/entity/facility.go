package facility

import (
	values "api/internal/domains/facility/values"

	"github.com/google/uuid"
)

type Facility struct {
	ID uuid.UUID
	values.Details
}

type Category struct {
	ID   uuid.UUID
	Name string
}
