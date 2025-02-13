package dto

import (
	"api/internal/domains/identity/values"
	errLib "api/internal/libs/errors"
	"api/internal/libs/validators"
)

type CustomerRegistrationDto struct {
	CustomerWaiverSigningDtos []CustomerWaiverSigningDto `json:"waivers"`
	UserInfoDto
	Password string `json:"password" validate:"omitempty,min=8"`
}

func (dto *CustomerRegistrationDto) ToValueObjects() (*values.CustomerRegistrationInfo, *errLib.CommonError) {

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

	vo := values.CustomerRegistrationInfo{
		UserInfo: values.UserInfo{
			Email: dto.Email,
		},
		Waivers: waiversVo,
	}

	if dto.FirstName != "" {
		vo.FirstName = &dto.FirstName
	}

	if dto.LastName != "" {
		vo.LastName = &dto.LastName
	}

	if dto.Password != "" {
		vo.Password = &dto.Password
	}

	return &vo, nil
}
