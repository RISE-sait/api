package staff

import (
	entity "api/internal/domains/staff/entity"
	"time"

	"github.com/google/uuid"
)

// ResponseDto represents a staff member's details in API responses.
type ResponseDto struct {
	ID        uuid.UUID `json:"id"`
	IsActive  bool      `json:"is_active"` // Indicates if the staff is still an active employee
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	RoleID    uuid.UUID `json:"role_id"`
	RoleName  string    `json:"role_name"`
}

// NewStaffResponse creates a new ResponseDto from an entity.Staff.
func NewStaffResponse(staff entity.Staff) ResponseDto {
	return ResponseDto{
		ID:        staff.ID,
		IsActive:  staff.IsActive,
		CreatedAt: staff.CreatedAt,
		UpdatedAt: staff.UpdatedAt,
		RoleID:    staff.RoleID,
		RoleName:  staff.RoleName,
	}
}
