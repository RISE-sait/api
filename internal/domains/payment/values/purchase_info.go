package payment

import (
	"time"

	"github.com/google/uuid"
)

type MembershipPlanJoiningRequirement struct {
	ID                 uuid.UUID
	Name               string
	SquarePlanID       string
	StripePriceID      string
	StripeJoiningFeeID string
	MembershipID       uuid.UUID
	AmtPeriods         *int32
	CreatedAt          time.Time
	UpdatedAt          time.Time
}
