package values

import "time"

type CourseCreate struct {
	Name        string
	Description string
	StartDate   time.Time
	EndDate     time.Time
}
