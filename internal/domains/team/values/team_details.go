package team

import (
	"github.com/google/uuid"
	"time"
)

type Details struct {
	Name       string
	Capacity   int32
	CoachID    uuid.UUID
	CoachName  string
	CoachEmail string
	LogoURL    *string
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
	Roster      []RosterMemberInfo
}

type RosterMemberInfo struct {
	ID       uuid.UUID
	Email    string
	Country  string
	Name     string
	Points   int32
	Wins     int32
	Losses   int32
	Assists  int32
	Rebounds int32
	Steals   int32
}
