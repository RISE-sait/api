package dto

import (
	"time"

	"github.com/google/uuid"
)

type StaffResponseDto struct {
	Id        uuid.UUID `json:"id"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	RoleID    uuid.UUID `json:"role_id"`
	RoleName  string    `json:"role_name"`
}
