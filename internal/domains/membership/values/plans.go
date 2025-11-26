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
	CreditAllocation    *int32
	WeeklyCreditLimit   *int32
}

type PlanCreateValues struct {
	PlanDetails
	// Fields for Stripe auto-creation (when StripePriceID is not provided)
	UnitAmount       *int64 // Price in cents
	Currency         string // "cad" or "usd"
	BillingInterval  string // "month", "year", "week", "day"
	IntervalCount    *int64 // defaults to 1
	JoiningFeeAmount *int64 // optional one-time fee in cents
}

type PlanUpdateValues struct {
	ID uuid.UUID
	PlanDetails
}

type PlanReadValues struct {
	ID uuid.UUID
	PlanDetails
	IsVisible  bool
	CreatedAt  time.Time
	UpdatedAt  time.Time
	UnitAmount int
	Currency   string
	Interval   string
}
