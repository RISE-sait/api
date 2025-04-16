package haircut_event

import (
	"api/internal/di"
	values "api/internal/domains/haircut/event"
	db "api/internal/domains/haircut/event/persistence/sqlc/generated"
	service "api/internal/domains/haircut/haircut_service/persistence"
	errLib "api/internal/libs/errors"
	"context"
	"database/sql"
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
	Queries     *db.Queries
	ServiceRepo *service.BarberServiceRepository
}

func NewEventsRepository(container *di.Container) *Repository {
	return &Repository{
		Queries:     container.Queries.HaircutEventDb,
		ServiceRepo: service.NewBarberServiceRepository(container),
	}
}

func (r *Repository) CreateEvent(ctx context.Context, eventDetails values.CreateEventValues) (values.EventReadValues, *errLib.CommonError) {

	var response values.EventReadValues

	availableServices, err := r.ServiceRepo.GetBarberServices(ctx)

	if err != nil {
		return response, err
	}

	serviceNames := make([]string, len(availableServices))

	dbParams := db.CreateHaircutEventParams{
		BeginDateTime: eventDetails.BeginDateTime,
		EndDateTime:   eventDetails.EndDateTime,
		BarberID:      eventDetails.BarberID,
		CustomerID:    eventDetails.CustomerID,
	}

	for _, service := range availableServices {
		if service.HaircutName == eventDetails.ServiceName {
			dbParams.ServiceTypeID = service.ServiceTypeID
			break
		}
	}

	if dbParams.ServiceTypeID == uuid.Nil {

		for i, service := range availableServices {
			serviceNames[i] = service.HaircutName
		}

		// join the slice into a string not using values.JoinString
		return response, errLib.New(
			fmt.Sprintf("Service '%s' not found. Available services: %s",
				eventDetails.ServiceName,
				strings.Join(serviceNames, ", ")),
			http.StatusBadRequest)
	}

	eventDb, dbErr := r.Queries.CreateHaircutEvent(ctx, dbParams)

	if dbErr != nil {

		var pqErr *pq.Error
		if errors.As(dbErr, &pqErr) {

			constraintErrors := map[string]struct {
				Message string
				Status  int
			}{
				"fk_barber": {
					Message: "Barber with the associated ID doesn't exist",
					Status:  http.StatusNotFound,
				},
				"fk_customer": {
					Message: "Customer with the associated ID doesn't exist",
					Status:  http.StatusNotFound,
				},
				"fk_service_type": {
					Message: "Service with the associated ID doesn't exist",
					Status:  http.StatusNotFound,
				},
				"check_end_time": {
					Message: "end_time must be after start_time",
					Status:  http.StatusBadRequest,
				},
				"unique_schedule": {
					Message: "An event at this schedule overlaps with an existing event",
					Status:  http.StatusConflict,
				},
			}

			if errInfo, found := constraintErrors[pqErr.Constraint]; found {
				return response, errLib.New(errInfo.Message, errInfo.Status)
			}
		}

		log.Printf("Failed to create eventDetails: %+v. Error: %v", eventDetails, dbErr.Error())
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
		BarberName:   eventDb.BarberName,
		CustomerName: eventDb.CustomerName,
		CreatedAt:    eventDb.CreatedAt,
		UpdatedAt:    eventDb.UpdatedAt,
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

//func (r *Service) UpdateEvent(c context.Context, event values.UpdateEventValues) (values.EventReadValues, *errLib.CommonError) {
//	dbEventParams := db.UpdateEventParams{
//		BeginDateTime: event.BeginDateTime,
//		EndDateTime:   event.EndDateTime,
//		BarberID:      event.BarberID,
//		CustomerID:    event.CustomerID,
//		ID:            event.ID,
//	}
//
//	dbEvent, err := r.paymentQueries.UpdateEvent(c, dbEventParams)
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
