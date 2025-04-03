package haircut

import (
	db "api/internal/domains/haircut/persistence/sqlc/generated"
	values "api/internal/domains/haircut/values"
	errLib "api/internal/libs/errors"
	"context"
	"database/sql"
	"errors"
	"github.com/lib/pq"
	"log"
	"net/http"
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

	var response values.EventReadValues

	dbParams := db.CreateHaircutEventParams{
		BeginDateTime: eventDetails.BeginDateTime,
		EndDateTime:   eventDetails.EndDateTime,
		BarberID:      eventDetails.BarberID,
		CustomerID:    eventDetails.CustomerID,
	}

	eventDb, err := r.Queries.CreateHaircutEvent(c, dbParams)

	if err != nil {

		var pqErr *pq.Error
		if errors.As(err, &pqErr) {

			constraintErrors := map[string]string{
				"fk_barber":              "Barber with the associated ID doesn't exist",
				"fk_customer":            "Customer with the associated ID doesn't exist",
				"check_end_time":         "end_time must be after start_time",
				"unique_barber_schedule": "An event at this schedule overlaps with an existing event. Please choose a different schedule.",
			}

			if msg, found := constraintErrors[pqErr.Constraint]; found {
				return response, errLib.New(msg, http.StatusBadRequest)
			}
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

	getEventsArgs := db.GetHaircutEventsParams{
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

	dbEvents, err := r.Queries.GetHaircutEvents(ctx, getEventsArgs)

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
				BeginDateTime: dbEvent.BeginDateTime,
				EndDateTime:   dbEvent.EndDateTime,
			},
			BarberName:   dbEvent.BarberName,
			CustomerName: dbEvent.CustomerName,
			CreatedAt:    dbEvent.CreatedAt,
			UpdatedAt:    dbEvent.UpdatedAt,
		}

		events[i] = event

	}

	return events, nil
}

//func (r *Repo) UpdateEvent(c context.Context, event values.UpdateEventValues) (values.EventReadValues, *errLib.CommonError) {
//	dbEventParams := db.UpdateEventParams{
//		BeginDateTime: event.BeginDateTime,
//		EndDateTime:   event.EndDateTime,
//		BarberID:      event.BarberID,
//		CustomerID:    event.CustomerID,
//		ID:            event.ID,
//	}
//
//	dbEvent, err := r.Queries.UpdateEvent(c, dbEventParams)
//
//	if err != nil {
//		log.Printf("Failed to update event: %+v. Error: %v", event, err.Error())
//		return values.EventReadValues{}, errLib.New("Internal server error", http.StatusInternalServerError)
//	}
//
//	var updatedEvent values.EventReadValues
//
//	updatedEvent = values.EventReadValues{
//		ID: dbEvent.ID,
//		EventValuesBase: values.EventValuesBase{
//			BarberID:      dbEvent.BarberID,
//			CustomerID:    dbEvent.CustomerID,
//			BeginDateTime: dbEvent.BeginDateTime,
//			EndDateTime:   dbEvent.EndDateTime,
//		},
//		CreatedAt: dbEvent.CreatedAt,
//		UpdatedAt: dbEvent.UpdatedAt,
//	}
//
//	return updatedEvent, nil
//
//}

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
			CustomerID:    dbEvent.CustomerID,
			BeginDateTime: dbEvent.BeginDateTime,
			EndDateTime:   dbEvent.EndDateTime,
		},
		BarberName:   dbEvent.BarberName,
		CustomerName: dbEvent.CustomerName,
		CreatedAt:    dbEvent.CreatedAt,
		UpdatedAt:    dbEvent.UpdatedAt,
	}

	return event, nil
}
