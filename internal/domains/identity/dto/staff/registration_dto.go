package staff

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
	PhoneNumber   string `json:"phone_number" validate:"omitempty,e164" example:"+15141234567"`
	Role          string `json:"role" validate:"required"`
	IsActiveStaff bool   `json:"is_active_staff"`
}

func (dto RegistrationRequestDto) ToCreateStaffValues() (values.StaffRegistrationRequestInfo, *errLib.CommonError) {
	if err := validators.ValidateDto(&dto); err != nil {
		return values.StaffRegistrationRequestInfo{}, err
	}

	if dto.CountryCode != "" && !countries.IsValidAlpha2Code(dto.CountryCode) {
		return values.StaffRegistrationRequestInfo{}, errLib.New("Invalid country code", http.StatusBadRequest)
	}

	return values.StaffRegistrationRequestInfo{
		UserRegistrationRequestNecessaryInfo: values.UserRegistrationRequestNecessaryInfo{
			Age:         dto.Age,
			FirstName:   dto.FirstName,
			LastName:    dto.LastName,
			CountryCode: dto.CountryCode,
		},
		StaffCreateValues: values.StaffCreateValues{
			IsActive: dto.IsActiveStaff,
			RoleName: dto.Role,
		},
	}, nil
}
