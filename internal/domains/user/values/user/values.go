package user

import (
	"github.com/google/uuid"
	"time"
)

type ReadValue struct {
	ID            uuid.UUID
	HubspotID     string
	ProfilePicUrl *string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type CustomerMembershipPlansReadValue struct {
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
