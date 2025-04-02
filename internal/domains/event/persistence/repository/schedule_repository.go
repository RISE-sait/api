package event

import (
	"api/internal/custom_types"
	db "api/internal/domains/event/persistence/sqlc/generated"
	values "api/internal/domains/event/values"
	errLib "api/internal/libs/errors"
	"context"
	"database/sql"
	"errors"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"log"
	"net/http"
	"strings"
)

type ScheduleRepository struct {
	Queries *db.Queries
}

func NewSchedulesRepository(dbQueries *db.Queries) *ScheduleRepository {
	return &ScheduleRepository{
		Queries: dbQueries,
	}
}

func (r *ScheduleRepository) GetSchedule(ctx context.Context, id uuid.UUID) (values.ReadScheduleValues, *errLib.CommonError) {

	schedule, dbErr := r.Queries.GetScheduleById(ctx, id)

	if dbErr != nil {

		var pqErr *pq.Error
		if errors.As(dbErr, &pqErr) {

			constraintErrors := map[string]string{
				"fk_program":            "The referenced program doesn't exist",
				"fk_location":           "The referenced location doesn't exist",
				"unique_event_time":     "An event is already scheduled at this time",
				"check_event_times":     "End time/date must be after Begin time/date",
				"event_end_after_start": "End time/date must be after Begin time/date",
			}

			if msg, found := constraintErrors[pqErr.Constraint]; found {
				return values.ReadScheduleValues{}, errLib.New(msg, http.StatusBadRequest)
			}
		}

		log.Printf("Failed to create scheduleValues: %+v. Error: %v", schedule, dbErr.Error())
		return values.ReadScheduleValues{}, errLib.New("Internal server error", http.StatusInternalServerError)
	}

	response := values.ReadScheduleValues{
		ID:                schedule.ID,
		CreatedAt:         schedule.CreatedAt,
		UpdatedAt:         schedule.UpdatedAt,
		Day:               string(schedule.Day),
		RecurrenceStartAt: schedule.RecurrenceStartAt,
		EventStartTime:    schedule.EventStartTime,
		EventEndTime:      schedule.EventEndTime,
		ReadScheduleLocationValues: values.ReadScheduleLocationValues{
			ID:      schedule.LocationID,
			Name:    schedule.LocationName,
			Address: schedule.LocationAddress,
		},
	}

	if schedule.RecurrenceEndAt.Valid {
		response.EventStartTime = custom_types.TimeWithTimeZone{
			Time: schedule.EventStartTime.Time,
		}
	}

	if schedule.ProgramID.Valid && schedule.ProgramName.Valid && schedule.ProgramType.Valid {
		response.ReadScheduleProgramValues = &values.ReadScheduleProgramValues{
			ID:   schedule.ProgramID.UUID,
			Name: schedule.ProgramName.String,
			Type: string(schedule.ProgramType.ProgramProgramType),
		}
	}

	if schedule.TeamID.Valid && schedule.TeamName.Valid {
		response.ReadScheduleTeamValues = &values.ReadScheduleTeamValues{
			ID:   schedule.TeamID.UUID,
			Name: schedule.TeamName.String,
		}
	}

	return response, nil
}

func (r *ScheduleRepository) GetSchedules(ctx context.Context, programID, locationID, userID, teamID uuid.UUID, programType string) ([]values.ReadScheduleValues, *errLib.CommonError) {

	if programType != "" && !db.DayEnum(programType).Valid() {
		validTypesDbValues := db.AllProgramProgramTypeValues()

		validTypes := make([]string, len(validTypesDbValues))

		for i, value := range validTypesDbValues {
			validTypes[i] = string(value)
		}

		return nil, errLib.New("Invalid type provided. Valid types are: "+strings.Join(validTypes, ", "), http.StatusBadRequest)
	}

	params := db.GetSchedulesParams{
		ProgramID: uuid.NullUUID{
			UUID:  programID,
			Valid: programID != uuid.Nil,
		},
		LocationID: uuid.NullUUID{
			UUID:  locationID,
			Valid: locationID != uuid.Nil,
		},
		UserID: uuid.NullUUID{
			UUID:  userID,
			Valid: userID != uuid.Nil,
		},
		TeamID: uuid.NullUUID{
			UUID:  teamID,
			Valid: teamID != uuid.Nil,
		},
		Type: db.NullProgramProgramType{
			ProgramProgramType: db.ProgramProgramType(programType),
			Valid:              programType != "",
		},
	}

	dbSchedules, dbErr := r.Queries.GetSchedules(ctx, params)

	if dbErr != nil {

		log.Printf("Failed to get schedules. Error: %v", dbErr.Error())
		return nil, errLib.New("Internal server error", http.StatusInternalServerError)
	}

	schedules := make([]values.ReadScheduleValues, len(dbSchedules))

	for i, dbSchedule := range dbSchedules {
		schedule := values.ReadScheduleValues{
			ID:                dbSchedule.ID,
			CreatedAt:         dbSchedule.CreatedAt,
			UpdatedAt:         dbSchedule.UpdatedAt,
			Day:               string(dbSchedule.Day),
			RecurrenceStartAt: dbSchedule.RecurrenceStartAt,
			EventStartTime:    dbSchedule.EventStartTime,
			EventEndTime:      dbSchedule.EventEndTime,
			ReadScheduleLocationValues: values.ReadScheduleLocationValues{
				ID:      dbSchedule.LocationID,
				Name:    dbSchedule.LocationName,
				Address: dbSchedule.LocationAddress,
			},
		}

		if dbSchedule.RecurrenceEndAt.Valid {
			schedule.RecurrenceEndAt = &dbSchedule.RecurrenceEndAt.Time
		}

		if dbSchedule.ProgramID.Valid && dbSchedule.ProgramName.Valid && dbSchedule.ProgramType.Valid {
			schedule.ReadScheduleProgramValues = &values.ReadScheduleProgramValues{
				ID:   dbSchedule.ProgramID.UUID,
				Name: dbSchedule.ProgramName.String,
				Type: string(dbSchedule.ProgramType.ProgramProgramType),
			}
		}

		if dbSchedule.TeamID.Valid && dbSchedule.TeamName.Valid {
			schedule.ReadScheduleTeamValues = &values.ReadScheduleTeamValues{
				ID:   dbSchedule.TeamID.UUID,
				Name: dbSchedule.TeamName.String,
			}
		}

		schedules[i] = schedule
	}

	return schedules, nil
}

