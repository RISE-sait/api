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
	ProgramID       *uuid.UUID `json:"program_id,omitempty"`
	ProgramName     *string    `json:"program_name,omitempty"`
	ProgramType     *string    `json:"program_type,omitempty"`
	LocationID      *uuid.UUID `json:"location_id,omitempty"`
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

	if event.ProgramID != uuid.Nil && event.ProgramName != "" && event.ProgramType != "" {
		response.ProgramID = &event.ProgramID
		response.ProgramName = &event.ProgramName
		response.ProgramType = &event.ProgramType
	}

	if event.LocationID != uuid.Nil && event.LocationName != "" && event.LocationAddress != "" {
		response.LocationID = &event.LocationID
		response.LocationName = &event.LocationName
		response.LocationAddress = &event.LocationAddress
	}

	if event.Capacity != nil {
		response.Capacity = event.Capacity
	}

	return response
}
