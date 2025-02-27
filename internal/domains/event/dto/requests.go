package event

import (
	entity "api/internal/domains/event/entity"
	values "api/internal/domains/event/values"
	errLib "api/internal/libs/errors"
	"api/internal/libs/validators"
	"github.com/google/uuid"
	"log"
	"time"
)

type RequestDto struct {
	BeginDateTime string    `json:"begin_time" validate:"required" example:"2023-10-05T07:00:00Z"`
	EndDateTime   string    `json:"end_time" validate:"required" example:"2023-10-05T07:00:00Z"`
	PracticeID    uuid.UUID `json:"practice_id" example:"f0e21457-75d4-4de6-b765-5ee13221fd72"`
	CourseID      uuid.UUID `json:"course_id" example:"00000000-0000-0000-0000-000000000000"`
	LocationID    uuid.UUID `json:"location_id" example:"0bab3927-50eb-42b3-9d6b-2350dd00a100"`
}

func (dto *RequestDto) validate() (time.Time, time.Time, *errLib.CommonError) {
	if err := validators.ValidateDto(dto); err != nil {
		return time.Time{}, time.Time{}, err
	}

	beginDateTime, err := validators.ParseDateTime(dto.BeginDateTime)

	if err != nil {
		return time.Time{}, time.Time{}, err
	}

	endTime, err := validators.ParseDateTime(dto.EndDateTime)

	if err != nil {
		return time.Time{}, time.Time{}, err
	}

	return beginDateTime, endTime, nil
}

func (dto *RequestDto) ToDetails() (*values.Details, *errLib.CommonError) {

	beginTime, endTime, err := dto.validate()

	if err != nil {

		return nil, err
	}

	return &values.Details{

		BeginDateTime: beginTime,
		EndDateTime:   endTime,
		PracticeID:    dto.PracticeID,
		LocationID:    dto.LocationID,
	}, nil
}

func (dto *RequestDto) ToEntity(idStr string) (*entity.Event, *errLib.CommonError) {

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
		ID:            id,
		BeginDateTime: beginTime,
		EndDateTime:   endTime,
		LocationID:    dto.LocationID,
	}, nil
}
