package staff

import (
	"time"

	"github.com/google/uuid"
)

type ReadValues struct {
	ID        uuid.UUID
	HubspotID string
	IsActive  bool
	CreatedAt time.Time
	UpdatedAt time.Time
	RoleID    uuid.UUID
	RoleName  string
}

type UpdateValues struct {
	ID       uuid.UUID
	IsActive bool
	RoleName string
}

type CreateValues struct {
	IsActive bool
	RoleName string
}
