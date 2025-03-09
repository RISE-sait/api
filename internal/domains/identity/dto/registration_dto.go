package identity

import (
	"api/internal/domains/identity/dto/common"
)

type UserNecessaryInfoRequestDto struct {
	FirstName string `json:"first_name" validate:"required,notwhitespace"`
	LastName  string `json:"last_name" validate:"required,notwhitespace"`
	Age       int    `json:"age" validate:"required,gt=0"`
}

type StaffRegistrationRequestDto struct {
	identity.UserNecessaryInfoRequestDto
	PhoneNumber   string `json:"phone_number" validate:"omitempty,e164" example:"+15141234567"`
	Role          string `json:"role" validate:"required"`
	IsActiveStaff bool   `json:"is_active_staff"`
}
