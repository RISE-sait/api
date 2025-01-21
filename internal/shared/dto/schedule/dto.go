package dto

import (
	"fmt"

	db "api/sqlc"
	"time"

	"github.com/google/uuid"
)

type CreateScheduleRequest struct {
	BeginDatetime time.Time `json:"begin_datetime" validate:"required"`
	EndDatetime   time.Time `json:"end_datetime" validate:"required,enddate"`
	CourseID      uuid.UUID `json:"course_id"`
	FacilityID    uuid.UUID `json:"facility_id" validate:"required"`
	Day           int       `json:"day" validate:"required,day"`
}

func (r *CreateScheduleRequest) ToDBParams() *db.CreateScheduleParams {

	dbParams := db.CreateScheduleParams{

		BeginDatetime: r.BeginDatetime,
		EndDatetime:   r.EndDatetime,
		CourseID: uuid.NullUUID{
			UUID:  r.CourseID,
			Valid: r.CourseID != uuid.Nil,
		},
		FacilityID: r.FacilityID,
		Day:        db.DayEnum(fmt.Sprint(r.Day)),
	}

	return &dbParams
}

type UpdateScheduleRequest struct {
	BeginDatetime time.Time `json:"begin_datetime" validate:"required"`
	EndDatetime   time.Time `json:"end_datetime" validate:"required,enddate"`
	CourseID      uuid.UUID `json:"course_id"`
	FacilityID    uuid.UUID `json:"facility_id"`
	Day           int32     `json:"day" validate:"required,day"`
	ID            uuid.UUID `json:"id" validate:"required"`
}

func (r *UpdateScheduleRequest) ToDBParams() *db.UpdateScheduleParams {

	dbParams := db.UpdateScheduleParams{

		BeginDatetime: r.BeginDatetime,
		EndDatetime:   r.EndDatetime,
		CourseID: uuid.NullUUID{
			UUID:  r.CourseID,
			Valid: r.CourseID != uuid.Nil,
		},
		FacilityID: r.FacilityID,
		Day:        db.DayEnum(fmt.Sprint(r.Day)),
	}

	return &dbParams
}
