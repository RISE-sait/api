package values

import (
	"github.com/google/uuid"
)

type MembershipPlanUpdate struct {
	ID               uuid.UUID
	MembershipID     uuid.UUID
	Name             string
	Price            int64
	PaymentFrequency string
	AmtPeriods       int
}
