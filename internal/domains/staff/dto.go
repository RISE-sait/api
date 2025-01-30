package staff

import (
	errLib "api/internal/libs/errors"
	"net/http"
	"regexp"
	"strings"
)

type CreateStaffDto struct {
	Email         string `json:"email"`
	Role          string `json:"role"`
	IsActiveStaff bool   `json:"is_active_staff"`
}

func NewCreateStaffDto(email, role string, isActive bool) *CreateStaffDto {
	return &CreateStaffDto{
		Email:         email,
		Role:          strings.TrimSpace(role),
		IsActiveStaff: isActive,
	}
}

func (sc *CreateStaffDto) Validate() *errLib.CommonError {
	if len(sc.Role) > 100 {
		return errLib.New("Role is too long", http.StatusBadRequest)
	}

	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(sc.Email) {
		return errLib.New("Invalid email format for field 'email'", http.StatusBadRequest)
	}

	return nil
}
