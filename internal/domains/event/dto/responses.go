package event

import (
	values "api/internal/domains/event/values"
	"github.com/google/uuid"
)

type ResponseDto struct {
	ID             uuid.UUID  `json:"id"`
	ProgramStartAt string     `json:"program_start_at"`
	ProgramEndAt   string     `json:"program_end_at"`
	SessionStart   string     `json:"session_start_at"`
	SessionEnd     string     `json:"session_end_at"`
	Day            string     `json:"day"`
	PracticeID     *uuid.UUID `json:"practice_id,omitempty"`
	CourseID       *uuid.UUID `json:"course_id,omitempty"`
	GameID         *uuid.UUID `json:"game_id,omitempty"`
	LocationID     *uuid.UUID `json:"location_id,omitempty"`
}

func NewEventResponse(event values.ReadEventValues) ResponseDto {
	response := ResponseDto{
		ID:             event.ID,
		ProgramStartAt: event.ProgramStartAt.String(),
		ProgramEndAt:   event.ProgramStartAt.String(),
		SessionStart:   event.SessionStartTime.Time,
		SessionEnd:     event.SessionEndTime.Time,
		Day:            event.Day,
	}

	if event.GameID != uuid.Nil {
		response.GameID = &event.GameID
	}

	if event.LocationID != uuid.Nil {
		response.LocationID = &event.LocationID
	}

	if event.PracticeID != uuid.Nil {
		response.PracticeID = &event.PracticeID
	}

	if event.CourseID != uuid.Nil {
		response.CourseID = &event.CourseID
	}

	return response
}
