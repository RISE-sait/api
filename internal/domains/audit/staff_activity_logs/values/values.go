package staff_activity_logs

import (
	"github.com/google/uuid"
	"time"
)

type StaffActivityLog struct {
	ID                  uuid.UUID
	StaffID             uuid.UUID
	ActivityDescription string
	CreatedAt           time.Time
	FirstName           string
	LastName            string
	Email               string
}