func (r *ScheduleRepository) CreateSchedule(ctx context.Context, scheduleValues values.CreateScheduleValues) *errLib.CommonError {

	if !db.DayEnum(scheduleValues.Day).Valid() {

		validDaysDbValues := db.AllDayEnumValues()

		validDays := make([]string, len(validDaysDbValues))

		for i, value := range validDaysDbValues {
			validDays[i] = string(value)
		}

		return errLib.New("Invalid day provided. Valid days are: "+strings.Join(validDays, ", "), http.StatusBadRequest)
	}

	dbParams := db.CreateScheduleParams{
		RecurrenceStartAt: scheduleValues.RecurrenceStartAt,
		EventStartTime:    scheduleValues.EventStartTime,
		EventEndTime:      scheduleValues.EventEndTime,
		Day:               db.DayEnum(scheduleValues.Day),
		LocationID:        scheduleValues.LocationID,
		ProgramID: uuid.NullUUID{
			UUID:  scheduleValues.ProgramID,
			Valid: scheduleValues.ProgramID != uuid.Nil,
		},
	}

	if scheduleValues.RecurrenceEndAt != nil {
		dbParams.RecurrenceEndAt = sql.NullTime{
			Time:  *scheduleValues.RecurrenceEndAt,
			Valid: true,
		}
	}

	if dbErr := r.Queries.CreateSchedule(ctx, dbParams); dbErr != nil {

		var pqErr *pq.Error
		if errors.As(dbErr, &pqErr) {

			constraintErrors := map[string]string{
				"fk_program":            "The referenced program doesn't exist",
				"fk_location":           "The referenced location doesn't exist",
				"unique_event_time":     "An event is already scheduled at this time",
				"check_event_times":     "End time/date must be after Begin time/date",
				"event_end_after_start": "End time/date must be after Begin time/date",
			}

			if msg, found := constraintErrors[pqErr.Constraint]; found {
				return errLib.New(msg, http.StatusBadRequest)
			}
		}

		log.Printf("Failed to create scheduleValues: %+v. Error: %v", scheduleValues, dbErr.Error())
		return errLib.New("Internal server error", http.StatusInternalServerError)
	}

	return nil
}

func (r *ScheduleRepository) UpdateSchedule(ctx context.Context, scheduleValues values.UpdateScheduleValues) *errLib.CommonError {
	if !db.DayEnum(scheduleValues.Day).Valid() {

		validDaysDbValues := db.AllDayEnumValues()

		validDays := make([]string, len(validDaysDbValues))

		for i, value := range validDaysDbValues {
			validDays[i] = string(value)
		}

		return errLib.New("Invalid day provided. Valid days are: "+strings.Join(validDays, ", "), http.StatusBadRequest)
	}

	dbParams := db.UpdateScheduleParams{
		ID:                scheduleValues.ID,
		RecurrenceStartAt: scheduleValues.RecurrenceStartAt,
		EventStartTime:    scheduleValues.EventStartTime,
		EventEndTime:      scheduleValues.EventEndTime,
		Day:               db.DayEnum(scheduleValues.Day),
		LocationID:        scheduleValues.LocationID,
		ProgramID: uuid.NullUUID{
			UUID:  scheduleValues.ProgramID,
			Valid: scheduleValues.ProgramID != uuid.Nil,
		},
	}

	if scheduleValues.RecurrenceEndAt != nil {
		dbParams.RecurrenceEndAt = sql.NullTime{
			Time:  *scheduleValues.RecurrenceEndAt,
			Valid: true,
		}
	}

	affectedRows, dbErr := r.Queries.UpdateSchedule(ctx, dbParams)

	if dbErr != nil {

		var pqErr *pq.Error
		if errors.As(dbErr, &pqErr) {

			constraintErrors := map[string]string{
				"fk_program":            "The referenced program doesn't exist",
				"fk_location":           "The referenced location doesn't exist",
				"unique_event_time":     "An event is already scheduled at this time",
				"check_event_times":     "End time/date must be after Begin time/date",
				"event_end_after_start": "End time/date must be after Begin time/date",
			}

			if msg, found := constraintErrors[pqErr.Constraint]; found {
				return errLib.New(msg, http.StatusBadRequest)
			}
		}

		log.Printf("Failed to update scheduleValues: %+v. Error: %v", scheduleValues, dbErr.Error())
		return errLib.New("Internal server error", http.StatusInternalServerError)
	}

	if affectedRows == 0 {
		return errLib.New("Schedule not found", http.StatusNotFound)
	}

	return nil
}

func (r *ScheduleRepository) DeleteSchedule(c context.Context, id uuid.UUID) *errLib.CommonError {
	err := r.Queries.DeleteSchedule(c, id)

	if err != nil {
		log.Printf("Failed to delete event with ID: %s. Error: %s", id, err.Error())
		return errLib.New("Internal server error", http.StatusInternalServerError)
	}

	return nil
}
