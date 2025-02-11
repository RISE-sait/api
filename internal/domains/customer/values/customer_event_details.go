package values

import "time"

type CustomerEventDetails struct {
	CheckedInAt *time.Time
	IsCancelled bool
}
