package dto

import (
	"api/internal/domains/identity/values"
	errLib "api/internal/libs/errors"
	"api/internal/libs/validators"
)

type StaffRegistrationDto struct {
	UserInfoDto
	RoleName string `json:"role_name" validate:"required"`
	IsActive bool   `json:"is_active" validate:"required"`
}

func (dto *StaffRegistrationDto) ToValueObjects() (*values.StaffRegistrationInfo, *errLib.CommonError) {

	if err := validators.ValidateDto(dto); err != nil {
		return nil, err
	}

	vo := values.StaffRegistrationInfo{
		UserInfo: values.UserInfo{
			Email: dto.Email,
		},
		StaffDetails: values.StaffDetails{
			RoleName: dto.RoleName,
			IsActive: dto.IsActive,
		},
	}

	if dto.FirstName != "" {
		vo.FirstName = &dto.FirstName
	}

	if dto.LastName != "" {
		vo.LastName = &dto.LastName
	}

	return &vo, nil
}
