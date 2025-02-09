package dto

import (
	"api/internal/domains/identity/values"
	errLib "api/internal/libs/errors"
	"api/internal/libs/validators"
)

type StaffCreationDto struct {
	RoleName string `json:"role_name" validate:"required"`
	IsActive bool   `json:"is_active"`
	RegisterCredentialsDto
}

func (dto *StaffCreationDto) ToValueObjects() (*values.StaffDetails, *errLib.CommonError) {

	if err := validators.ValidateDto(dto); err != nil {
		return nil, err
	}

	vo := values.StaffDetails{
		RoleName: dto.RoleName,
		IsActive: dto.IsActive,
	}

	return &vo, nil
}
