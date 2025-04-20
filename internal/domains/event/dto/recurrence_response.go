package event

import (
	values "api/internal/domains/event/values"
	"strings"

	"github.com/google/uuid"
)

//goland:noinspection GoNameStartsWithPackageName
type (
	RecurrenceResponseDto struct {
		RecurrenceStartAt string   `json:"recurrence_start_at"`
		RecurrenceEndAt   string   `json:"recurrence_end_at"`
		SessionStart      string   `json:"session_start_at"`
		SessionEnd        string   `json:"session_end_at"`
		Day               string   `json:"day"`
		Team              *Team    `json:"team,omitempty"`
		Location          Location `json:"location"`
		Program           Program  `json:"program"`
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

func NewRecurrenceResponseDto(schedule values.ReadRecurrenceValues) RecurrenceResponseDto {
	response := RecurrenceResponseDto{
		RecurrenceStartAt: schedule.FirstOccurrence.String(),
		RecurrenceEndAt:   schedule.LastOccurrence.String(),
		SessionStart:      schedule.StartTime,
		SessionEnd:        schedule.EndTime,
		Day:               strings.ToUpper(schedule.DayOfWeek.String()),
		Location: Location{
			ID:      schedule.Location.ID,
			Name:    schedule.Location.Name,
			Address: schedule.Location.Address,
		},
		Program: Program{
			ID:   schedule.Program.ID,
			Name: schedule.Program.Name,
			Type: schedule.Program.Type,
		},
	}

	if schedule.Team != nil {
		response.Team = &Team{
			ID:   schedule.Team.ID,
			Name: schedule.Team.Name,
		}
	}

	return response
}
