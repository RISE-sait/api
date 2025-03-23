package team

import (
	"github.com/google/uuid"
	"time"
)

type Details struct {
	Name     string
	Capacity int32
	CoachID  uuid.UUID
}

type CreateTeamValues struct {
	Details
}

type UpdateTeamValues struct {
	ID          uuid.UUID
	TeamDetails Details
}

type GetTeamValues struct {
	TeamDetails Details
	ID          uuid.UUID
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
