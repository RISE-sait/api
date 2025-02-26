package facility

import (
	entity "api/internal/domains/facility/entity"
	values "api/internal/domains/facility/values"
	errLib "api/internal/libs/errors"
	"api/internal/libs/validators"

	"github.com/google/uuid"
)

// RequestDto represents the data transfer object for facility-related requests.
// It is used to validate and map incoming JSON data to domain value objects.
type RequestDto struct {
	Name           string    `json:"name" validate:"required,notwhitespace"`
	Location       string    `json:"location" validate:"required,notwhitespace"`
	FacilityTypeID uuid.UUID `json:"facility_type_id" validate:"required"`
}

// ToDetails converts the FacilityRequestDto into a FacilityCreate value object.
// It validates the DTO before conversion and returns an error if validation fails.
func (dto *RequestDto) ToDetails() (*values.Details, *errLib.CommonError) {

	if err := validators.ValidateDto(dto); err != nil {
		return nil, err
	}

	return &values.Details{
		Name:               dto.Name,
		Address:            dto.Location,
		FacilityCategoryID: dto.FacilityTypeID,
	}, nil
}

// ToEntity converts the FacilityRequestDto into a FacilityUpdate value object.
// It parses and validates the provided HubSpotId string and ensures the DTO passes validation before conversion.
func (dto *RequestDto) ToEntity(idStr string) (*entity.Facility, *errLib.CommonError) {

	id, err := validators.ParseUUID(idStr)

	if err != nil {
		return nil, err
	}

	if err := validators.ValidateDto(dto); err != nil {
		return nil, err
	}
	return &entity.Facility{
		ID: id,
		Details: values.Details{
			Name:               dto.Name,
			Address:            dto.Location,
			FacilityCategoryID: dto.FacilityTypeID,
		},
	}, nil
}

type CategoryRequestDto struct {
	Name string `json:"name" validate:"required,notwhitespace"`
}

func (dto *CategoryRequestDto) ToCreateFacilityCategoryValueObject() (*string, *errLib.CommonError) {

	if err := validators.ValidateDto(dto); err != nil {
		return nil, err
	}
	return &dto.Name, nil
}

func (dto *CategoryRequestDto) ToUpdateFacilityCategoryValueObject(idStr string) (*entity.Category, *errLib.CommonError) {

	id, err := validators.ParseUUID(idStr)

	if err != nil {
		return nil, err
	}

	if err := validators.ValidateDto(dto); err != nil {
		return nil, err
	}
	return &entity.Category{
		ID:   id,
		Name: dto.Name,
	}, nil
}
