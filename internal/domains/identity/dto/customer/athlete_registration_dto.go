package customer

import (
	commonDto "api/internal/domains/identity/dto/common"
	values "api/internal/domains/identity/values"
	errLib "api/internal/libs/errors"
	"api/internal/libs/validators"
	"net/http"
	"time"
)

type AthleteRegistrationRequestDto struct {
	commonDto.UserBaseInfoRequestDto
	CustomerWaiversSigningDto       []WaiverSigningRequestDto `json:"waivers"`
	PhoneNumber                     string                    `json:"phone_number" validate:"omitempty,e164" example:"+15141234567"`
	HasConsentToSmS                 bool                      `json:"has_consent_to_sms"`
	HasConsentToEmailMarketing      bool                      `json:"has_consent_to_email_marketing"`
	EmergencyContactName            string                    `json:"emergency_contact_name" validate:"omitempty,max=100"`
	EmergencyContactPhone           string                    `json:"emergency_contact_phone" validate:"omitempty,max=25"`
	EmergencyContactRelationship    string                    `json:"emergency_contact_relationship" validate:"omitempty,max=50"`
}

// ToAthlete validates the DTO and converts waiver signing details into value objects.
// Returns a slice of CustomerWaiverSigning value objects and an error if validation fails.
func (dto AthleteRegistrationRequestDto) ToAthlete(email string) (values.AthleteRegistrationRequestInfo, *errLib.CommonError) {
	if err := validators.ValidateDto(&dto); err != nil {
		return values.AthleteRegistrationRequestInfo{}, err
	}

	if err := dto.Validate(); err != nil {
		return values.AthleteRegistrationRequestInfo{}, err
	}

	if len(dto.CustomerWaiversSigningDto) == 0 {
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

	dob, err := time.Parse("2006-01-02", dto.DOB)

	if err != nil {
		return values.AthleteRegistrationRequestInfo{}, errLib.New("Invalid date format", http.StatusBadRequest)
	}

	return values.AthleteRegistrationRequestInfo{
		UserRegistrationRequestNecessaryInfo: values.UserRegistrationRequestNecessaryInfo{
			DOB:         dob,
			FirstName:   dto.FirstName,
			LastName:    dto.LastName,
			CountryCode: dto.CountryCode,
		},
		Email:                        email,
		Phone:                        dto.PhoneNumber,
		HasConsentToSms:              dto.HasConsentToSmS,
		HasConsentToEmailMarketing:   dto.HasConsentToEmailMarketing,
		Waivers:                      waiversVo,
		EmergencyContactName:         dto.EmergencyContactName,
		EmergencyContactPhone:        dto.EmergencyContactPhone,
		EmergencyContactRelationship: dto.EmergencyContactRelationship,
	}, nil
}
