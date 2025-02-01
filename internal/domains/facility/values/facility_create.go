package values

import (
	"github.com/google/uuid"
)

type FacilityCreate struct {
	Name           string
	Location       string
	FacilityTypeID uuid.UUID
}
