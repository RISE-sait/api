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
	PhotoURL    *string

	*CoachStatsReadValues
}

type StaffRoleReadValue struct {
	ID        uuid.UUID
	RoleName  string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type CoachStatsReadValues struct {
	Wins   int32
	Losses int32
}

type UpdateStaffValues struct {
	ID       uuid.UUID
	IsActive bool
	RoleName string
}

type UpdateStaffProfileValues struct {
	ID       uuid.UUID
	PhotoURL *string
}
