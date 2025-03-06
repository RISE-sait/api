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

	"github.com/google/uuid"
)

type Repository struct {
	Queries *db.Queries
}

var _ IEventsRepository = (*Repository)(nil)

func NewEventsRepository(dbQueries *db.Queries) *Repository {
	return &Repository{
		Queries: dbQueries,
	}
}

func (r *Repository) CreateEvent(c context.Context, eventDetails values.CreateEventValues) (values.ReadEventValues, *errLib.CommonError) {

	var createdEvent values.ReadEventValues

	if !db.DayEnum(eventDetails.Day).Valid() {

		validDays := db.AllDayEnumValues()

		return createdEvent, errLib.New(
			fmt.Sprintf("Invalid day provided. Valid days are: %v", validDays),
			http.StatusBadRequest,
		)
	}

	dbParams := db.CreateEventParams{
		Day:              db.DayEnum(eventDetails.Day),
		EventStartAt:     eventDetails.EventStartAt,
		EventEndAt:       eventDetails.EventEndAt,
		SessionStartTime: eventDetails.SessionStartTime,
		SessionEndTime:   eventDetails.SessionEndTime,
		PracticeID: uuid.NullUUID{
			UUID:  eventDetails.PracticeID,
			Valid: eventDetails.PracticeID != uuid.Nil,
		},
		CourseID: uuid.NullUUID{
			UUID:  eventDetails.CourseID,
			Valid: eventDetails.CourseID != uuid.Nil,
		},
		LocationID: uuid.NullUUID{
			UUID:  eventDetails.LocationID,
			Valid: eventDetails.LocationID != uuid.Nil,
		},
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

func (r *Repository) GetEvents(ctx context.Context, courseId, locationId, practiceId, gameId uuid.UUID) ([]values.ReadEventValues, *errLib.CommonError) {

	getEventsArgs := db.GetEventsParams{}

	getEventsArgs.PracticeID = uuid.NullUUID{
		Valid: practiceId != uuid.Nil,
		UUID:  practiceId,
	}

	// Set the LocationID, assuming it's always provided as non-null
	getEventsArgs.LocationID = uuid.NullUUID{
		Valid: locationId != uuid.Nil,
		UUID:  locationId,
	}

	// Set the CourseID if provided
	getEventsArgs.CourseID = uuid.NullUUID{
		Valid: courseId != uuid.Nil,
		UUID:  courseId,
	}

	getEventsArgs.GameID = uuid.NullUUID{
		Valid: gameId != uuid.Nil,
		UUID:  gameId,
	}

	dbEvents, err := r.Queries.GetEvents(ctx, getEventsArgs)

	if err != nil {
		log.Println("Failed to get events: ", err.Error())
		return nil, errLib.New("Internal server error", http.StatusInternalServerError)
	}

	events := make([]values.ReadEventValues, len(dbEvents))
	for i, dbEvent := range dbEvents {

		event := values.ReadEventValues{
			ID: dbEvent.ID,
			Details: values.Details{
				Day:              string(dbEvent.Day),
				EventStartAt:     dbEvent.EventStartAt,
				EventEndAt:       dbEvent.EventEndAt,
				SessionStartTime: dbEvent.SessionStartTime,
				SessionEndTime:   dbEvent.SessionEndTime,
				PracticeID:       dbEvent.PracticeID.UUID,
				CourseID:         dbEvent.CourseID.UUID,
				GameID:           dbEvent.GameID.UUID,
				LocationID:       dbEvent.LocationID.UUID,
			},
		}

		events[i] = event

	}

	return events, nil
}

func (r *Repository) UpdateEvent(c context.Context, event values.UpdateEventValues) (values.ReadEventValues, *errLib.CommonError) {

	dbEventParams := db.UpdateEventParams{
		EventStartAt:     event.EventStartAt,
		EventEndAt:       event.EventEndAt,
		SessionStartTime: event.SessionStartTime,
		SessionEndTime:   event.SessionEndTime,
		LocationID:       uuid.NullUUID{UUID: event.LocationID, Valid: event.LocationID != uuid.Nil},
		PracticeID:       uuid.NullUUID{UUID: event.PracticeID, Valid: event.PracticeID != uuid.Nil},
		CourseID:         uuid.NullUUID{UUID: event.CourseID, Valid: event.CourseID != uuid.Nil},
		GameID:           uuid.NullUUID{UUID: event.GameID, Valid: event.GameID != uuid.Nil},
		ID:               event.ID,
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
			LocationID:       dbEvent.LocationID.UUID,
			EventStartAt:     dbEvent.EventStartAt,
			EventEndAt:       dbEvent.EventEndAt,
			SessionStartTime: dbEvent.SessionStartTime,
			SessionEndTime:   dbEvent.SessionEndTime,
			PracticeID:       dbEvent.PracticeID.UUID,
			CourseID:         dbEvent.CourseID.UUID,
			GameID:           dbEvent.GameID.UUID,
		},
	}

	return updatedEvent, nil

}

func (r *Repository) DeleteEvent(c context.Context, id uuid.UUID) *errLib.CommonError {
	row, err := r.Queries.DeleteEvent(c, id)

	if err != nil {
		log.Printf("Failed to delete event with HubSpotId: %s. Error: %s", id, err.Error())
		return errLib.New("Internal server error", http.StatusInternalServerError)
	}

	if row == 0 {
		return errLib.New("Event not found", http.StatusNotFound)
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
			Day:              string(dbEvent.Day),
			LocationID:       dbEvent.LocationID.UUID,
			EventStartAt:     dbEvent.EventStartAt,
			EventEndAt:       dbEvent.EventEndAt,
			SessionStartTime: dbEvent.SessionStartTime,
			SessionEndTime:   dbEvent.SessionEndTime,
			PracticeID:       dbEvent.PracticeID.UUID,
			CourseID:         dbEvent.CourseID.UUID,
			GameID:           dbEvent.GameID.UUID,
		},
	}

	return event, nil
}
