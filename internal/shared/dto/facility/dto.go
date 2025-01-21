package dto

import (
	db "api/sqlc"

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

func (r *UpdateFacilityRequest) ToDBParams() *db.UpdateFacilityParams {

	dbParams := db.UpdateFacilityParams{
		Name:           r.Name,
		Location:       r.Location,
		FacilityTypeID: r.FacilityTypeID,
	}

	return &dbParams
}

type FacilityResponse struct {
	ID             uuid.UUID `json:"id"`
	Name           string    `json:"name"`
	Location       string    `json:"location"`
	FacilityTypeID uuid.UUID `json:"facility_type_id"`
}

func ToFacilityResponse(course *db.Facility) *FacilityResponse {
	return &FacilityResponse{
		ID:             course.ID,
		Name:           course.Name,
		Location:       course.Location,
		FacilityTypeID: course.FacilityTypeID,
	}
}
