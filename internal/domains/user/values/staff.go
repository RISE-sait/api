package user

import (
	"time"

	"github.com/google/uuid"
)

type ReadValues struct {
	ID          uuid.UUID
	Email       string
	FirstName   string
	LastName    string
	HubspotID   string
	IsActive    bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
	RoleName    string
	CountryCode string
	Phone       string
}

type UpdateValues struct {
	ID       uuid.UUID
	IsActive bool
	RoleName string
}
