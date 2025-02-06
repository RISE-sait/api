package dto

import (
	"api/internal/domains/schedule/values"
	errLib "api/internal/libs/errors"
	"api/internal/libs/validators"
	"log"
	"time"

	"github.com/google/uuid"
)

type ScheduleRequestDto struct {
	BeginTime  string    `json:"begin_time" validate:"required"`
	EndTime    string    `json:"end_time" validate:"required"`
	CourseID   uuid.UUID `json:"course_id" validate:"required"`
	FacilityID uuid.UUID `json:"facility_id" validate:"required"`
	Day        string    `json:"day" validate:"required"`
}

func (dto *ScheduleRequestDto) validate() (time.Time, time.Time, *errLib.CommonError) {
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

func (dto *ScheduleRequestDto) ToScheduleDetails() (*values.ScheduleDetails, *errLib.CommonError) {

	beginTime, endTime, err := dto.validate()

	if err != nil {

		return nil, err
	}

	return &values.ScheduleDetails{

		BeginTime:  beginTime,
		EndTime:    endTime,
		CourseID:   dto.CourseID,
		FacilityID: dto.FacilityID,
		Day:        dto.Day,
	}, nil
}

func (dto *ScheduleRequestDto) ToScheduleAllFields(idStr string) (*values.ScheduleAllFields, *errLib.CommonError) {

	id, err := validators.ParseUUID(idStr)

	if err != nil {
		return nil, err
	}

	beginTime, endTime, err := dto.validate()

	if err != nil {

		log.Println("Error: ", err)
		return nil, err
	}

	return &values.ScheduleAllFields{
		ID: id,
		ScheduleDetails: values.ScheduleDetails{

			BeginTime:  beginTime,
			EndTime:    endTime,
			CourseID:   dto.CourseID,
			FacilityID: dto.FacilityID,
			Day:        dto.Day,
		},
	}, nil
}
