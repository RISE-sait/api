package dto

import (
	"github.com/google/uuid"
)

type FacilityResponse struct {
	ID             uuid.UUID `json:"id"`
	Name           string    `json:"name"`
	Location       string    `json:"location"`
	FacilityTypeID uuid.UUID `json:"facility_type_id"`
}
