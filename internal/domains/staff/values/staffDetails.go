package values

import (
	"time"

	"github.com/google/uuid"
)

type StaffDetails struct {
	IsActive  bool
	CreatedAt time.Time
	UpdatedAt time.Time
	RoleID    uuid.UUID
	RoleName  string
}
