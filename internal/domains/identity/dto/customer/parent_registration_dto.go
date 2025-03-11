package customer

import (
	"api/internal/domains/identity/dto/common"
	values "api/internal/domains/identity/values"
	errLib "api/internal/libs/errors"
	"api/internal/libs/validators"
)

type ParentRegistrationRequestDto struct {
	identity.UserNecessaryInfoRequestDto
	PhoneNumber                string `json:"phone_number" validate:"omitempty,e164" example:"+15141234567"`
	HasConsentToSmS            bool   `json:"has_consent_to_sms"`
	HasConsentToEmailMarketing bool   `json:"has_consent_to_email_marketing"`
}

// ToParent validates the DTO and converts waiver signing details into value objects.
// Returns a slice of CustomerWaiverSigning value objects and an error if validation fails.
func (dto ParentRegistrationRequestDto) ToParent(email string) (values.ParentRegistrationRequestInfo, *errLib.CommonError) {
	if err := validators.ValidateDto(&dto); err != nil {
		return values.ParentRegistrationRequestInfo{}, err
	}

	if err := dto.Validate(); err != nil {
		return values.ParentRegistrationRequestInfo{}, err
	}

	return values.ParentRegistrationRequestInfo{
		UserRegistrationRequestNecessaryInfo: values.UserRegistrationRequestNecessaryInfo{
			Age:         dto.Age,
			FirstName:   dto.FirstName,
			LastName:    dto.LastName,
			CountryCode: dto.CountryCode,
		},
		Email:                      email,
		Phone:                      dto.PhoneNumber,
		HasConsentToSms:            dto.HasConsentToSmS,
		HasConsentToEmailMarketing: dto.HasConsentToEmailMarketing,
	}, nil
}
