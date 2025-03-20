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
