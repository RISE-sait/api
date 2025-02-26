package staff

import (
	"time"

	"github.com/google/uuid"
)

type Details struct {
	IsActive  bool
	CreatedAt time.Time
	UpdatedAt time.Time
	RoleID    uuid.UUID
	RoleName  string
}
