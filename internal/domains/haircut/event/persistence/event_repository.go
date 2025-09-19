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

// GetAvailableTimeSlots returns available time slots for a barber on a specific date
func (r *Repository) GetAvailableTimeSlots(ctx context.Context, barberID uuid.UUID, date time.Time, serviceDurationMinutes int) ([]string, *errLib.CommonError) {
	// Get the day of week (0=Sunday, 6=Saturday)
	dayOfWeek := int32(date.Weekday())

	// Get barber's working hours for this day
	workingHours, err := r.Queries.GetBarberWorkingHoursForDay(ctx, db.GetBarberWorkingHoursForDayParams{
		BarberID:  barberID,
		DayOfWeek: dayOfWeek,
	})

	if err != nil {
		log.Printf("Failed to get barber working hours: %v", err)
		return nil, errLib.New("Failed to get barber availability", http.StatusInternalServerError)
	}

	if len(workingHours) == 0 {
		// Barber doesn't work on this day
		return []string{}, nil
	}

	// Get existing bookings for this date
	bookings, err := r.Queries.GetBarberBookingsForDate(ctx, db.GetBarberBookingsForDateParams{
		BarberID:      barberID,
		BeginDateTime: date,
	})

	if err != nil {
		log.Printf("Failed to get barber bookings: %v", err)
		return nil, errLib.New("Failed to get existing bookings", http.StatusInternalServerError)
	}

	var availableSlots []string
	serviceDuration := time.Duration(serviceDurationMinutes) * time.Minute

	// For each working time period
	for _, workHours := range workingHours {
		// Convert times to datetime for this date
		startTime := time.Date(date.Year(), date.Month(), date.Day(), 
			workHours.StartTime.Hour(), workHours.StartTime.Minute(), 0, 0, date.Location())
		endTime := time.Date(date.Year(), date.Month(), date.Day(), 
			workHours.EndTime.Hour(), workHours.EndTime.Minute(), 0, 0, date.Location())

		// Generate time slots every 15 minutes within working hours
		current := startTime
		for current.Add(serviceDuration).Before(endTime) || current.Add(serviceDuration).Equal(endTime) {
			slotEnd := current.Add(serviceDuration)
			
			// Check if this slot conflicts with any existing booking
			isAvailable := true
			for _, booking := range bookings {
				if timesOverlap(current, slotEnd, booking.BeginDateTime, booking.EndDateTime) {
					isAvailable = false
					break
				}
			}

			if isAvailable {
				availableSlots = append(availableSlots, current.Format("15:04"))
			}

			// Move to next 15-minute slot
			current = current.Add(15 * time.Minute)
		}
	}

	return availableSlots, nil
}

// timesOverlap checks if two time ranges overlap
func timesOverlap(start1, end1, start2, end2 time.Time) bool {
	return start1.Before(end2) && start2.Before(end1)
}

// ===== BARBER AVAILABILITY MANAGEMENT METHODS =====

// GetBarberFullAvailability returns all availability for a barber
func (r *Repository) GetBarberFullAvailability(ctx context.Context, barberID uuid.UUID) ([]db.HaircutBarberAvailability, *errLib.CommonError) {
	availability, err := r.Queries.GetBarberFullAvailability(ctx, barberID)
	if err != nil {
		log.Printf("Failed to get barber availability: %v", err)
		return nil, errLib.New("Failed to get barber availability", http.StatusInternalServerError)
	}
	return availability, nil
}

// GetBarberAvailabilityByID returns a specific availability record
func (r *Repository) GetBarberAvailabilityByID(ctx context.Context, id uuid.UUID) (db.HaircutBarberAvailability, *errLib.CommonError) {
	availability, err := r.Queries.GetBarberAvailabilityByID(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return db.HaircutBarberAvailability{}, errLib.New("Availability record not found", http.StatusNotFound)
		}
		log.Printf("Failed to get availability by ID: %v", err)
		return db.HaircutBarberAvailability{}, errLib.New("Failed to get availability", http.StatusInternalServerError)
	}
	return availability, nil
}

