package user

import (
	"github.com/google/uuid"
	"time"
)

type MembershipPlansReadValue struct {
	ID               uuid.UUID
	CustomerID       uuid.UUID
	MembershipPlanID uuid.UUID
	StartDate        time.Time
	RenewalDate      *time.Time
	Status           string
	CreatedAt        time.Time
	UpdatedAt        time.Time
	MembershipName   string
}

type StatsUpdateValue struct {
	ID       uuid.UUID
	Wins     *int32
	Losses   *int32
	Points   *int32
	Steals   *int32
	Assists  *int32
	Rebounds *int32
}

type AthleteReadValue struct {
	ID            uuid.UUID
	ProfilePicUrl string
	Wins          int32
	Losses        int32
	Points        int32
	Steals        int32
	Assists       int32
	Rebounds      int32
	CreatedAt     time.Time
	UpdatedAt     time.Time
}
