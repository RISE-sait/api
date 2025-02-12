package values

import (
	"github.com/google/uuid"
)

type FacilityDetails struct {
	Name           string
	Location       string
	FacilityType   string
	FacilityTypeID uuid.UUID
}
