package dto

import (
	"api/internal/domains/identity/values"
	errLib "api/internal/libs/errors"
	"api/internal/libs/validators"
)

type ConfirmChildDto struct {
	ChildEmail  string `json:"child_email" validate:"required,email"`
	ParentEmail string `json:"parent_email" validate:"required,email"`
}

func (dto *ConfirmChildDto) ToValueObjects() (*values.ConfirmChild, *errLib.CommonError) {

	if err := validators.ValidateDto(dto); err != nil {
		return nil, err
	}

	return &values.ConfirmChild{
		ChildEmail:  dto.ChildEmail,
		ParentEmail: dto.ParentEmail,
	}, nil
}
