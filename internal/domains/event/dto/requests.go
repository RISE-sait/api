package event

import (
	values "api/internal/domains/event/values"
	errLib "api/internal/libs/errors"
	"api/internal/libs/validators"
	"fmt"
	"github.com/google/uuid"
	"net/http"
	"strings"
	"time"
)

type CreateRequestDto struct {
	Day               string    `json:"day" example:"THURSDAY"`
	RecurrenceStartAt string    `json:"recurrence_start_at" validate:"required" example:"2023-10-05T07:00:00Z"`
	RecurrenceEndAt   string    `json:"recurrence_end_at" validate:"required" example:"2023-10-05T07:00:00Z"`
	EventStartTime    string    `json:"start_at" validate:"required" example:"23:00:00+00:00"`
	EventEndTime      string    `json:"end_at" validate:"required" example:"23:00:00+00:00"`
	ProgramID         uuid.UUID `json:"program_id" example:"f0e21457-75d4-4de6-b765-5ee13221fd72"`
	LocationID        uuid.UUID `json:"location_id" example:"0bab3927-50eb-42b3-9d6b-2350dd00a100"`
	TeamID            uuid.UUID `json:"team_id" example:"0bab3927-50eb-42b3-9d6b-2350dd00a100"`
	Capacity          int32     `json:"capacity" example:"100"`
}

type UpdateRequestDto struct {
	StartAt    string    `json:"start_at" validate:"required" example:"2023-10-05T07:00:00Z"`
	EndAt      string    `json:"end_at" validate:"required" example:"2023-10-05T07:00:00Z"`
	ProgramID  uuid.UUID `json:"program_id" example:"f0e21457-75d4-4de6-b765-5ee13221fd72"`
	LocationID uuid.UUID `json:"location_id" example:"0bab3927-50eb-42b3-9d6b-2350dd00a100"`
	TeamID     uuid.UUID `json:"team_id" example:"0bab3927-50eb-42b3-9d6b-2350dd00a100"`
	Capacity   int32     `json:"capacity" example:"100"`
}

type UpdateEventsRequestDto struct {
	OriginalRecurrenceStartAt string    `json:"original_recurrence_start_at" validate:"required" example:"2023-10-05T07:00:00Z"`
	OriginalRecurrenceEndAt   string    `json:"original_recurrence_end_at" validate:"required" example:"2023-10-05T07:00:00Z"`
	OriginalEventStartTime    string    `json:"original_event_start_at" validate:"required" example:"10:00:00+00:00"`
	OriginalEventEndTime      string    `json:"original_event_end_at" validate:"required" example:"13:00:00+00:00"`
	OriginalProgramID         uuid.UUID `json:"original_program_id" example:"f0e21457-75d4-4de6-b765-5ee13221fd72"`
	OriginalLocationID        uuid.UUID `json:"original_location_id" example:"0bab3927-50eb-42b3-9d6b-2350dd00a100"`
	OriginalTeamID            uuid.UUID `json:"original_team_id" example:"0bab3927-50eb-42b3-9d6b-2350dd00a100"`
	OriginalCapacity          int32     `json:"original_capacity" example:"100"`

	NewRecurrenceStartAt string    `json:"new_recurrence_start_at" validate:"required" example:"2023-10-05T07:00:00Z"`
	NewRecurrenceEndAt   string    `json:"new_recurrence_end_at" validate:"required" example:"2023-10-05T07:00:00Z"`
	NewEventStartTime    string    `json:"new_event_start_at" validate:"required" example:"21:00:00+00:00"`
	NewEventEndTime      string    `json:"new_event_end_at" validate:"required" example:"23:00:00+00:00"`
	NewProgramID         uuid.UUID `json:"new_program_id" example:"f0e21457-75d4-4de6-b765-5ee13221fd72"`
	NewLocationID        uuid.UUID `json:"new_location_id" example:"0bab3927-50eb-42b3-9d6b-2350dd00a100"`
	NewTeamID            uuid.UUID `json:"new_team_id" example:"0bab3927-50eb-42b3-9d6b-2350dd00a100"`
	NewCapacity          int32     `json:"new_capacity" example:"100"`
}

type DeleteRequestDto struct {
	IDs []uuid.UUID `json:"ids" validate:"required,min=1"`
}

// validate validates the request DTO and returns parsed values.
// It performs the following validations:
//   - Validates the DTO structure using validators.ValidateDto
//   - Ensures the StartAt and EndAt strings are valid date-time formats
//   - Verifies ProgramID and LocationID and TeamID are valid UUIDs
//
// Returns:
//   - startAt (time.Time): The parsed start date-time
//   - endAt (time.Time): The parsed end date-time
//   - error (*errLib.CommonError): Error information if validation fails, nil otherwise
func (dto UpdateRequestDto) validate() (time.Time, time.Time, *errLib.CommonError) {

	if err := validators.ValidateDto(&dto); err != nil {
		return time.Time{}, time.Time{}, err
	}

	startAt, err := validators.ParseDateTime(dto.StartAt)

	if err != nil {
		return time.Time{}, time.Time{}, err
	}

	endAt, err := validators.ParseDateTime(dto.EndAt)

	if err != nil {
		return time.Time{}, time.Time{}, err
	}

	return startAt, endAt, nil

}

