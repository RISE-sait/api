package values

import "github.com/google/uuid"

type MembershipInfo struct {
	Name     string
	PlanInfo MembershipPlanInfo
}

type MembershipPlanInfo struct {
	Id               uuid.UUID
	Name             string
	PlanRenewalDate  string
	Status           string
	StartDate        string
	UpdatedAt        string
	PaymentFrequency string
	AmtPeriods       *int32
	Price            int32
}
