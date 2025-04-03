package membership

import (
	"github.com/google/uuid"
	"time"
)

type PlanDetails struct {
	MembershipID        uuid.UUID
	Name                string
	AmtPeriods          *int32
	StripeJoiningFeesID string
	StripePriceID       string
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
	CreatedAt time.Time
	UpdatedAt time.Time
}