func (dto CreateRequestDto) ToCreateEventsValues(creator uuid.UUID) (values.CreateEventsRecurrenceValues, *errLib.CommonError) {

	if err := validators.ValidateDto(&dto); err != nil {
		return values.CreateEventsRecurrenceValues{}, err
	}

	recurrenceStartAt, err := validators.ParseDateTime(dto.RecurrenceStartAt)

	if err != nil {
		return values.CreateEventsRecurrenceValues{}, err
	}

	recurrenceEndAt, err := validators.ParseDateTime(dto.RecurrenceEndAt)

	if err != nil {
		return values.CreateEventsRecurrenceValues{}, err
	}

	eventStartTime, err := validators.ParseTime(dto.EventStartTime)

	if err != nil {
		return values.CreateEventsRecurrenceValues{}, err
	}

	eventEndTime, err := validators.ParseTime(dto.EventEndTime)

	if err != nil {
		return values.CreateEventsRecurrenceValues{}, err
	}

	var day *time.Weekday

	if dto.Day != "" {
		weekDay, err := validateWeekday(dto.Day)
		if err == nil {
			return values.CreateEventsRecurrenceValues{}, err
		}
		day = &weekDay
	}

	return values.CreateEventsRecurrenceValues{
		CreatedBy:         creator,
		Day:               day,
		RecurrenceStartAt: recurrenceStartAt,
		RecurrenceEndAt:   recurrenceEndAt,
		EventStartTime:    eventStartTime,
		EventEndTime:      eventEndTime,
		ProgramID:         dto.ProgramID,
		LocationID:        dto.LocationID,
		Capacity:          dto.Capacity,
	}, nil
}

func validateWeekday(day string) (time.Weekday, *errLib.CommonError) {
	// Map of valid weekdays
	weekdays := map[string]time.Weekday{
		"SUNDAY":    time.Sunday,
		"MONDAY":    time.Monday,
		"TUESDAY":   time.Tuesday,
		"WEDNESDAY": time.Wednesday,
		"THURSDAY":  time.Thursday,
		"FRIDAY":    time.Friday,
		"SATURDAY":  time.Saturday,
	}

	// Convert input to uppercase for case-insensitive comparison
	day = strings.ToUpper(day)

	// Check if the input matches a valid weekday
	if weekday, exists := weekdays[day]; exists {
		return weekday, nil
	}

	// Return an error if the input is invalid
	errMsg := fmt.Sprintf("Invalid weekday: %s. Expected one of: SUNDAY, MONDAY, TUESDAY, WEDNESDAY, THURSDAY, FRIDAY, SATURDAY", day)
	return time.Weekday(0), errLib.New(errMsg, http.StatusBadRequest)
}

func (dto UpdateRequestDto) ToUpdateEventValues(idStr string, updater uuid.UUID) (values.UpdateEventValues, *errLib.CommonError) {
	id, err := validators.ParseUUID(idStr)
	if err != nil {
		return values.UpdateEventValues{}, err
	}

	startAt, endAt, err := dto.validate()
	if err != nil {
		return values.UpdateEventValues{}, err
	}

	v := values.UpdateEventValues{
		ID:        id,
		UpdatedBy: updater,
		Details: values.Details{
			Capacity:   dto.Capacity,
			StartAt:    startAt,
			EndAt:      endAt,
			ProgramID:  dto.ProgramID,
			LocationID: dto.LocationID,
			TeamID:     dto.TeamID,
		},
	}

	return v, nil
}

func (dto UpdateEventsRequestDto) ToUpdateEventsValues(updater uuid.UUID) (values.UpdateEventsValues, *errLib.CommonError) {

	originalRecurrenceStartAt, err := validators.ParseDateTime(dto.OriginalRecurrenceStartAt)

	if err != nil {
		return values.UpdateEventsValues{}, err
	}

	newRecurrenceStartAt, err := validators.ParseDateTime(dto.NewRecurrenceStartAt)

	if err != nil {
		return values.UpdateEventsValues{}, err
	}

	originalRecurrenceEndAt, err := validators.ParseDateTime(dto.OriginalRecurrenceEndAt)

	if err != nil {
		return values.UpdateEventsValues{}, err
	}

	newRecurrenceEndAt, err := validators.ParseDateTime(dto.NewRecurrenceEndAt)

	if err != nil {
		return values.UpdateEventsValues{}, err
	}

	originalEventStartTime, err := validators.ParseTime(dto.OriginalEventStartTime)

	if err != nil {
		return values.UpdateEventsValues{}, err
	}

	newEventStartTime, err := validators.ParseTime(dto.NewEventStartTime)

	if err != nil {
		return values.UpdateEventsValues{}, err
	}

	originalEventEndTime, err := validators.ParseTime(dto.OriginalEventEndTime)

	if err != nil {
		return values.UpdateEventsValues{}, err
	}

	newEventEndTime, err := validators.ParseTime(dto.NewEventEndTime)

	if err != nil {
		return values.UpdateEventsValues{}, err
	}

	v := values.UpdateEventsValues{
		UpdatedBy:                 updater,
		OriginalRecurrenceStartAt: originalRecurrenceStartAt,
		OriginalRecurrenceEndAt:   originalRecurrenceEndAt,
		OriginalEventStartTime:    originalEventStartTime,
		OriginalEventEndTime:      originalEventEndTime,
		OriginalProgramID:         dto.OriginalProgramID,
		OriginalLocationID:        dto.OriginalLocationID,
		OriginalTeamID:            dto.OriginalTeamID,
		OriginalCapacity:          dto.OriginalCapacity,
		NewRecurrenceStartAt:      newRecurrenceStartAt,
		NewRecurrenceEndAt:        newRecurrenceEndAt,
		NewEventStartTime:         newEventStartTime,
		NewEventEndTime:           newEventEndTime,
		NewProgramID:              dto.NewProgramID,
	}

	return v, nil
}
