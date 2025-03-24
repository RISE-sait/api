package event

import (
	db "api/internal/domains/event/persistence/sqlc/generated"
	values "api/internal/domains/event/values"
	errLib "api/internal/libs/errors"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

type Repository struct {
	Queries *db.Queries
}

func NewEventsRepository(dbQueries *db.Queries) *Repository {
	return &Repository{
		Queries: dbQueries,
	}
}

func (r *Repository) CreateEvent(c context.Context, eventDetails values.CreateEventValues) *errLib.CommonError {

	if !db.DayEnum(eventDetails.Day).Valid() {

		validDaysDbValues := db.AllDayEnumValues()

		validDays := make([]string, len(validDaysDbValues))

		for i, value := range validDaysDbValues {
			validDays[i] = string(value)
		}

		return errLib.New("Invalid day provided. Valid days are: "+strings.Join(validDays, ", "), http.StatusBadRequest)
	}

	dbParams := db.CreateEventParams{
		ProgramStartAt: eventDetails.ProgramStartAt,
		ProgramEndAt:   eventDetails.ProgramEndAt,
		EventStartTime: eventDetails.EventStartTime,
		EventEndTime:   eventDetails.EventEndTime,
		Day:            db.DayEnum(eventDetails.Day),
		LocationID: uuid.NullUUID{
			UUID:  eventDetails.LocationID,
			Valid: eventDetails.LocationID != uuid.Nil,
		},
		ProgramID: uuid.NullUUID{
			UUID:  eventDetails.ProgramID,
			Valid: eventDetails.ProgramID != uuid.Nil,
		},
	}

	if eventDetails.Capacity != nil {
		dbParams.Capacity = sql.NullInt32{
			Int32: *eventDetails.Capacity,
			Valid: true,
		}
	}

	if err := r.Queries.CreateEvent(c, dbParams); err != nil {

		var pqErr *pq.Error
		if errors.As(err, &pqErr) {

			foreignKeyErrors := map[string]string{
				"fk_program":  "The referenced program doesn't exist",
				"fk_game":     "The referenced game doesn't exist",
				"fk_location": "The referenced location doesn't exist",
				"fk_course":   "The referenced course doesn't exist",
			}

			if msg, found := foreignKeyErrors[pqErr.Constraint]; found {
				return errLib.New(msg, http.StatusBadRequest)
			}

			switch pqErr.Constraint {
			case "check_end_time":
				return errLib.New(pqErr.Message, http.StatusBadRequest)
			case "check_session_times", "check_event_times":
				return errLib.New("End time/date must be after Begin time/date", http.StatusBadRequest)
			}

			if strings.Contains(pqErr.Message, "overlaps") {
				return errLib.New(pqErr.Message, http.StatusBadRequest)
			}

			log.Println(fmt.Sprintf("Error creating event: %v", pqErr.Error()))
			return errLib.New("Internal db error", http.StatusInternalServerError)
		}

		log.Printf("Failed to create eventDetails: %+v. Error: %v", eventDetails, err.Error())
		return errLib.New("Internal server error", http.StatusInternalServerError)
	}

	return nil
}

func (r *Repository) GetEvents(ctx context.Context, programID, locationID uuid.UUID, before, after time.Time) ([]values.ReadEventValues, *errLib.CommonError) {

	dbEvents, err := r.Queries.GetEvents(ctx, db.GetEventsParams{
		ProgramID: uuid.NullUUID{
			UUID:  programID,
			Valid: programID != uuid.Nil,
		},
		LocationID: uuid.NullUUID{
			UUID:  locationID,
			Valid: locationID != uuid.Nil,
		},
		Before: sql.NullTime{
			Time:  before,
			Valid: !before.IsZero(),
		},
		After: sql.NullTime{
			Time:  after,
			Valid: !after.IsZero(),
		},
	})

	if err != nil {
		log.Println("Failed to get events: ", err.Error())
		return nil, errLib.New("Internal server error", http.StatusInternalServerError)
	}

	events := make([]values.ReadEventValues, len(dbEvents))
	for i, dbEvent := range dbEvents {

		event := values.ReadEventValues{
			ID: dbEvent.ID,
			Details: values.Details{
				Day:            string(dbEvent.Day),
				ProgramStartAt: dbEvent.ProgramStartAt,
				ProgramEndAt:   dbEvent.ProgramEndAt,
				EventStartTime: dbEvent.EventStartTime,
				EventEndTime:   dbEvent.EventEndTime,
				ProgramID:      dbEvent.ID,
				LocationID:     dbEvent.LocationID.UUID,
			},
			LocationName:    dbEvent.LocationName.String,
			LocationAddress: dbEvent.Address.String,
			ProgramName:     dbEvent.ProgramName.String,
			ProgramType:     string(dbEvent.ProgramType.ProgramProgramType),
		}

		if dbEvent.Capacity.Valid {
			event.Details.Capacity = &dbEvent.Capacity.Int32
		}

		events[i] = event

	}

	return events, nil
}

func (r *Repository) UpdateEvent(c context.Context, event values.UpdateEventValues) *errLib.CommonError {

	if !db.DayEnum(event.Day).Valid() {

		validDaysDbValues := db.AllDayEnumValues()

		validDays := make([]string, len(validDaysDbValues))

		for i, value := range validDaysDbValues {
			validDays[i] = string(value)
		}

		return errLib.New("Invalid day provided. Valid days are: "+strings.Join(validDays, ", "), http.StatusBadRequest)
	}

	dbEventParams := db.UpdateEventParams{
		ProgramStartAt: event.ProgramStartAt,
		ProgramEndAt:   event.ProgramEndAt,
		LocationID:     uuid.NullUUID{UUID: event.LocationID, Valid: event.LocationID != uuid.Nil},
		ProgramID:      uuid.NullUUID{UUID: event.ProgramID, Valid: event.ProgramID != uuid.Nil},
		EventStartTime: event.EventStartTime,
		EventEndTime:   event.EventEndTime,
		Day:            db.DayEnum(event.Day),
		ID:             event.ID,
	}

	err := r.Queries.UpdateEvent(c, dbEventParams)

	if err != nil {
		log.Printf("Failed to update event: %+v. Error: %v", event, err.Error())
		return errLib.New("Internal server error", http.StatusInternalServerError)
	}

	return nil

}

func (r *Repository) DeleteEvent(c context.Context, id uuid.UUID) *errLib.CommonError {
	err := r.Queries.DeleteEvent(c, id)

	if err != nil {
		log.Printf("Failed to delete event with HubSpotId: %s. Error: %s", id, err.Error())
		return errLib.New("Internal server error", http.StatusInternalServerError)
	}

	return nil
}

func (r *Repository) GetEvent(ctx context.Context, id uuid.UUID) (values.ReadEventValues, *errLib.CommonError) {

	dbEvent, err := r.Queries.GetEventById(ctx, id)

	if err != nil {
		log.Println("Failed to get event details: ", err.Error())
		return values.ReadEventValues{}, errLib.New("Internal server error", http.StatusInternalServerError)
	}

	event := values.ReadEventValues{
		ID:        dbEvent.ID,
		CreatedAt: dbEvent.CreatedAt,
		UpdatedAt: dbEvent.UpdatedAt,
		Details: values.Details{
			Day:            string(dbEvent.Day),
			ProgramStartAt: dbEvent.ProgramStartAt,
			ProgramEndAt:   dbEvent.ProgramEndAt,
			EventStartTime: dbEvent.EventStartTime,
			EventEndTime:   dbEvent.EventEndTime,
			LocationID:     dbEvent.LocationID.UUID,
			ProgramID:      dbEvent.ProgramID.UUID,
		},
		ProgramName:     dbEvent.ProgramName.String,
		ProgramType:     string(dbEvent.ProgramType.ProgramProgramType),
		LocationName:    dbEvent.LocationName.String,
		LocationAddress: dbEvent.Address.String,
	}

	if dbEvent.Capacity.Valid {
		event.Details.Capacity = &dbEvent.Capacity.Int32
	}

	return event, nil
}
