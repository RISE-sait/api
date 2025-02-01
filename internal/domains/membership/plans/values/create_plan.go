package values

import (
	"github.com/google/uuid"
)

type MembershipPlanCreate struct {
	MembershipID     uuid.UUID
	Name             string
	Price            int64
	PaymentFrequency string
	AmtPeriods       int
}
