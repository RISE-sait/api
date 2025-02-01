package values

import (
	"github.com/google/uuid"
)

type FacilityUpdate struct {
	ID             uuid.UUID
	Name           string
	Location       string
	FacilityTypeID uuid.UUID
}
