package values

import (
	"github.com/google/uuid"
)

type MembershipPlanDetails struct {
	MembershipID     uuid.UUID
	Name             string
	Price            int64
	PaymentFrequency string
	AmtPeriods       *int
}
