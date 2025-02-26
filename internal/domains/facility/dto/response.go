package facility

import (
	entity "api/internal/domains/facility/entity"
	"github.com/google/uuid"
)

type ResponseDto struct {
	ID               uuid.UUID `json:"id"`
	Name             string    `json:"name"`
	Address          string    `json:"address"`
	FacilityCategory string    `json:"facility_category"`
}

func NewFacilityResponse(facility entity.Facility) ResponseDto {
	return ResponseDto{
		ID:               facility.ID,
		Name:             facility.Name,
		Address:          facility.Address,
		FacilityCategory: facility.FacilityCategoryName,
	}
}

type CategoryResponseDto struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}

func NewFacilityCategoryResponse(category entity.Category) CategoryResponseDto {
	return CategoryResponseDto{
		ID:   category.ID,
		Name: category.Name,
	}
}
