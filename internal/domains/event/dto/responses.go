package event

import (
	values "api/internal/domains/event/values"
	"github.com/google/uuid"
)

type ResponseDto struct {
	ID              uuid.UUID  `json:"id"`
	ProgramStartAt  string     `json:"program_start_at"`
	ProgramEndAt    string     `json:"program_end_at"`
	SessionStart    string     `json:"session_start_at"`
	SessionEnd      string     `json:"session_end_at"`
	Day             string     `json:"day"`
	PracticeID      *uuid.UUID `json:"practice_id,omitempty"`
	CourseID        *uuid.UUID `json:"course_id,omitempty"`
	GameID          *uuid.UUID `json:"game_id,omitempty"`
	LocationID      *uuid.UUID `json:"location_id,omitempty"`
	PracticeName    *string    `json:"practice_name,omitempty"`
	CourseName      *string    `json:"course_name,omitempty"`
	GameName        *string    `json:"game_name,omitempty"`
	LocationName    *string    `json:"location_name,omitempty"`
	LocationAddress *string    `json:"location_address,omitempty"`
	Capacity        *int32     `json:"capacity,omitempty"`
}

func NewEventResponse(event values.ReadEventValues) ResponseDto {
	response := ResponseDto{
		ID:             event.ID,
		ProgramStartAt: event.ProgramStartAt.String(),
		ProgramEndAt:   event.ProgramStartAt.String(),
		SessionStart:   event.EventStartTime.Time,
		SessionEnd:     event.EventEndTime.Time,
		Day:            event.Day,
		Capacity:       event.Capacity,
	}

	if event.GameID != uuid.Nil && event.GameName != "" {
		response.GameID = &event.GameID
		response.GameName = &event.GameName
	}

	if event.LocationID != uuid.Nil && event.LocationName != "" && event.LocationAddress != "" {
		response.LocationID = &event.LocationID
		response.LocationName = &event.LocationName
		response.LocationAddress = &event.LocationAddress
	}

	if event.PracticeID != uuid.Nil && event.PracticeName != "" {
		response.PracticeID = &event.PracticeID
		response.PracticeName = &event.PracticeName
	}

	if event.CourseID != uuid.Nil && event.CourseName != "" {
		response.CourseID = &event.CourseID
		response.CourseName = &event.CourseName
	}

	if event.Capacity != nil {
		response.Capacity = event.Capacity
	}

	return response
}
