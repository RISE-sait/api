package event

import (
	db "api/internal/domains/event/persistence/sqlc/generated"
	values "api/internal/domains/event/values"
	errLib "api/internal/libs/errors"
	"context"
	"errors"
	"fmt"
	"github.com/lib/pq"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
)

type Repository struct {
	Queries *db.Queries
}

func NewEventsRepository(dbQueries *db.Queries) *Repository {
	return &Repository{
		Queries: dbQueries,
	}
}

func (r *Repository) CreateEvent(c context.Context, eventDetails values.CreateEventValues) (values.ReadEventValues, *errLib.CommonError) {

	var createdEvent values.ReadEventValues

	dbParams := db.CreateEventParams{
		EventStartAt: eventDetails.EventStartAt,
		EventEndAt:   eventDetails.EventEndAt,
		GameID: uuid.NullUUID{
			UUID:  eventDetails.GameID,
			Valid: eventDetails.GameID != uuid.Nil,
		},
		PracticeID: uuid.NullUUID{
			UUID:  eventDetails.PracticeID,
			Valid: eventDetails.PracticeID != uuid.Nil,
		},
		CourseID: uuid.NullUUID{
			UUID:  eventDetails.CourseID,
			Valid: eventDetails.CourseID != uuid.Nil,
		},
		LocationID: eventDetails.LocationID,
	}

	eventDb, err := r.Queries.CreateEvent(c, dbParams)

	if err != nil {

		var pqErr *pq.Error
		if errors.As(err, &pqErr) {

			foreignKeyErrors := map[string]string{
				"fk_practice": "The referenced practice doesn't exist",
				"fk_game":     "The referenced game doesn't exist",
				"fk_location": "The referenced location doesn't exist",
				"fk_course":   "The referenced course doesn't exist",
			}

			if msg, found := foreignKeyErrors[pqErr.Constraint]; found {
				return createdEvent, errLib.New(msg, http.StatusBadRequest)
			}

			switch pqErr.Constraint {
			case "check_end_time":
				return createdEvent, errLib.New(pqErr.Message, http.StatusBadRequest)
			case "check_session_times", "check_event_times":
				return createdEvent, errLib.New("End time/date must be after Begin time/date", http.StatusBadRequest)
			}

			if strings.Contains(pqErr.Message, "overlaps") {
				return createdEvent, errLib.New(pqErr.Message, http.StatusBadRequest)
			}

			log.Println(fmt.Sprintf("Error creating event: %v", pqErr.Error()))
			return createdEvent, errLib.New("Internal db error", http.StatusInternalServerError)
		}

		log.Printf("Failed to create eventDetails: %+v. Error: %v", eventDetails, err.Error())
		return createdEvent, errLib.New("Internal server error", http.StatusInternalServerError)
	}

	createdEvent = values.ReadEventValues{
		ID:        eventDb.ID,
		CreatedAt: eventDb.CreatedAt,
		UpdatedAt: eventDb.UpdatedAt,
	}

	if eventDb.PracticeID.Valid {
		createdEvent.PracticeID = eventDb.PracticeID.UUID
	}

	if eventDb.CourseID.Valid {
		createdEvent.CourseID = eventDb.CourseID.UUID
	}

	if eventDb.GameID.Valid {
		createdEvent.GameID = eventDb.CourseID.UUID
	}

	return createdEvent, nil
}

func (r *Repository) GetEvents(ctx context.Context, after, before time.Time, courseID, practiceID, gameID, locationID uuid.UUID) ([]values.ReadEventValues, *errLib.CommonError) {

	dbEvents, err := r.Queries.GetEvents(ctx, db.GetEventsParams{
		After:  after,
		Before: before,
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
				EventStartAt: dbEvent.EventStartAt,
				EventEndAt:   dbEvent.EventEndAt,
				PracticeID:   dbEvent.PracticeID.UUID,
				CourseID:     dbEvent.CourseID.UUID,
				GameID:       dbEvent.GameID.UUID,
				LocationID:   dbEvent.LocationID,
			},
		}

		events[i] = event

	}

	return events, nil
}

func (r *Repository) UpdateEvent(c context.Context, event values.UpdateEventValues) (values.ReadEventValues, *errLib.CommonError) {

	dbEventParams := db.UpdateEventParams{
		EventStartAt: event.EventStartAt,
		EventEndAt:   event.EventEndAt,
		LocationID:   event.LocationID,
		PracticeID:   uuid.NullUUID{UUID: event.PracticeID, Valid: event.PracticeID != uuid.Nil},
		CourseID:     uuid.NullUUID{UUID: event.CourseID, Valid: event.CourseID != uuid.Nil},
		GameID:       uuid.NullUUID{UUID: event.GameID, Valid: event.GameID != uuid.Nil},
		ID:           event.ID,
	}

	dbEvent, err := r.Queries.UpdateEvent(c, dbEventParams)

	if err != nil {
		log.Printf("Failed to update event: %+v. Error: %v", event, err.Error())
		return values.ReadEventValues{}, errLib.New("Internal server error", http.StatusInternalServerError)
	}

	updatedEvent := values.ReadEventValues{
		ID:        dbEvent.ID,
		CreatedAt: dbEvent.CreatedAt,
		UpdatedAt: dbEvent.UpdatedAt,
		Details: values.Details{
			LocationID:   dbEvent.LocationID,
			EventStartAt: dbEvent.EventStartAt,
			EventEndAt:   dbEvent.EventEndAt,
			PracticeID:   dbEvent.PracticeID.UUID,
			CourseID:     dbEvent.CourseID.UUID,
			GameID:       dbEvent.GameID.UUID,
		},
	}

	return updatedEvent, nil

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
			LocationID:   dbEvent.LocationID,
			EventStartAt: dbEvent.EventStartAt,
			EventEndAt:   dbEvent.EventEndAt,
			PracticeID:   dbEvent.PracticeID.UUID,
			CourseID:     dbEvent.CourseID.UUID,
			GameID:       dbEvent.GameID.UUID,
		},
	}

	return event, nil
}
