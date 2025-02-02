package values

import (
	"github.com/google/uuid"
)

type MembershipPlanUpdate struct {
	ID uuid.UUID
	MembershipPlanRequest
}
