package dto

import (
	"api/internal/domains/identity/values"
	errLib "api/internal/libs/errors"
	"api/internal/libs/validators"
)

type CustomerRegistrationDto struct {
	CustomerWaiverSigningDtos []CustomerWaiverSigningDto `json:"waivers" validate:"required"`
	Email                     string                     `json:"email" validate:"required,email"`
	Password                  string                     `json:"password" validate:"omitempty,min=8"`
}

func (dto *CustomerRegistrationDto) ToValueObjects() (*values.CustomerRegistrationValueObject, *errLib.CommonError) {

	if err := validators.ValidateDto(dto); err != nil {
		return nil, err
	}

	waiversVo := make([]values.CustomerWaiverSigning, 0, len(dto.CustomerWaiverSigningDtos))

	for _, waiver := range dto.CustomerWaiverSigningDtos {
		vo, err := waiver.ToValueObjects()
		if err != nil {
			return nil, err
		}
		waiversVo = append(waiversVo, *vo)
	}

	vo := values.CustomerRegistrationValueObject{
		Email:    dto.Email,
		Password: &dto.Password,
		Waivers:  waiversVo,
	}

	if dto.Password != "" {
		vo.Password = &dto.Password
	}

	return &vo, nil
}
