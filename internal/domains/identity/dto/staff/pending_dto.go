package staff

import (
	values "api/internal/domains/identity/values"
	"time"

	"github.com/google/uuid"
)

type PendingStaffResponseDto struct {
	ID          uuid.UUID `json:"id"`
	FirstName   string    `json:"first_name"`
	LastName    string    `json:"last_name"`
	Email       string    `json:"email"`
	Phone       *string   `json:"phone"`
	Gender      *string   `json:"gender,omitempty"`
	CountryCode string    `json:"country_code"`
	RoleID      uuid.UUID `json:"role_id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Dob         time.Time `json:"dob"`
}

func NewPendingStaffResponse(v values.PendingStaffInfo) PendingStaffResponseDto {
	return PendingStaffResponseDto{
		ID:          v.ID,
		FirstName:   v.FirstName,
		LastName:    v.LastName,
		Email:       v.Email,
		Phone:       v.Phone,
		Gender:      v.Gender,
		CountryCode: v.CountryCode,
		RoleID:      v.RoleID,
		CreatedAt:   v.CreatedAt,
		UpdatedAt:   v.UpdatedAt,
		Dob:         v.Dob,
	}
}
