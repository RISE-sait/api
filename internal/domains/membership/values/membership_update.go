package values

import (
	"github.com/google/uuid"
)

type MembershipUpdate struct {
	ID uuid.UUID
	Membership
}
