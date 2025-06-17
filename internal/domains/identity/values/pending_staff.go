package identity

import (
	"time"

	"github.com/google/uuid"
)

// PendingStaffInfo represents a staff member awaiting approval.
type PendingStaffInfo struct {
	ID          uuid.UUID
	FirstName   string
	LastName    string
	Email       string
	Gender      *string
	Phone       *string
	CountryCode string
	RoleID      uuid.UUID
	CreatedAt   time.Time
	UpdatedAt   time.Time
	Dob         time.Time
}
