package dto

import (
	"time"

	"github.com/google/uuid"
)

type ScheduleResponse struct {
	ID            uuid.UUID `json:"id"`
	BeginDatetime time.Time `json:"begin_datetime"`
	EndDatetime   time.Time `json:"end_datetime"`
	CourseID      uuid.UUID `json:"course_id"`
	FacilityID    uuid.UUID `json:"facility_id" `
	Day           string    `json:"day" `
}
