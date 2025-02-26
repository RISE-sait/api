package event

import (
	entity "api/internal/domains/event/entity"
	"github.com/google/uuid"
)

type ResponseDto struct {
	ID         uuid.UUID  `json:"id"`
	BeginTime  string     `json:"begin_time"`
	EndTime    string     `json:"end_time"`
	PracticeID *uuid.UUID `json:"practice_id,omitempty"`
	CourseID   *uuid.UUID `json:"course_id,omitempty"`
	LocationID uuid.UUID  `json:"location_id"`
	Day        string     `json:"day" `
}

func NewEventResponse(event entity.Event) ResponseDto {
	return ResponseDto{
		ID:         event.ID,
		BeginTime:  event.BeginTime.Time,
		EndTime:    event.EndTime.Time,
		PracticeID: event.PracticeID,
		CourseID:   event.CourseID,
		LocationID: event.LocationID,
		Day:        string(event.Day),
	}
}
