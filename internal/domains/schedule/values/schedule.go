package values

import (
	"time"

	"github.com/google/uuid"
)

type ScheduleDetails struct {
	BeginDatetime time.Time
	EndDatetime   time.Time
	CourseID      uuid.UUID
	FacilityID    uuid.UUID
	Day           string
}
