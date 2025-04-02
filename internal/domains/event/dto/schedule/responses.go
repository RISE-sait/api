package schedule

import (
	values "api/internal/domains/event/values"
	"github.com/google/uuid"
	"time"
)

//goland:noinspection GoNameStartsWithPackageName
type (
	ProgramInfo struct {
		ID   uuid.UUID `json:"id"`
		Name string    `json:"name"`
		Type string    `json:"type"`
	}

	LocationInfo struct {
		ID      uuid.UUID `json:"id"`
		Name    string    `json:"name"`
		Address string    `json:"address"`
	}

	TeamInfo struct {
		ID   uuid.UUID `json:"id"`
		Name string    `json:"name"`
	}

	ScheduleResponseDto struct {
		ID                uuid.UUID    `json:"id"`
		Program           *ProgramInfo `json:"program,omitempty"`
		Location          LocationInfo `json:"location"`
		Team              *TeamInfo    `json:"team,omitempty"`
		RecurrenceStartAt string       `json:"recurrence_start_at"`
		RecurrenceEndAt   *string      `json:"recurrence_end_at,omitempty"`
		SessionStart      string       `json:"session_start_at"`
		SessionEnd        string       `json:"session_end_at"`
		Day               string       `json:"day"`
	}
)

func NewScheduleResponseDto(schedule values.ReadScheduleValues) ScheduleResponseDto {

	eventStartTimeStr := ""
	if schedule.EventStartTime.Time != "" {
		if t, err := time.Parse("15:04:05-07:00", schedule.EventStartTime.Time); err == nil {
			eventStartTimeStr = t.Format("15:04:05-07:00") // or your preferred format
		}
	}

	eventEndTimeStr := ""
	if schedule.EventEndTime.Time != "" {
		if t, err := time.Parse("15:04:05-07:00", schedule.EventEndTime.Time); err == nil {
			eventEndTimeStr = t.Format("15:04:05-07:00") // or your preferred format
		}
	}

	response := ScheduleResponseDto{
		ID: schedule.ID,
		Location: LocationInfo{
			ID:      schedule.ReadScheduleLocationValues.ID,
			Name:    schedule.ReadScheduleLocationValues.Name,
			Address: schedule.ReadScheduleLocationValues.Address,
		},
		RecurrenceStartAt: schedule.RecurrenceStartAt.String(),
		SessionStart:      eventStartTimeStr,
		SessionEnd:        eventEndTimeStr,
		Day:               schedule.Day,
	}

	if schedule.RecurrenceEndAt != nil {
		end := schedule.RecurrenceEndAt.String()
		response.RecurrenceEndAt = &end
	}

	if schedule.ReadScheduleProgramValues != nil {
		response.Program = &ProgramInfo{
			ID:   schedule.ReadScheduleProgramValues.ID,
			Name: schedule.ReadScheduleProgramValues.Name,
			Type: schedule.ReadScheduleProgramValues.Type,
		}
	}

	if schedule.ReadScheduleTeamValues != nil {
		response.Team = &TeamInfo{
			ID:   schedule.ReadScheduleTeamValues.ID,
			Name: schedule.ReadScheduleTeamValues.Name,
		}
	}

	return response
}
