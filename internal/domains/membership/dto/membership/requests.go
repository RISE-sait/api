package membership

import (
	values "api/internal/domains/membership/values"
	errLib "api/internal/libs/errors"
	"api/internal/libs/validators"
)

type RequestDto struct {
	Name        string `json:"name" validate:"required,notwhitespace" example:"Premium Membership"`
	Description string `json:"description" validate:"omitempty,notwhitespace" example:"Access to all premium features"`
}

func (dto *RequestDto) validate() *errLib.CommonError {
	if err := validators.ValidateDto(dto); err != nil {
		return err
	}
	return nil
}

func (dto *RequestDto) ToMembershipCreateValueObject() (*values.CreateValues, *errLib.CommonError) {

	if err := dto.validate(); err != nil {
		return nil, err
	}

	return &values.CreateValues{
		BaseValue: values.BaseValue{
			Name:        dto.Name,
			Description: dto.Description,
		},
	}, nil
}

func (dto *RequestDto) ToMembershipUpdateValueObject(idStr string) (*values.UpdateValues, *errLib.CommonError) {

	id, err := validators.ParseUUID(idStr)

	if err != nil {
		return nil, err
	}

	if err = dto.validate(); err != nil {
		return nil, err
	}

	return &values.UpdateValues{
		ID: id,
		BaseValue: values.BaseValue{
			Name:        dto.Name,
			Description: dto.Description,
		},
	}, nil
}
