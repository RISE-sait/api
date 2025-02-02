package dto

import (
	"api/internal/domains/schedule/values"
	errLib "api/internal/libs/errors"
	"api/internal/libs/validators"
	"time"

	"github.com/google/uuid"
)

type ScheduleRequestDto struct {
	BeginDatetime time.Time `json:"begin_datetime" validate:"required"`
	EndDatetime   time.Time `json:"end_datetime" validate:"required,gtcsfield=BeginDatetime"`
	CourseID      uuid.UUID `json:"course_id" validate:"required"`
	FacilityID    uuid.UUID `json:"facility_id" validate:"required"`
	Day           string    `json:"day" validate:"required"`
}

func (dto *ScheduleRequestDto) validate() *errLib.CommonError {
	if err := validators.ValidateDto(dto); err != nil {
		return err
	}
	return nil
}

func (dto *ScheduleRequestDto) ToScheduleDetails() (*values.ScheduleDetails, *errLib.CommonError) {

	if err := dto.validate(); err != nil {
		return nil, err
	}

	return &values.ScheduleDetails{

		BeginDatetime: dto.BeginDatetime,
		EndDatetime:   dto.EndDatetime,
		CourseID:      dto.CourseID,
		FacilityID:    dto.FacilityID,
		Day:           dto.Day,
	}, nil
}

func (dto *ScheduleRequestDto) ToScheduleAllFields(idStr string) (*values.ScheduleAllFields, *errLib.CommonError) {

	id, err := validators.ParseUUID(idStr)

	if err != nil {
		return nil, err
	}

	if err := dto.validate(); err != nil {
		return nil, err
	}

	return &values.ScheduleAllFields{
		ID: id,
		ScheduleDetails: values.ScheduleDetails{

			BeginDatetime: dto.BeginDatetime,
			EndDatetime:   dto.EndDatetime,
			CourseID:      dto.CourseID,
			FacilityID:    dto.FacilityID,
			Day:           dto.Day,
		},
	}, nil
}
