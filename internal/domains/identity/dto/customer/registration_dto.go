package customer

import (
	"api/internal/domains/identity/dto/common"
	values "api/internal/domains/identity/values"
	errLib "api/internal/libs/errors"
	"api/internal/libs/validators"
)

type RegistrationRequestDto struct {
	identity.UserNecessaryInfoRequestDto
	CustomerWaiversSigningDto  []WaiverSigningRequestDto `json:"waivers"`
	PhoneNumber                string                    `json:"phone_number" validate:"omitempty,e164" example:"+15141234567"`
	HasConsentToSmS            bool                      `json:"has_consent_to_sms"`
	HasConsentToEmailMarketing bool                      `json:"has_consent_to_email_marketing"`
	Role                       string                    `json:"role" validate:"required"`
}

// toValueObjectBase validates the DTO and converts waiver signing details into value objects.
// Returns a slice of CustomerWaiverSigning value objects and an error if validation fails.
func (dto RegistrationRequestDto) toValueObjectBase() ([]values.CustomerWaiverSigning, *errLib.CommonError) {
	if err := validators.ValidateDto(&dto); err != nil {
		return nil, err
	}

	waiversVo := make([]values.CustomerWaiverSigning, len(dto.CustomerWaiversSigningDto))
	for i, waiver := range dto.CustomerWaiversSigningDto {
		vo, err := waiver.ToValueObjects()

		if err != nil {
			return nil, err
		}

		waiversVo[i] = values.CustomerWaiverSigning{
			IsWaiverSigned: vo.IsWaiverSigned,
			WaiverUrl:      vo.WaiverUrl,
		}
	}

	return waiversVo, nil
}

type ChildRegistrationRequestDto struct {
	identity.UserNecessaryInfoRequestDto
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
			Age:       dto.Age,
			FirstName: dto.FirstName,
			LastName:  dto.LastName,
		},
		ParentEmail: parentEmail,
		Waivers:     waiversVo,
	}

	return vo, nil
}
