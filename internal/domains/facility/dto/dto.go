package dto

import (
	db "api/internal/domains/facility/infra/persistence/sqlc/generated"

	"github.com/google/uuid"
)

type CreateFacilityRequest struct {
	Name           string    `json:"name" validate:"required_and_notwhitespace"`
	Location       string    `json:"location" validate:"required_and_notwhitespace"`
	FacilityTypeID uuid.UUID `json:"facility_type_id" validate:"required"`
}

func (r *CreateFacilityRequest) ToDBParams() *db.CreateFacilityParams {

	dbParams := db.CreateFacilityParams{
		Name:           r.Name,
		Location:       r.Location,
		FacilityTypeID: r.FacilityTypeID,
	}

	return &dbParams
}

type UpdateFacilityRequest struct {
	ID             uuid.UUID `json:"id" validate:"required"`
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
