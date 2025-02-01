package dto

import (
	"api/internal/domains/identity/values"
	errLib "api/internal/libs/errors"
	"api/internal/libs/validators"
)

type LoginCredentialsDto struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

func (dto *LoginCredentialsDto) ToValueObjects() (*values.LoginCredentials, *errLib.CommonError) {

	if err := validators.ValidateDto(dto); err != nil {
		return nil, err
	}

	return &values.LoginCredentials{
		Email:    dto.Email,
		Password: dto.Password,
	}, nil
}
