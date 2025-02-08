package values

import (
	"time"

	"github.com/google/uuid"
)

type EventDetails struct {
	BeginTime  time.Time
	EndTime    time.Time
	CourseID   uuid.UUID
	FacilityID uuid.UUID
	Day        string
}
