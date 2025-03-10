package location

import (
	values "api/internal/domains/location/values"
	errLib "api/internal/libs/errors"
	"api/internal/libs/validators"
)

// RequestDto represents the data transfer object for facility-related requests.
// It is used to validate and map incoming JSON data to domain value objects.
type RequestDto struct {
	Name     string `json:"name" validate:"required,notwhitespace"`
	Location string `json:"location" validate:"required,notwhitespace"`
}

// ToCreateDetails converts the FacilityRequestDto into a FacilityCreate value object.
// It validates the DTO before conversion and returns an error if validation fails.
func (dto *RequestDto) ToCreateDetails() (values.CreateDetails, *errLib.CommonError) {

	if err := validators.ValidateDto(dto); err != nil {
		return values.CreateDetails{}, err
	}

	return values.CreateDetails{
		BaseDetails: values.BaseDetails{
			Name:    dto.Name,
			Address: dto.Location,
		},
	}, nil
}

// ToUpdateDetails converts the FacilityRequestDto into a FacilityUpdate value object.
// It parses and validates the provided HubSpotId string and ensures the DTO passes validation before conversion.
func (dto *RequestDto) ToUpdateDetails(idStr string) (values.UpdateDetails, *errLib.CommonError) {

	id, err := validators.ParseUUID(idStr)

	if err != nil {
		return values.UpdateDetails{}, err
	}

	if err = validators.ValidateDto(dto); err != nil {
		return values.UpdateDetails{}, err
	}
	return values.UpdateDetails{
		ID: id,
		BaseDetails: values.BaseDetails{
			Name:    dto.Name,
			Address: dto.Location,
		},
	}, nil
}
