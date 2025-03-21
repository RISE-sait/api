package haircut

import (
	db "api/internal/domains/haircut/persistence/sqlc/generated"
	values "api/internal/domains/haircut/values"
	errLib "api/internal/libs/errors"
	"context"
	"database/sql"
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

func (r *Repository) CreateEvent(c context.Context, eventDetails values.CreateEventValues) (values.EventReadValues, *errLib.CommonError) {

	dbParams := db.CreateBarberEventParams{
		BeginDateTime: eventDetails.BeginDateTime,
		EndDateTime:   eventDetails.EndDateTime,
		BarberID:      eventDetails.BarberID,
		CustomerID:    eventDetails.CustomerID,
	}

	eventDb, err := r.Queries.CreateBarberEvent(c, dbParams)

	if err != nil {

		if strings.Contains(err.Error(), "overlaps with an existing event") {
			return values.EventReadValues{}, errLib.New("An event at this location on the selected day overlaps with an existing event. Please choose a different time.", http.StatusBadRequest)
		}

		log.Printf("Failed to create eventDetails: %+v. Error: %v", eventDetails, err.Error())
		return values.EventReadValues{}, errLib.New("Internal server error", http.StatusInternalServerError)
	}

	event := values.EventReadValues{
		ID: eventDb.ID,
		EventValuesBase: values.EventValuesBase{
			BarberID:      eventDb.BarberID,
			CustomerID:    eventDb.CustomerID,
			BeginDateTime: eventDb.BeginDateTime,
			EndDateTime:   eventDb.EndDateTime,
		},
		CreatedAt: eventDb.CreatedAt,
		UpdatedAt: eventDb.UpdatedAt,
	}

	return event, nil
}

func (r *Repository) GetEvents(ctx context.Context, barberID, customerID uuid.UUID, before, after time.Time) ([]values.EventReadValues, *errLib.CommonError) {

	getEventsArgs := db.GetBarberEventsParams{
		BarberID: uuid.NullUUID{
			UUID:  barberID,
			Valid: barberID != uuid.Nil,
		},
		CustomerID: uuid.NullUUID{
			UUID:  customerID,
			Valid: customerID != uuid.Nil,
		},
		Before: sql.NullTime{
			Time:  before,
			Valid: !before.IsZero(),
		},
		After: sql.NullTime{
			Time:  after,
			Valid: !after.IsZero(),
		},
	}

	dbEvents, err := r.Queries.GetBarberEvents(ctx, getEventsArgs)

	if err != nil {
		log.Println("Failed to get events: ", err.Error())
		return nil, errLib.New("Internal server error", http.StatusInternalServerError)
	}

	events := make([]values.EventReadValues, len(dbEvents))
	for i, dbEvent := range dbEvents {

		event := values.EventReadValues{
			ID: dbEvent.ID,
			EventValuesBase: values.EventValuesBase{
				BarberID:      dbEvent.BarberID,
				CustomerID:    dbEvent.CustomerID,
				BarberName:    dbEvent.BarberName,
				CustomerName:  dbEvent.CustomerName,
				BeginDateTime: dbEvent.BeginDateTime,
				EndDateTime:   dbEvent.EndDateTime,
			},
			CreatedAt: dbEvent.CreatedAt,
			UpdatedAt: dbEvent.UpdatedAt,
		}

		events[i] = event

	}

	return events, nil
}

func (r *Repository) UpdateEvent(c context.Context, event values.UpdateEventValues) (values.EventReadValues, *errLib.CommonError) {
	dbEventParams := db.UpdateEventParams{
		BeginDateTime: event.BeginDateTime,
		EndDateTime:   event.EndDateTime,
		BarberID:      event.BarberID,
		CustomerID:    event.CustomerID,
		ID:            event.ID,
	}

	dbEvent, err := r.Queries.UpdateEvent(c, dbEventParams)

	if err != nil {
		log.Printf("Failed to update event: %+v. Error: %v", event, err.Error())
		return values.EventReadValues{}, errLib.New("Internal server error", http.StatusInternalServerError)
	}

	var updatedEvent values.EventReadValues

	updatedEvent = values.EventReadValues{
		ID: dbEvent.ID,
		EventValuesBase: values.EventValuesBase{
			BarberID:      dbEvent.BarberID,
			CustomerID:    dbEvent.CustomerID,
			BeginDateTime: dbEvent.BeginDateTime,
			EndDateTime:   dbEvent.EndDateTime,
		},
		CreatedAt: dbEvent.CreatedAt,
		UpdatedAt: dbEvent.UpdatedAt,
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

func (r *Repository) GetEvent(ctx context.Context, id uuid.UUID) (values.EventReadValues, *errLib.CommonError) {

	dbEvent, err := r.Queries.GetEventById(ctx, id)

	if err != nil {
		log.Println("Failed to get event details: ", err.Error())
		return values.EventReadValues{}, errLib.New("Internal server error", http.StatusInternalServerError)
	}

	var event values.EventReadValues

	event = values.EventReadValues{
		ID: dbEvent.ID,
		EventValuesBase: values.EventValuesBase{
			BarberID:      dbEvent.BarberID,
			BarberName:    dbEvent.BarberName,
			CustomerName:  dbEvent.CustomerName,
			CustomerID:    dbEvent.CustomerID,
			BeginDateTime: dbEvent.BeginDateTime,
			EndDateTime:   dbEvent.EndDateTime,
		},
		CreatedAt: dbEvent.CreatedAt,
		UpdatedAt: dbEvent.UpdatedAt,
	}

	return event, nil
}
