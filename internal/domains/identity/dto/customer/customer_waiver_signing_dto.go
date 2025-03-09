package customer

import (
	"api/internal/domains/identity/values"
	errLib "api/internal/libs/errors"
	"api/internal/libs/validators"
)

type WaiverSigningRequestDto struct {
	WaiverURL      string `json:"waiver_url" validate:"required,url"`
	IsWaiverSigned bool   `json:"is_waiver_signed"`
}

func (dto WaiverSigningRequestDto) ToValueObjects() (identity.CustomerWaiverSigning, *errLib.CommonError) {

	if err := validators.ValidateDto(&dto); err != nil {
		return identity.CustomerWaiverSigning{}, err
	}

	return identity.CustomerWaiverSigning{
		WaiverUrl:      dto.WaiverURL,
		IsWaiverSigned: dto.IsWaiverSigned,
	}, nil
}
