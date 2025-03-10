package event

import (
	"api/internal/custom_types"
	values "api/internal/domains/event/values"
	errLib "api/internal/libs/errors"
	"api/internal/libs/validators"
	"github.com/google/uuid"
	"log"
	"time"
)

type RequestDto struct {
	Day              string    `json:"day" validate:"required" example:"THURSDAY"`
	EventStartAt     string    `json:"event_start_at" validate:"required" example:"2023-10-05T07:00:00Z"`
	EventEndAt       string    `json:"event_end_at" validate:"required" example:"2023-10-05T07:00:00Z"`
	SessionStartTime string    `json:"session_start_time" validate:"required" example:"23:00:00+00:00"`
	SessionEndTime   string    `json:"session_end_time" validate:"required" example:"23:00:00+00:00"`
	PracticeID       uuid.UUID `json:"practice_id" example:"f0e21457-75d4-4de6-b765-5ee13221fd72"`
	CourseID         uuid.UUID `json:"course_id" example:"00000000-0000-0000-0000-000000000000"`
	GameID           uuid.UUID `json:"game_id" example:"00000000-0000-0000-0000-000000000000"`
	LocationID       uuid.UUID `json:"location_id" example:"0bab3927-50eb-42b3-9d6b-2350dd00a100"`
}

// validate validates the request DTO, parses the event and session start and end times,
// and returns the parsed values. If any validation or parsing fails, an error is returned.
//
// @return eventBeginDateTime The parsed event start date and time (time.Time). This is the first return value.
// @return eventEndDateTime The parsed event end date and time (time.Time). This is the second return value.
// @return sessionBeginTime The parsed session start time with timezone (custom_types.TimeWithTimeZone). This is the third return value.
// @return sessionEndTime The parsed session end time with timezone (custom_types.TimeWithTimeZone). This is the fourth return value.
// @return An error *errLib.CommonError if any validation or parsing fails. This is the last return value.
func (dto RequestDto) validate() (time.Time, time.Time, custom_types.TimeWithTimeZone, custom_types.TimeWithTimeZone, *errLib.CommonError) {
	if err := validators.ValidateDto(&dto); err != nil {
		return time.Time{}, time.Time{}, custom_types.TimeWithTimeZone{}, custom_types.TimeWithTimeZone{}, err
	}

	eventBeginDateTime, err := validators.ParseDateTime(dto.EventStartAt)

	if err != nil {
		return time.Time{}, time.Time{}, custom_types.TimeWithTimeZone{}, custom_types.TimeWithTimeZone{}, err
	}

	eventEndDateTime, err := validators.ParseDateTime(dto.EventEndAt)

	if err != nil {
		return time.Time{}, time.Time{}, custom_types.TimeWithTimeZone{}, custom_types.TimeWithTimeZone{}, err
	}

	sessionBeginTime, err := validators.ParseTime(dto.SessionStartTime)

	if err != nil {
		return time.Time{}, time.Time{}, custom_types.TimeWithTimeZone{}, custom_types.TimeWithTimeZone{}, err
	}

	sessionEndTime, err := validators.ParseTime(dto.SessionEndTime)

	if err != nil {
		return time.Time{}, time.Time{}, custom_types.TimeWithTimeZone{}, custom_types.TimeWithTimeZone{}, err
	}

	return eventBeginDateTime, eventEndDateTime, sessionBeginTime, sessionEndTime, nil
}

func (dto RequestDto) ToCreateEventValues() (values.CreateEventValues, *errLib.CommonError) {

	eventBeginDateTime, eventEndDateTime, sessionBeginTime, sessionEndTime, err := dto.validate()

	if err != nil {
		return values.CreateEventValues{}, err
	}

	return values.CreateEventValues{
		Details: values.Details{
			EventStartAt:     eventBeginDateTime,
			EventEndAt:       eventEndDateTime,
			SessionStartTime: sessionBeginTime,
			SessionEndTime:   sessionEndTime,
			PracticeID:       dto.PracticeID,
			CourseID:         dto.CourseID, // Assuming you need to map this
			GameID:           dto.GameID,
			LocationID:       dto.LocationID,
			Day:              dto.Day,
		},
	}, nil
}

func (dto RequestDto) ToUpdateEventValues(idStr string) (values.UpdateEventValues, *errLib.CommonError) {

	id, err := validators.ParseUUID(idStr)

	if err != nil {
		return values.UpdateEventValues{}, err
	}

	eventBeginDateTime, eventEndDateTime, sessionBeginTime, sessionEndTime, err := dto.validate()

	if err != nil {

		log.Println("Error: ", err)
		return values.UpdateEventValues{}, err
	}

	return values.UpdateEventValues{
		ID: id,
		Details: values.Details{
			EventStartAt:     eventBeginDateTime,
			EventEndAt:       eventEndDateTime,
			SessionStartTime: sessionBeginTime,
			SessionEndTime:   sessionEndTime,
			PracticeID:       dto.PracticeID,
			CourseID:         dto.CourseID, // Assuming you want to keep this
			GameID:           dto.GameID,
			LocationID:       dto.LocationID,
			Day:              dto.Day,
		},
	}, nil
}
