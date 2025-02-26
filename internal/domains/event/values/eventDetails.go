package values

import (
	"api/internal/custom_types"
	"github.com/google/uuid"
)

type EventDetails struct {
	BeginTime  custom_types.TimeWithTimeZone
	EndTime    custom_types.TimeWithTimeZone
	PracticeID uuid.UUID
	CourseID   uuid.UUID
	LocationID uuid.UUID
	Day        string
}
