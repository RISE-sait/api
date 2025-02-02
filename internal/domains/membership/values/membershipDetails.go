package values

import (
	"time"
)

type MembershipDetails struct {
	Name        string
	Description string
	StartDate   time.Time
	EndDate     time.Time
}
