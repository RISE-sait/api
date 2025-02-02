package values

import (
	"time"
)

type Membership struct {
	Name        string
	Description string
	StartDate   time.Time
	EndDate     time.Time
}
