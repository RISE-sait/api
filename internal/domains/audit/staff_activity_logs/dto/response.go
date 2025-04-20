package staff_activity_logs

import (
	values "api/internal/domains/audit/staff_activity_logs/values"
	"github.com/google/uuid"
	"time"
)

type StaffActivityLogResponse struct {
	ID                  uuid.UUID `json:"id"`
	StaffID             uuid.UUID `json:"staff_id"`
	ActivityDescription string    `json:"activity_description"`
	CreatedAt           time.Time `json:"created_at"`
	FirstName           string    `json:"first_name"`
	LastName            string    `json:"last_name"`
	Email               string    `json:"email"`
}

func NewStaffActivityLogResponse(details values.StaffActivityLog) StaffActivityLogResponse {
	return StaffActivityLogResponse{
		ID:                  details.ID,
		StaffID:             details.StaffID,
		ActivityDescription: details.ActivityDescription,
		CreatedAt:           details.CreatedAt,
		FirstName:           details.FirstName,
		LastName:            details.LastName,
		Email:               details.Email,
	}
}