// CreateBarberAvailability creates a new availability record
func (r *Repository) CreateBarberAvailability(ctx context.Context, params db.InsertBarberAvailabilityParams) (db.HaircutBarberAvailability, *errLib.CommonError) {
	availability, err := r.Queries.InsertBarberAvailability(ctx, params)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) {
			if pqErr.Constraint == "unique_barber_day_time" {
				return db.HaircutBarberAvailability{}, errLib.New("Availability for this time slot already exists", http.StatusConflict)
			}
			if pqErr.Constraint == "fk_barber_availability" {
				return db.HaircutBarberAvailability{}, errLib.New("Barber not found", http.StatusNotFound)
			}
			if pqErr.Constraint == "check_time_order" {
				return db.HaircutBarberAvailability{}, errLib.New("End time must be after start time", http.StatusBadRequest)
			}
		}
		log.Printf("Failed to create availability: %v", err)
		return db.HaircutBarberAvailability{}, errLib.New("Failed to create availability", http.StatusInternalServerError)
	}
	return availability, nil
}

// UpsertBarberAvailability creates or updates availability record
func (r *Repository) UpsertBarberAvailability(ctx context.Context, params db.UpsertBarberAvailabilityParams) (db.HaircutBarberAvailability, *errLib.CommonError) {
	availability, err := r.Queries.UpsertBarberAvailability(ctx, params)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) {
			if pqErr.Constraint == "fk_barber_availability" {
				return db.HaircutBarberAvailability{}, errLib.New("Barber not found", http.StatusNotFound)
			}
			if pqErr.Constraint == "check_time_order" {
				return db.HaircutBarberAvailability{}, errLib.New("End time must be after start time", http.StatusBadRequest)
			}
		}
		log.Printf("Failed to upsert availability: %v", err)
		return db.HaircutBarberAvailability{}, errLib.New("Failed to save availability", http.StatusInternalServerError)
	}
	return availability, nil
}

// UpdateBarberAvailability updates an existing availability record
func (r *Repository) UpdateBarberAvailability(ctx context.Context, params db.UpdateBarberAvailabilityParams) (db.HaircutBarberAvailability, *errLib.CommonError) {
	availability, err := r.Queries.UpdateBarberAvailability(ctx, params)
	if err != nil {
		if err == sql.ErrNoRows {
			return db.HaircutBarberAvailability{}, errLib.New("Availability record not found", http.StatusNotFound)
		}
		var pqErr *pq.Error
		if errors.As(err, &pqErr) {
			if pqErr.Constraint == "check_time_order" {
				return db.HaircutBarberAvailability{}, errLib.New("End time must be after start time", http.StatusBadRequest)
			}
		}
		log.Printf("Failed to update availability: %v", err)
		return db.HaircutBarberAvailability{}, errLib.New("Failed to update availability", http.StatusInternalServerError)
	}
	return availability, nil
}

// DeleteBarberAvailability deletes an availability record
func (r *Repository) DeleteBarberAvailability(ctx context.Context, id uuid.UUID) *errLib.CommonError {
	rowsAffected, err := r.Queries.DeleteBarberAvailability(ctx, id)
	if err != nil {
		log.Printf("Failed to delete availability: %v", err)
		return errLib.New("Failed to delete availability", http.StatusInternalServerError)
	}
	if rowsAffected == 0 {
		return errLib.New("Availability record not found", http.StatusNotFound)
	}
	return nil
}

// DeleteBarberAvailabilityByDay deletes all availability for a specific day
func (r *Repository) DeleteBarberAvailabilityByDay(ctx context.Context, barberID uuid.UUID, dayOfWeek int32) *errLib.CommonError {
	params := db.DeleteBarberAvailabilityByDayParams{
		BarberID:  barberID,
		DayOfWeek: dayOfWeek,
	}
	_, err := r.Queries.DeleteBarberAvailabilityByDay(ctx, params)
	if err != nil {
		log.Printf("Failed to delete availability by day: %v", err)
		return errLib.New("Failed to delete availability", http.StatusInternalServerError)
	}
	return nil
}
