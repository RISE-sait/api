package practice

import (
	"strings"
	"time"

	values "api/internal/domains/practice/values"
	errLib "api/internal/libs/errors"
	"api/internal/libs/validators"

	"github.com/google/uuid"
)

type RequestDto struct {
	TeamID     uuid.UUID   `json:"team_id" validate:"required"`
	StartTime  time.Time   `json:"start_time" validate:"required"`
	EndTime    *time.Time  `json:"end_time"`
	LocationID uuid.UUID   `json:"location_id" validate:"required"`
	CourtID    uuid.UUID   `json:"court_id" validate:"required"`
	Status     string      `json:"status" validate:"oneof=scheduled completed canceled"`
	BookedBy   *uuid.UUID  `json:"booked_by"`
}

func (dto *RequestDto) ToCreateValue() (values.CreatePracticeValue, *errLib.CommonError) {
	if err := validators.ValidateDto(dto); err != nil {
		return values.CreatePracticeValue{}, err
	}
	return values.CreatePracticeValue{
		TeamID:     dto.TeamID,
		StartTime:  dto.StartTime,
		EndTime:    dto.EndTime,
		LocationID: dto.LocationID,
		CourtID:    dto.CourtID,
		Status:     dto.Status,
		BookedBy:   dto.BookedBy,
	}, nil
}

func (dto *RequestDto) ToUpdateValue(idStr string) (values.UpdatePracticeValue, *errLib.CommonError) {
	id, err := validators.ParseUUID(idStr)
	if err != nil {
		return values.UpdatePracticeValue{}, err
	}
	if err = validators.ValidateDto(dto); err != nil {
		return values.UpdatePracticeValue{}, err
	}
	return values.UpdatePracticeValue{
		ID: id,
		CreatePracticeValue: values.CreatePracticeValue{
			TeamID:     dto.TeamID,
			StartTime:  dto.StartTime,
			EndTime:    dto.EndTime,
			LocationID: dto.LocationID,
			CourtID:    dto.CourtID,
			Status:     dto.Status,
			BookedBy:   dto.BookedBy,
		},
	}, nil
}

type RecurrenceRequestDto struct {
	Day               string    `json:"day"`
	RecurrenceStartAt string    `json:"recurrence_start_at" validate:"required"`
	RecurrenceEndAt   string    `json:"recurrence_end_at" validate:"required"`
	PracticeStartTime string    `json:"practice_start_at" validate:"required"`
	PracticeEndTime   string    `json:"practice_end_at" validate:"required"`
	TeamID            uuid.UUID `json:"team_id" validate:"required"`
	LocationID        uuid.UUID `json:"location_id" validate:"required"`
	CourtID           uuid.UUID `json:"court_id" validate:"required"`
	Status            string    `json:"status" validate:"oneof=scheduled completed canceled"`
}

func (dto *RecurrenceRequestDto) ToRecurrenceValues() (values.RecurrenceValues, *errLib.CommonError) {
	if err := validators.ValidateDto(dto); err != nil {
		return values.RecurrenceValues{}, err
	}
	start, err := validators.ParseDateTime(dto.RecurrenceStartAt)
	if err != nil {
		return values.RecurrenceValues{}, err
	}
	end, err := validators.ParseDateTime(dto.RecurrenceEndAt)
	if err != nil {
		return values.RecurrenceValues{}, err
	}
	startTime, err := validators.ParseTime(dto.PracticeStartTime)
	if err != nil {
		return values.RecurrenceValues{}, err
	}
	endTime, err := validators.ParseTime(dto.PracticeEndTime)
	if err != nil {
		return values.RecurrenceValues{}, err
	}
	weekday, err := validateWeekday(dto.Day)
	if err != nil {
		return values.RecurrenceValues{}, err
	}
	return values.RecurrenceValues{
		DayOfWeek:       weekday,
		FirstOccurrence: start,
		LastOccurrence:  end,
		StartTime:       startTime,
		EndTime:         endTime,
	}, nil
}

func validateWeekday(day string) (time.Weekday, *errLib.CommonError) {
	weekdays := map[string]time.Weekday{
		"SUNDAY":    time.Sunday,
		"MONDAY":    time.Monday,
		"TUESDAY":   time.Tuesday,
		"WEDNESDAY": time.Wednesday,
		"THURSDAY":  time.Thursday,
		"FRIDAY":    time.Friday,
		"SATURDAY":  time.Saturday,
	}
	d := strings.ToUpper(day)
	if weekday, ok := weekdays[d]; ok {
		return weekday, nil
	}
	return time.Sunday, errLib.New("invalid day", 400)
}