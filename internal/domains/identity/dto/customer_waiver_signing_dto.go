package dto

import (
	"api/internal/domains/identity/values"
	errLib "api/internal/libs/errors"
	"api/internal/libs/validators"
)

type CustomerWaiverSigningDto struct {
	WaiverUrl      string `json:"waiver_url" validate:"required,url"`
	IsWaiverSigned bool   `json:"is_waiver_signed"`
}

func (dto *CustomerWaiverSigningDto) ToValueObjects() (*values.CustomerWaiverSigning, *errLib.CommonError) {

	if err := validators.ValidateDto(dto); err != nil {
		return nil, err
	}

	return &values.CustomerWaiverSigning{
		WaiverUrl:      dto.WaiverUrl,
		IsWaiverSigned: dto.IsWaiverSigned,
	}, nil
}
