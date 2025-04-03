package event

import (
	values "api/internal/domains/event/values"
	"github.com/google/uuid"
	"strings"
)

//goland:noinspection GoNameStartsWithPackageName
type (
	ScheduleResponseDto struct {
		RecurrenceStartAt string   `json:"recurrence_start_at"`
		RecurrenceEndAt   string   `json:"recurrence_end_at"`
		SessionStart      string   `json:"session_start_at"`
		SessionEnd        string   `json:"session_end_at"`
		Day               string   `json:"day"`
		Team              *Team    `json:"team,omitempty"`
		Location          Location `json:"location"`
		Program           *Program `json:"program,omitempty"`
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

func NewScheduleResponseDto(schedule values.Schedule) ScheduleResponseDto {
	response := ScheduleResponseDto{
		RecurrenceStartAt: schedule.FirstOccurrence.String(),
		RecurrenceEndAt:   schedule.LastOccurrence.String(),
		SessionStart:      schedule.StartTime,
		SessionEnd:        schedule.EndTime,
		Day:               strings.ToUpper(schedule.DayOfWeek),
		Location: Location{
			ID:      schedule.Location.ID,
			Name:    schedule.Location.Name,
			Address: schedule.Location.Address,
		},
	}

	if schedule.Program != nil {
		response.Program = &Program{
			ID:   schedule.Program.ID,
			Name: schedule.Program.Name,
			Type: schedule.Program.Type,
		}
	}

	if schedule.Team != nil {
		response.Team = &Team{
			ID:   schedule.Team.ID,
			Name: schedule.Team.Name,
		}
	}

	return response
}
