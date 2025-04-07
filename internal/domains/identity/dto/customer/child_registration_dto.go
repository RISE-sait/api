package customer

import (
	commonDto "api/internal/domains/identity/dto/common"
	values "api/internal/domains/identity/values"
	errLib "api/internal/libs/errors"
	"api/internal/libs/validators"
)

type ChildRegistrationRequestDto struct {
	commonDto.UserBaseInfoRequestDto
	CustomerWaiversSigningDto []WaiverSigningRequestDto `json:"waivers"`
}

func (dto ChildRegistrationRequestDto) ToCreateChildValueObject(parentEmail string) (values.ChildRegistrationRequestInfo, *errLib.CommonError) {

	if err := validators.ValidateDto(&dto); err != nil {
		return values.ChildRegistrationRequestInfo{}, err
	}

	waiversVo := make([]values.CustomerWaiverSigning, len(dto.CustomerWaiversSigningDto))
	for i, waiver := range dto.CustomerWaiversSigningDto {
		vo, err := waiver.ToValueObjects()

		if err != nil {
			return values.ChildRegistrationRequestInfo{}, err
		}

		waiversVo[i] = values.CustomerWaiverSigning{
			IsWaiverSigned: vo.IsWaiverSigned,
			WaiverUrl:      vo.WaiverUrl,
		}
	}

	vo := values.ChildRegistrationRequestInfo{
		UserRegistrationRequestNecessaryInfo: values.UserRegistrationRequestNecessaryInfo{
			CountryCode: dto.CountryCode,
			Age:         dto.Age,
			FirstName:   dto.FirstName,
			LastName:    dto.LastName,
		},
		ParentEmail: parentEmail,
		Waivers:     waiversVo,
	}

	return vo, nil
}
