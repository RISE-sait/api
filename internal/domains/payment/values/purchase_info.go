package values

import (
	"github.com/google/uuid"
	"time"
)

type MembershipPlanPurchaseInfo struct {
	CustomerId       uuid.UUID
	MembershipPlanId uuid.UUID
	Status           string
	RenewalDate      *time.Time
}

type MembershipPlanJoiningRequirement struct {
	ID                 uuid.UUID
	Name               string
	StripePriceID      string
	StripeJoiningFeeID string
	MembershipID       uuid.UUID
	AmtPeriods         *int32
	CreatedAt          time.Time
	UpdatedAt          time.Time
}
