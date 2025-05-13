package user

import (
	"time"

	"github.com/google/uuid"
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
	ID        uuid.UUID
	FirstName string     
	LastName  string       
	Points    int32    
	Wins      int32     
	Losses    int32     
	Assists   int32    
	Rebounds  int32     
	Steals    int32   
	PhotoURL  *string    
	TeamID    *uuid.UUID 
}
