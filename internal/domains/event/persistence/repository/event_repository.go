package event

import (
	db "api/internal/domains/event/persistence/sqlc/generated"
	values "api/internal/domains/event/values"
	errLib "api/internal/libs/errors"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"log"
	"net/http"
	"strings"
	"time"
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
		CourseID: uuid.NullUUID{
			UUID:  eventDetails.CourseID,
			Valid: eventDetails.CourseID != uuid.Nil,
		},
		PracticeID: uuid.NullUUID{
			UUID:  eventDetails.PracticeID,
			Valid: eventDetails.PracticeID != uuid.Nil,
		},
		GameID: uuid.NullUUID{
			UUID:  eventDetails.GameID,
			Valid: eventDetails.GameID != uuid.Nil,
		},
	}

	if err := r.Queries.CreateEvent(c, dbParams); err != nil {

		var pqErr *pq.Error
		if errors.As(err, &pqErr) {

			foreignKeyErrors := map[string]string{
				"fk_practice": "The referenced practice doesn't exist",
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

func (r *Repository) GetEvents(ctx context.Context, courseID, practiceID, gameID, locationID uuid.UUID, before, after time.Time) ([]values.ReadEventValues, *errLib.CommonError) {

	dbEvents, err := r.Queries.GetEvents(ctx, db.GetEventsParams{
		CourseID: uuid.NullUUID{
			UUID:  courseID,
			Valid: courseID != uuid.Nil,
		},
		GameID: uuid.NullUUID{
			UUID:  gameID,
			Valid: gameID != uuid.Nil,
		},
		PracticeID: uuid.NullUUID{
			UUID:  practiceID,
			Valid: practiceID != uuid.Nil,
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
			ReadDetails: values.ReadDetails{
				Day:             string(dbEvent.Day),
				ProgramStartAt:  dbEvent.ProgramStartAt,
				ProgramEndAt:    dbEvent.ProgramEndAt,
				EventStartTime:  dbEvent.EventStartTime,
				EventEndTime:    dbEvent.EventEndTime,
				GameID:          dbEvent.GameID.UUID,
				GameName:        dbEvent.GameName.String,
				PracticeID:      dbEvent.PracticeID.UUID,
				PracticeName:    dbEvent.PracticeName.String,
				CourseID:        dbEvent.CourseID.UUID,
				CourseName:      dbEvent.CourseName.String,
				LocationID:      dbEvent.LocationID.UUID,
				LocationName:    dbEvent.LocationName.String,
				LocationAddress: dbEvent.Address.String,
			},
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
		PracticeID:     uuid.NullUUID{UUID: event.PracticeID, Valid: event.PracticeID != uuid.Nil},
		CourseID:       uuid.NullUUID{UUID: event.CourseID, Valid: event.CourseID != uuid.Nil},
		GameID:         uuid.NullUUID{UUID: event.GameID, Valid: event.GameID != uuid.Nil},
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
		ReadDetails: values.ReadDetails{
			Day:             string(dbEvent.Day),
			ProgramStartAt:  dbEvent.ProgramStartAt,
			ProgramEndAt:    dbEvent.ProgramEndAt,
			EventStartTime:  dbEvent.EventStartTime,
			EventEndTime:    dbEvent.EventEndTime,
			GameID:          dbEvent.GameID.UUID,
			GameName:        dbEvent.GameName.String,
			PracticeID:      dbEvent.PracticeID.UUID,
			PracticeName:    dbEvent.PracticeName.String,
			CourseID:        dbEvent.CourseID.UUID,
			CourseName:      dbEvent.CourseName.String,
			LocationID:      dbEvent.LocationID.UUID,
			LocationName:    dbEvent.LocationName.String,
			LocationAddress: dbEvent.Address.String,
		},
	}

	return event, nil
}
