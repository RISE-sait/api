package customer

import (
	"api/internal/domains/identity/dto/common"
	values "api/internal/domains/identity/values"
	errLib "api/internal/libs/errors"
	"api/internal/libs/validators"
	"api/utils/countries"
	"net/http"
)

type RegistrationRequestDto struct {
	identity.UserNecessaryInfoRequestDto
	CountryCode                string                    `json:"country_code"`
	CustomerWaiversSigningDto  []WaiverSigningRequestDto `json:"waivers"`
	PhoneNumber                string                    `json:"phone_number" validate:"omitempty,e164" example:"+15141234567"`
	HasConsentToSmS            bool                      `json:"has_consent_to_sms"`
	HasConsentToEmailMarketing bool                      `json:"has_consent_to_email_marketing"`
	Role                       string                    `json:"role" validate:"required"`
}

// ToParent validates the DTO and converts waiver signing details into value objects.
// Returns a slice of CustomerWaiverSigning value objects and an error if validation fails.
func (dto RegistrationRequestDto) ToParent(email string) (values.AdultCustomerRegistrationRequestInfo, *errLib.CommonError) {
	if err := validators.ValidateDto(&dto); err != nil {
		return values.AdultCustomerRegistrationRequestInfo{}, err
	}

	waiversVo := make([]values.CustomerWaiverSigning, len(dto.CustomerWaiversSigningDto))

	if dto.CustomerWaiversSigningDto != nil {

		for i, waiver := range dto.CustomerWaiversSigningDto {
			vo, err := waiver.ToValueObjects()

			if err != nil {
				return values.AdultCustomerRegistrationRequestInfo{}, err
			}

			waiversVo[i] = values.CustomerWaiverSigning{
				IsWaiverSigned: vo.IsWaiverSigned,
				WaiverUrl:      vo.WaiverUrl,
			}

		}
	}

	return values.AdultCustomerRegistrationRequestInfo{
		UserRegistrationRequestNecessaryInfo: values.UserRegistrationRequestNecessaryInfo{
			Age:       dto.Age,
			FirstName: dto.FirstName,
			LastName:  dto.LastName,
		},
		Email:                      email,
		Phone:                      dto.PhoneNumber,
		HasConsentToSms:            dto.HasConsentToSmS,
		HasConsentToEmailMarketing: dto.HasConsentToEmailMarketing,
	}, nil
}

// ToAthlete validates the DTO and converts waiver signing details into value objects.
// Returns a slice of CustomerWaiverSigning value objects and an error if validation fails.
func (dto RegistrationRequestDto) ToAthlete(email string) (values.AthleteRegistrationRequestInfo, *errLib.CommonError) {
	if err := validators.ValidateDto(&dto); err != nil {
		return values.AthleteRegistrationRequestInfo{}, err
	}

	if dto.CountryCode != "" && !countries.IsValidAlpha2Code(dto.CountryCode) {
		return values.AthleteRegistrationRequestInfo{}, errLib.New("Invalid country code", http.StatusBadRequest)
	}

	if dto.CustomerWaiversSigningDto == nil || len(dto.CustomerWaiversSigningDto) == 0 {
		return values.AthleteRegistrationRequestInfo{}, errLib.New("waivers: required", http.StatusBadRequest)
	}

	waiversVo := make([]values.CustomerWaiverSigning, len(dto.CustomerWaiversSigningDto))

	if dto.CustomerWaiversSigningDto != nil {

		for i, waiver := range dto.CustomerWaiversSigningDto {
			vo, err := waiver.ToValueObjects()

			if err != nil {
				return values.AthleteRegistrationRequestInfo{}, err
			}

			waiversVo[i] = values.CustomerWaiverSigning{
				IsWaiverSigned: vo.IsWaiverSigned,
				WaiverUrl:      vo.WaiverUrl,
			}

		}
	}

	return values.AthleteRegistrationRequestInfo{
		AdultCustomerRegistrationRequestInfo: values.AdultCustomerRegistrationRequestInfo{
			UserRegistrationRequestNecessaryInfo: values.UserRegistrationRequestNecessaryInfo{
				Age:       dto.Age,
				FirstName: dto.FirstName,
				LastName:  dto.LastName,
			},
			Email:                      email,
			Phone:                      dto.PhoneNumber,
			HasConsentToSms:            dto.HasConsentToSmS,
			HasConsentToEmailMarketing: dto.HasConsentToEmailMarketing,
		},
		Waivers:     waiversVo,
		CountryCode: dto.CountryCode,
	}, nil
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
