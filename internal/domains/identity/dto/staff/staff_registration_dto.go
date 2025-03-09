package staff

import (
	"api/internal/domains/identity/values"
	staffValues "api/internal/domains/user/values/staff"
	errLib "api/internal/libs/errors"
	"api/internal/libs/validators"
)

type RegistrationRequestDto struct {
	HubSpotID   string `json:"hubspot_id" validate:"required,notwhitespace"`
	RoleName    string `json:"role_name" validate:"required,notwhitespace"`
	IsActive    bool   `json:"is_active" validate:"required"`
	PhoneNumber string `json:"phone_number" validate:"e164" example:"+15141234567"`
}

func (dto RegistrationRequestDto) ToDetails() (*identity.StaffRegistrationRequestInfo, *errLib.CommonError) {

	if err := validators.ValidateDto(&dto); err != nil {
		return nil, err
	}

	vo := identity.StaffRegistrationRequestInfo{
		HubSpotID: dto.HubSpotID,
		CreateValues: staffValues.CreateValues{
			RoleName: dto.RoleName,
			IsActive: dto.IsActive,
		},
	}

	return &vo, nil
}
