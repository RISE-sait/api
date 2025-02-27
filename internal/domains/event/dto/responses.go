package event

import (
	entity "api/internal/domains/event/entity"
	"github.com/google/uuid"
)

type ResponseDto struct {
	ID            uuid.UUID  `json:"id"`
	BeginDateTime string     `json:"begin_time"`
	EndDateTime   string     `json:"end_time"`
	PracticeID    *uuid.UUID `json:"practice_id,omitempty"`
	CourseID      *uuid.UUID `json:"course_id,omitempty"`
	LocationID    uuid.UUID  `json:"location_id"`
}

func NewEventResponse(event entity.Event) ResponseDto {
	return ResponseDto{
		ID:            event.ID,
		BeginDateTime: event.BeginDateTime.String(),
		EndDateTime:   event.EndDateTime.String(),
		PracticeID:    event.PracticeID,
		CourseID:      event.CourseID,
		LocationID:    event.LocationID,
	}
}
