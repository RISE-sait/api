package values

import (
	"time"
)

type MembershipCreate struct {
	Name        string
	Description string
	StartDate   time.Time
	EndDate     time.Time
}
