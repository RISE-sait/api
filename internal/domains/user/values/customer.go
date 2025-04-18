package user

import (
	"github.com/google/uuid"
	"time"
)

type MembershipReadValue struct {
	MembershipPlanID      uuid.UUID
	MembershipPlanName    string
	MembershipName        string
	MembershipStartDate   time.Time
	MembershipRenewalDate time.Time
}

type MembershipPlansReadValue struct {
	ID                 uuid.UUID
	CustomerID         uuid.UUID
	StartDate          time.Time
	RenewalDate        *time.Time
	Status             string
	CreatedAt          time.Time
	UpdatedAt          time.Time
	MembershipID       uuid.UUID
	MembershipPlanID   uuid.UUID
	MembershipName     string
	MembershipPlanName string
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
	Wins     int32
	Losses   int32
	Points   int32
	Steals   int32
	Assists  int32
	Rebounds int32
}
