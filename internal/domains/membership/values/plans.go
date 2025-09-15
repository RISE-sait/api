package membership

import (
	"time"

	"github.com/google/uuid"
)

type PlanDetails struct {
	MembershipID        uuid.UUID
	Name                string
	AmtPeriods          *int32
	StripeJoiningFeesID string
	StripePriceID       string
	JoiningFee          int
}

type PlanCreateValues struct {
	PlanDetails
}

type PlanUpdateValues struct {
	ID uuid.UUID
	PlanDetails
}

type PlanReadValues struct {
	ID uuid.UUID
	PlanDetails
	CreatedAt  time.Time
	UpdatedAt  time.Time
	UnitAmount int
	Currency   string
	Interval   string
}
