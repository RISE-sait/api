package dto

import (
	"github.com/google/uuid"
)

type CreateFacilityRequest struct {
	Name           string    `json:"name" validate:"required_and_notwhitespace"`
	Location       string    `json:"location" validate:"required_and_notwhitespace"`
	FacilityTypeID uuid.UUID `json:"facility_type_id" validate:"required"`
}

type UpdateFacilityRequest struct {
	Name           string    `json:"name" validate:"required_and_notwhitespace"`
	Location       string    `json:"location" validate:"required_and_notwhitespace"`
	FacilityTypeID uuid.UUID `json:"facility_type_id" validate:"required"`
}

type FacilityResponse struct {
	ID             uuid.UUID `json:"id"`
	Name           string    `json:"name"`
	Location       string    `json:"location"`
	FacilityTypeID uuid.UUID `json:"facility_type_id"`
}
