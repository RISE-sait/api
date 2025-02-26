package membership

import (
	"github.com/google/uuid"
)

type PlanCreateValues struct {
	MembershipID     uuid.UUID
	Name             string
	Price            int64
	PaymentFrequency *string
	AmtPeriods       *int
}

type PlanUpdateValues struct {
	ID               uuid.UUID
	MembershipID     uuid.UUID
	Name             string
	Price            int64
	PaymentFrequency *string
	AmtPeriods       *int
}

type PlanReadValues struct {
	ID               uuid.UUID
	MembershipID     uuid.UUID
	Name             string
	Price            int64
	PaymentFrequency *string
	AmtPeriods       *int
}
