package payment

import (
	"github.com/google/uuid"
	"time"
)

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
