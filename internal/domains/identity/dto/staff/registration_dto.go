package staff

import (
	"api/internal/domains/identity/dto/common"
)

type RegistrationRequestDto struct {
	identity.UserNecessaryInfoRequestDto
	PhoneNumber   string `json:"phone_number" validate:"omitempty,e164" example:"+15141234567"`
	Role          string `json:"role" validate:"required"`
	IsActiveStaff bool   `json:"is_active_staff"`
}
