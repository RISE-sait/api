package values

import (
	"time"

	"github.com/google/uuid"
)

type ScheduleDetails struct {
	BeginTime  time.Time
	EndTime    time.Time
	CourseID   uuid.UUID
	FacilityID uuid.UUID
	Day        string
}
