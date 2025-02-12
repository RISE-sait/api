package dto

import (
	entity "api/internal/domains/events/entities"
	"api/internal/domains/events/values"
	errLib "api/internal/libs/errors"
	"api/internal/libs/validators"
	"log"
	"time"

	"github.com/google/uuid"
)

type EventRequestDto struct {
	BeginTime  string    `json:"begin_time" validate:"required"`
	EndTime    string    `json:"end_time" validate:"required"`
	CourseID   uuid.UUID `json:"course_id" validate:"required"`
	FacilityID uuid.UUID `json:"facility_id" validate:"required"`
	Day        string    `json:"day" validate:"required"`
}

func (dto *EventRequestDto) validate() (time.Time, time.Time, *errLib.CommonError) {
	if err := validators.ValidateDto(dto); err != nil {
		return time.Time{}, time.Time{}, err
	}

	beginTime, err := validators.ParseTime(dto.BeginTime)

	if err != nil {
		return time.Time{}, time.Time{}, err
	}

	endTime, err := validators.ParseTime(dto.EndTime)

	if err != nil {
		return time.Time{}, time.Time{}, err
	}

	return beginTime, endTime, nil
}

func (dto *EventRequestDto) ToEventDetails() (*values.EventDetails, *errLib.CommonError) {

	beginTime, endTime, err := dto.validate()

	if err != nil {

		return nil, err
	}

	return &values.EventDetails{

		BeginTime:  beginTime,
		EndTime:    endTime,
		CourseID:   dto.CourseID,
		FacilityID: dto.FacilityID,
		Day:        dto.Day,
	}, nil
}

func (dto *EventRequestDto) ToEvent(idStr string) (*entity.Event, *errLib.CommonError) {

	id, err := validators.ParseUUID(idStr)

	if err != nil {
		return nil, err
	}

	beginTime, endTime, err := dto.validate()

	if err != nil {

		log.Println("Error: ", err)
		return nil, err
	}

	return &entity.Event{
		ID:         id,
		BeginTime:  beginTime,
		EndTime:    endTime,
		FacilityID: dto.FacilityID,
		Day:        dto.Day,
	}, nil
}
