package court

import (
	values "api/internal/domains/court/values"
	errLib "api/internal/libs/errors"
	"api/internal/libs/validators"

	"github.com/google/uuid"
)

// RequestDto validates court create/update requests
// swagger:model CourtRequest
// Note: simple fields only
// Example: {"name":"Court 1","location_id":"uuid"}
type RequestDto struct {
	Name       string    `json:"name" validate:"required,notwhitespace"`
	LocationID uuid.UUID `json:"location_id" validate:"required"`
}

// ToCreateDetails converts to CreateDetails after validation
func (dto *RequestDto) ToCreateDetails() (values.CreateDetails, *errLib.CommonError) {
	if err := validators.ValidateDto(dto); err != nil {
		return values.CreateDetails{}, err
	}
	return values.CreateDetails{
		BaseDetails: values.BaseDetails{
			Name:       dto.Name,
			LocationID: dto.LocationID,
		},
	}, nil
}

// ToUpdateDetails converts to UpdateDetails with given id
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
			Name:       dto.Name,
			LocationID: dto.LocationID,
		},
	}, nil
}
