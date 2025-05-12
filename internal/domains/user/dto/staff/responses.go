package staff

import (
	values "api/internal/domains/user/values"
	"time"

	"github.com/google/uuid"
)

// ResponseDto represents a staff member's details in API responses.
type ResponseDto struct {
	ID          uuid.UUID              `json:"id"`
	FirstName   string                 `json:"first_name"`
	LastName    string                 `json:"last_name"`
	CountryCode string                 `json:"country_code"`
	Email       string                 `json:"email"`
	Phone       string                 `json:"phone"`
	HubspotID   string                 `json:"hubspot_id"`
	IsActive    bool                   `json:"is_active"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	RoleName    string                 `json:"role_name"`
	PhotoURL    *string                `json:"photo_url"`
	CoachStats  *CoachStatsResponseDto `json:"coach_stats,omitempty"`
}

type CoachStatsResponseDto struct {
	Wins   int32 `json:"wins"`
	Losses int32 `json:"losses"`
}

// NewStaffResponse creates a new ResponseDto from an entity.Staff.
func NewStaffResponse(staff values.ReadValues) ResponseDto {
	return ResponseDto{
		ID:          staff.ID,
		Email:       staff.Email,
		FirstName:   staff.FirstName,
		LastName:    staff.LastName,
		CountryCode: staff.CountryCode,
		HubspotID:   staff.HubspotID,
		IsActive:    staff.IsActive,
		CreatedAt:   staff.CreatedAt,
		UpdatedAt:   staff.UpdatedAt,
		RoleName:    staff.RoleName,
		Phone:       staff.Phone,
		PhotoURL:    staff.PhotoURL,
		CoachStats:  (*CoachStatsResponseDto)(staff.CoachStatsReadValues),
	}
}
