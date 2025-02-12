package dto

import (
	entity "api/internal/domains/facility/entities"
	"api/internal/domains/facility/values"
	errLib "api/internal/libs/errors"
	"api/internal/libs/validators"

	"github.com/google/uuid"
)

// FacilityRequestDto represents the data transfer object for facility-related requests.
// It is used to validate and map incoming JSON data to domain value objects.
type FacilityRequestDto struct {
	Name           string    `json:"name" validate:"notwhitespace"`
	Location       string    `json:"location" validate:"notwhitespace"`
	FacilityTypeID uuid.UUID `json:"facility_type_id" validate:"required"`
}

// ToFacilityCreateValueObject converts the FacilityRequestDto into a FacilityCreate value object.
// It validates the DTO before conversion and returns an error if validation fails.
func (dto *FacilityRequestDto) ToFacilityCreateValueObject() (*values.FacilityDetails, *errLib.CommonError) {

	if err := validators.ValidateDto(dto); err != nil {
		return nil, err
	}

	return &values.FacilityDetails{
		Name:           dto.Name,
		Location:       dto.Location,
		FacilityTypeID: dto.FacilityTypeID,
	}, nil
}

// ToFacilityUpdateValueObject converts the FacilityRequestDto into a FacilityUpdate value object.
// It parses and validates the provided ID string and ensures the DTO passes validation before conversion.
func (dto FacilityRequestDto) ToFacilityUpdateValueObject(idStr string) (*entity.Facility, *errLib.CommonError) {

	id, err := validators.ParseUUID(idStr)

	if err != nil {
		return nil, err
	}

	if err := validators.ValidateDto(&dto); err != nil {
		return nil, err
	}
	return &entity.Facility{
		ID: id,
		FacilityDetails: values.FacilityDetails{
			Name:           dto.Name,
			Location:       dto.Location,
			FacilityTypeID: dto.FacilityTypeID,
		},
	}, nil
}

type FacilityTypeRequestDto struct {
	Name string `json:"name" validate:"notwhitespace"`
}

func (dto FacilityTypeRequestDto) ToFacilityTypeUpdateValueObject(idStr string) (*entity.FacilityType, *errLib.CommonError) {

	id, err := validators.ParseUUID(idStr)

	if err != nil {
		return nil, err
	}

	if err := validators.ValidateDto(&dto); err != nil {
		return nil, err
	}
	return &entity.FacilityType{
		ID:   id,
		Name: dto.Name,
	}, nil
}
