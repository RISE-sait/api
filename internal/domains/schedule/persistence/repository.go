package persistence

import (
	"api/cmd/server/di"
	db "api/internal/domains/schedule/persistence/sqlc/generated"
	"api/internal/domains/schedule/values"
	errLib "api/internal/libs/errors"
	"database/sql"
	"errors"

	"context"
	"log"
	"net/http"

	"github.com/google/uuid"
)

type SchedulesRepository struct {
	Queries *db.Queries
}

func NewScheduleRepository(container *di.Container) *SchedulesRepository {
	return &SchedulesRepository{
		Queries: container.Queries.ScheduleDb,
	}
}

func (r *SchedulesRepository) CreateSchedule(c context.Context, schedule *values.ScheduleDetails) *errLib.CommonError {

	dbParams := db.CreateScheduleParams{
		BeginDatetime: schedule.BeginDatetime,
		EndDatetime:   schedule.EndDatetime,
		CourseID: uuid.NullUUID{
			UUID:  schedule.CourseID,
			Valid: schedule.CourseID != uuid.Nil,
		},
		FacilityID: schedule.FacilityID,
		Day:        db.DayEnum(schedule.Day),
	}

	row, err := r.Queries.CreateSchedule(c, dbParams)

	if err != nil {
		log.Printf("Failed to create schedule: %+v. Error: %v", schedule, err.Error())
		return errLib.New("Internal server error", http.StatusInternalServerError)
	}

	if row == 0 {
		return errLib.New("Course or facility not found", http.StatusNotFound)
	}

	return nil
}

func (r *SchedulesRepository) GetSchedules(c context.Context, fields *values.ScheduleDetails) ([]values.ScheduleAllFields, *errLib.CommonError) {

	dbParams := db.GetSchedulesParams{
		BeginDatetime: fields.BeginDatetime,
		EndDatetime:   fields.EndDatetime,
		FacilityID:    fields.FacilityID,
		CourseID: uuid.NullUUID{
			UUID:  fields.CourseID,
			Valid: fields.CourseID != uuid.Nil,
		},
	}

	dbSchedules, err := r.Queries.GetSchedules(c, dbParams)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errLib.New("No schedule found", http.StatusNotFound)
		}
		return nil, errLib.New("Internal server error", http.StatusInternalServerError)
	}

	schedules := make([]values.ScheduleAllFields, len(dbSchedules))
	for i, dbSchedule := range dbSchedules {
		schedules[i] = values.ScheduleAllFields{
			ID: dbSchedule.ID,
			ScheduleDetails: values.ScheduleDetails{
				BeginDatetime: dbSchedule.BeginDatetime,
				EndDatetime:   dbSchedule.EndDatetime,
				CourseID:      dbSchedule.CourseID.UUID,
				FacilityID:    dbSchedule.FacilityID,
				Day:           string(dbSchedule.Day),
			},
		}
	}

	return schedules, nil
}

func (r *SchedulesRepository) UpdateSchedule(c context.Context, schedule *values.ScheduleAllFields) *errLib.CommonError {
	dbMembershipParams := db.UpdateScheduleParams{
		BeginDatetime: schedule.BeginDatetime,
		EndDatetime:   schedule.EndDatetime,
		CourseID: uuid.NullUUID{
			UUID:  schedule.CourseID,
			Valid: schedule.CourseID != uuid.Nil,
		},
		FacilityID: schedule.FacilityID,
		Day:        db.DayEnum(schedule.Day),
		ID:         schedule.ID,
	}

	row, err := r.Queries.UpdateSchedule(c, dbMembershipParams)

	if err != nil {
		log.Printf("Failed to update schedule: %+v. Error: %v", schedule, err.Error())
		return errLib.New("Internal server error", http.StatusInternalServerError)
	}

	if row == 0 {
		return errLib.New("Course or facility not found", http.StatusNotFound)
	}
	return nil
}

func (r *SchedulesRepository) DeleteSchedule(c context.Context, id uuid.UUID) *errLib.CommonError {
	row, err := r.Queries.DeleteSchedule(c, id)

	if err != nil {
		log.Printf("Failed to delete schedule with ID: %s. Error: %s", id, err.Error())
		return errLib.New("Internal server error", http.StatusInternalServerError)
	}

	if row == 0 {
		return errLib.New("Schedule not found", http.StatusNotFound)
	}

	return nil
}
