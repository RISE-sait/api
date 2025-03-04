package staff

import (
	values "api/internal/domains/user/values/staff"
	"time"

	"github.com/google/uuid"
)

// ResponseDto represents a staff member's details in API responses.
type ResponseDto struct {
	ID        uuid.UUID `json:"id"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Email     string    `json:"email"`
	Phone     string    `json:"phone"`
	HubspotID string    `json:"hubspot_id"`
	IsActive  bool      `json:"is_active"` // Indicates if the staff is still an active employee
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	RoleID    uuid.UUID `json:"role_id"`
	RoleName  string    `json:"role_name"`
}

// NewStaffResponse creates a new ResponseDto from an entity.Staff.
func NewStaffResponse(staff values.ReadValues) ResponseDto {
	return ResponseDto{
		ID:        staff.ID,
		HubspotID: staff.HubspotID,
		IsActive:  staff.IsActive,
		CreatedAt: staff.CreatedAt,
		UpdatedAt: staff.UpdatedAt,
		RoleID:    staff.RoleID,
		RoleName:  staff.RoleName,
	}
}
