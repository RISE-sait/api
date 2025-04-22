package event

import (
	"strings"

	values "api/internal/domains/event/values"

	"github.com/google/uuid"
)

//goland:noinspection GoNameStartsWithPackageName
type (
	RecurrenceResponseDto struct {
		ID                uuid.UUID `json:"id"`
		RecurrenceStartAt string    `json:"recurrence_start_at"`
		RecurrenceEndAt   string    `json:"recurrence_end_at"`
		SessionStart      string    `json:"session_start_at"`
		SessionEnd        string    `json:"session_end_at"`
		Day               string    `json:"day"`
		Team              *Team     `json:"team,omitempty"`
		Location          Location  `json:"location"`
		Program           Program   `json:"program"`
	}

	Program struct {
		ID   uuid.UUID `json:"id"`
		Name string    `json:"name"`
		Type string    `json:"type"`
	}

	Location struct {
		ID      uuid.UUID `json:"id"`
		Name    string    `json:"name"`
		Address string    `json:"address"`
	}

	Team struct {
		ID   uuid.UUID `json:"id"`
		Name string    `json:"name"`
	}
)

func NewRecurrenceResponseDto(recurrence values.ReadRecurrenceValues) RecurrenceResponseDto {
	response := RecurrenceResponseDto{
		ID:                recurrence.ID,
		RecurrenceStartAt: recurrence.FirstOccurrence.String(),
		RecurrenceEndAt:   recurrence.LastOccurrence.String(),
		SessionStart:      recurrence.StartTime,
		SessionEnd:        recurrence.EndTime,
		Day:               strings.ToUpper(recurrence.DayOfWeek.String()),
		Location: Location{
			ID:      recurrence.Location.ID,
			Name:    recurrence.Location.Name,
			Address: recurrence.Location.Address,
		},
		Program: Program{
			ID:   recurrence.Program.ID,
			Name: recurrence.Program.Name,
			Type: recurrence.Program.Type,
		},
	}

	if recurrence.Team != nil {
		response.Team = &Team{
			ID:   recurrence.Team.ID,
			Name: recurrence.Team.Name,
		}
	}

	return response
}
