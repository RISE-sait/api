package event

import (
	entity "api/internal/domains/event/entity"
	db "api/internal/domains/event/persistence/sqlc/generated"
	values "api/internal/domains/event/values"
	errLib "api/internal/libs/errors"
	"context"
	"log"
	"net/http"
	"strings"

	"github.com/google/uuid"
)

type Repository struct {
	Queries *db.Queries
}

var _ EventsRepositoryInterface = (*Repository)(nil)

func NewEventsRepository(dbQueries *db.Queries) *Repository {
	return &Repository{
		Queries: dbQueries,
	}
}

func (r *Repository) CreateEvent(c context.Context, eventDetails *values.Details) (entity.Event, *errLib.CommonError) {

	dbParams := db.CreateEventParams{
		BeginDateTime: eventDetails.BeginDateTime,
		EndDateTime:   eventDetails.EndDateTime,
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

		if strings.Contains(err.Error(), "overlaps with an existing event") {
			return entity.Event{}, errLib.New("An event at this location on the selected day overlaps with an existing event. Please choose a different time.", http.StatusBadRequest)
		}

		log.Printf("Failed to create eventDetails: %+v. Error: %v", eventDetails, err.Error())
		return entity.Event{}, errLib.New("Internal server error", http.StatusInternalServerError)
	}

	event := entity.Event{
		ID:            eventDb.ID,
		LocationID:    eventDb.LocationID,
		BeginDateTime: eventDb.BeginDateTime,
		EndDateTime:   eventDb.EndDateTime,
	}

	if eventDb.PracticeID.Valid {
		event.PracticeID = &eventDb.PracticeID.UUID
	}

	if eventDb.CourseID.Valid {
		event.CourseID = &eventDb.CourseID.UUID
	}

	if eventDb.GameID.Valid {
		event.GameID = &eventDb.CourseID.UUID
	}

	return event, nil
}

func (r *Repository) GetEvents(ctx context.Context, courseId, locationId, practiceId, gameId *uuid.UUID) ([]entity.Event, *errLib.CommonError) {

	getEventsArgs := db.GetEventsParams{}

	if practiceId != nil {
		getEventsArgs.PracticeID = uuid.NullUUID{
			Valid: true,
			UUID:  *practiceId,
		}
	}

	if locationId != nil {
		getEventsArgs.LocationID = uuid.NullUUID{
			Valid: true,
			UUID:  *locationId,
		}
	}

	if courseId != nil {
		getEventsArgs.CourseID = uuid.NullUUID{
			Valid: true,
			UUID:  *courseId,
		}
	}

	dbEvents, err := r.Queries.GetEvents(ctx, getEventsArgs)

	if err != nil {
		log.Println("Failed to get events: ", err.Error())
		return nil, errLib.New("Internal server error", http.StatusInternalServerError)
	}

	events := make([]entity.Event, len(dbEvents))
	for i, dbEvent := range dbEvents {

		event := entity.Event{
			ID:            dbEvent.ID,
			LocationID:    dbEvent.LocationID,
			BeginDateTime: dbEvent.BeginDateTime,
			EndDateTime:   dbEvent.EndDateTime,
		}

		if dbEvent.PracticeID.Valid {
			event.PracticeID = &dbEvent.PracticeID.UUID
		}

		if dbEvent.CourseID.Valid {
			event.CourseID = &dbEvent.CourseID.UUID
		}

		if dbEvent.GameID.Valid {
			event.GameID = &dbEvent.GameID.UUID
		}

		events[i] = event

	}

	return events, nil
}

func (r *Repository) UpdateEvent(c context.Context, event *entity.Event) (*entity.Event, *errLib.CommonError) {
	dbEventParams := db.UpdateEventParams{
		BeginDateTime: event.BeginDateTime,
		EndDateTime:   event.EndDateTime,
		LocationID:    event.LocationID,
		ID:            event.ID,
	}

	if event.PracticeID != nil {
		dbEventParams.PracticeID = uuid.NullUUID{
			UUID:  *event.PracticeID,
			Valid: true,
		}
	}

	if event.CourseID != nil {
		dbEventParams.CourseID = uuid.NullUUID{
			UUID:  *event.CourseID,
			Valid: true,
		}
	}

	if event.GameID != nil {
		dbEventParams.GameID = uuid.NullUUID{
			UUID:  *event.GameID,
			Valid: true,
		}
	}

	dbEvent, err := r.Queries.UpdateEvent(c, dbEventParams)

	if err != nil {
		log.Printf("Failed to update event: %+v. Error: %v", event, err.Error())
		return nil, errLib.New("Internal server error", http.StatusInternalServerError)
	}

	updatedEvent := entity.Event{
		ID:            dbEvent.ID,
		LocationID:    dbEvent.LocationID,
		BeginDateTime: dbEvent.BeginDateTime,
		EndDateTime:   dbEvent.EndDateTime,
	}

	if dbEvent.PracticeID.Valid {
		updatedEvent.PracticeID = &dbEvent.PracticeID.UUID
	}

	if dbEvent.CourseID.Valid {
		updatedEvent.CourseID = &dbEvent.CourseID.UUID
	}

	if dbEvent.GameID.Valid {
		updatedEvent.GameID = &dbEvent.GameID.UUID
	}

	return &updatedEvent, nil

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

func (r *Repository) GetEventDetails(ctx context.Context, id uuid.UUID) (*entity.Event, *errLib.CommonError) {

	dbEvent, err := r.Queries.GetEventById(ctx, id)

	if err != nil {
		log.Println("Failed to get event details: ", err.Error())
		return nil, errLib.New("Internal server error", http.StatusInternalServerError)
	}

	event := &entity.Event{
		ID:            dbEvent.ID,
		LocationID:    dbEvent.LocationID,
		BeginDateTime: dbEvent.BeginDateTime,
		EndDateTime:   dbEvent.EndDateTime,
	}

	if dbEvent.PracticeID.Valid {
		event.PracticeID = &dbEvent.PracticeID.UUID
	}

	if dbEvent.CourseID.Valid {
		event.CourseID = &dbEvent.CourseID.UUID
	}

	if dbEvent.GameID.Valid {
		event.GameID = &dbEvent.GameID.UUID
	}

	return event, nil
}
