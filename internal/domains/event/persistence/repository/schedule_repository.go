package event

import (
	"api/internal/di"
	db "api/internal/domains/event/persistence/sqlc/generated"
	values "api/internal/domains/event/values"
	errLib "api/internal/libs/errors"
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
)

type SchedulesRepository struct {
	Queries *db.Queries
}

func NewSchedulesRepository(container *di.Container) *SchedulesRepository {
	return &SchedulesRepository{
		Queries: container.Queries.EventDb,
	}
}

func (r *SchedulesRepository) GetEventsSchedules(ctx context.Context, programTypeStr string, programID, locationID, userID, teamID, createdBy, updatedBy uuid.UUID, before, after time.Time) ([]values.Schedule, *errLib.CommonError) {

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
	param := db.GetEventsSchedulesParams{
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

	rows, err := r.Queries.GetEventsSchedules(ctx, param)
	if err != nil {
		log.Println("Failed to get events schedules from db: ", err.Error())
		return nil, errLib.New("Internal server error", http.StatusInternalServerError)
	}

	var schedules []values.Schedule

	for _, row := range rows {
		schedule := values.Schedule{
			DayOfWeek: row.DayOfWeek,
			StartTime: row.StartTime,
			EndTime:   row.EndTime,
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
			EventCount:      row.EventCount,
			FirstOccurrence: row.FirstOccurrence,
			LastOccurrence:  row.LastOccurrence,
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
