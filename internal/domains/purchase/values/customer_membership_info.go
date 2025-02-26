package values

import (
	"github.com/google/uuid"
	"time"
)

type MembershipPlanPurchaseInfo struct {
	CustomerId       uuid.UUID
	MembershipPlanId uuid.UUID
	StartDate        time.Time
	Status           string
	RenewalDate      *time.Time
}
