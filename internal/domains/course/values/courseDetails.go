package values

import "time"

type CourseDetails struct {
	Name        string
	Description string
	StartDate   time.Time
	EndDate     time.Time
}
