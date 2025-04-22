package event

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"time"

	"api/internal/di"
	db "api/internal/domains/event/persistence/sqlc/generated"
	values "api/internal/domains/event/values"
	errLib "api/internal/libs/errors"

	"github.com/google/uuid"
)

type RecurrencesRepository struct {
	Queries *db.Queries
}

func NewRecurrencesRepository(container *di.Container) *RecurrencesRepository {
	return &RecurrencesRepository{
		Queries: container.Queries.EventDb,
	}
}

// create a function to map from day in string format to time.Weekday
func mapDayToWeekday(day string) (time.Weekday, *errLib.CommonError) {
	switch day {
	case "Sunday":
		return time.Sunday, nil
	case "Monday":
		return time.Monday, nil
	case "Tuesday":
		return time.Tuesday, nil
	case "Wednesday":
		return time.Wednesday, nil
	case "Thursday":
		return time.Thursday, nil
	case "Friday":
		return time.Friday, nil
	case "Saturday":
		return time.Saturday, nil
	default:
		return 0, errLib.New(fmt.Sprintf("invalid day: %s", day), http.StatusInternalServerError)
	}
}

func (r *RecurrencesRepository) GetEventsRecurrences(ctx context.Context, programTypeStr string, programID, locationID,
	userID, teamID, createdBy, updatedBy uuid.UUID, before, after time.Time,
) ([]values.ReadRecurrenceValues, *errLib.CommonError) {
	var programType db.NullProgramProgramType

	if programTypeStr == "" {
		programType.Valid = false
	} else {
		if !db.ProgramProgramType(programTypeStr).Valid() {

			validTypes := db.AllProgramProgramTypeValues()

			return nil, errLib.New(fmt.Sprintf("Invalid program type. Valid types are: %v", validTypes), http.StatusBadRequest)
		}
	}

	// Execute the query using SQLC generated function
	param := db.GetEventsRecurrenceParams{
		ProgramID:  uuid.NullUUID{UUID: programID, Valid: programID != uuid.Nil},
		LocationID: uuid.NullUUID{UUID: locationID, Valid: locationID != uuid.Nil},
		Before:     sql.NullTime{Time: before, Valid: !before.IsZero()},
		After:      sql.NullTime{Time: after, Valid: !after.IsZero()},
		UserID:     uuid.NullUUID{UUID: userID, Valid: userID != uuid.Nil},
		TeamID:     uuid.NullUUID{UUID: teamID, Valid: teamID != uuid.Nil},
		CreatedBy:  uuid.NullUUID{UUID: createdBy, Valid: createdBy != uuid.Nil},
		UpdatedBy:  uuid.NullUUID{UUID: updatedBy, Valid: updatedBy != uuid.Nil},
		Type:       programType,
	}

	rows, err := r.Queries.GetEventsRecurrence(ctx, param)
	if err != nil {
		log.Println("Failed to get events schedules from db: ", err.Error())
		return nil, errLib.New("Internal server error", http.StatusInternalServerError)
	}

	var schedules []values.ReadRecurrenceValues

	for _, row := range rows {

		day, err := mapDayToWeekday(row.DayOfWeek)
		if err != nil {
			return nil, err
		}

		schedule := values.ReadRecurrenceValues{
			ID: row.RecurrenceID.UUID,
			BaseRecurrenceValues: values.BaseRecurrenceValues{
				DayOfWeek:       time.Weekday(day),
				StartTime:       row.StartTime,
				EndTime:         row.EndTime,
				FirstOccurrence: row.FirstOccurrence,
				LastOccurrence:  row.LastOccurrence,
			},
			Location: struct {
				ID      uuid.UUID
				Name    string
				Address string
			}{
				ID:      row.LocationID,
				Name:    row.LocationName,
				Address: row.LocationAddress,
			},
			Program: struct {
				ID          uuid.UUID
				Name        string
				Description string
				Type        string
			}{
				ID:          row.ProgramID,
				Name:        row.ProgramName,
				Description: row.ProgramDescription,
				Type:        string(row.ProgramType),
			},
			EventCount: row.EventCount,
		}

		if row.TeamID.Valid && row.TeamName.Valid {
			schedule.Team = &struct {
				ID   uuid.UUID
				Name string
			}{
				ID:   row.TeamID.UUID,
				Name: row.TeamName.String,
			}
		}

		schedules = append(schedules, schedule)
	}

	return schedules, nil
}
