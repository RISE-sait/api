package entity

import "github.com/google/uuid"

type MembershipPlan struct {
	ID               uuid.UUID
	MembershipID     uuid.UUID
	Name             string
	Price            int64
	PaymentFrequency string
	AmtPeriods       int
}
