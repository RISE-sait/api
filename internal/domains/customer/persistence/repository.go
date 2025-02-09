package persistence

import (
	"api/internal/di"
	entity "api/internal/domains/customer/entities"
	db "api/internal/domains/customer/persistence/sqlc/generated"
	errLib "api/internal/libs/errors"
	"time"

	"context"
	"log"
	"net/http"

	"github.com/google/uuid"
)

type CustomersRepository struct {
	Queries *db.Queries
}

func NewCustomersRepository(container *di.Container) *CustomersRepository {
	return &CustomersRepository{
		Queries: container.Queries.CustomerDb,
	}
}

// func (r *CustomersRepository) CreateEvent(c context.Context, event *values.EventDetails) *errLib.CommonError {

// 	dbParams := db.CreateEventParams{
// 		BeginTime: event.BeginTime,
// 		EndTime:   event.EndTime,
// 		CourseID: uuid.NullUUID{
// 			UUID:  event.CourseID,
// 			Valid: event.CourseID != uuid.Nil,
// 		},
// 		FacilityID: event.FacilityID,
// 		Day:        db.DayEnum(event.Day),
// 	}

// 	row, err := r.Queries.CreateEvent(c, dbParams)

// 	if err != nil {
// 		log.Printf("Failed to create event: %+v. Error: %v", event, err.Error())
// 		return errLib.New("Internal server error", http.StatusInternalServerError)
// 	}

// 	if row == 0 {
// 		return errLib.New("Course or facility not found", http.StatusNotFound)
// 	}

// 	return nil
// }

func (r *CustomersRepository) GetCustomersByEventId(ctx context.Context, id uuid.UUID) ([]entity.Customer, *errLib.CommonError) {

	dbCustomers, err := r.Queries.GetCustomersForEvent(ctx, id)

	if err != nil {
		log.Println("Failed to get events: ", err.Error())
		return nil, errLib.New("Internal server error", http.StatusInternalServerError)
	}

	customers := make([]entity.Customer, len(dbCustomers))
	for i, dbCustomer := range dbCustomers {

		var name *string

		if dbCustomer.Name.Valid {
			name = &dbCustomer.Name.String
		}

		membershipRenewalDate := dbCustomer.MembershipRenewalDate.Time
		var membershipRenewalDatePtr *time.Time
		if !membershipRenewalDate.IsZero() {
			membershipRenewalDatePtr = &membershipRenewalDate
		}

		customers[i] = entity.Customer{
			CustomerID:            dbCustomer.CustomerID,
			Name:                  name,
			Email:                 dbCustomer.Email,
			MembershipName:        dbCustomer.MembershipName,
			Attendance:            dbCustomer.Attendance,
			MembershipRenewalDate: membershipRenewalDatePtr,
		}

	}

	return customers, nil
}

// func (r *CustomersRepository) UpdateEvent(c context.Context, event *values.EventAllFields) *errLib.CommonError {
// 	dbEventParams := db.UpdateEventParams{
// 		BeginTime: event.BeginTime,
// 		EndTime:   event.EndTime,
// 		CourseID: uuid.NullUUID{
// 			UUID:  event.CourseID,
// 			Valid: event.CourseID != uuid.Nil,
// 		},
// 		FacilityID: event.FacilityID,
// 		Day:        db.DayEnum(event.Day),
// 		ID:         event.ID,
// 	}

// 	row, err := r.Queries.UpdateEvent(c, dbEventParams)

// 	if err != nil {
// 		log.Printf("Failed to update event: %+v. Error: %v", event, err.Error())
// 		return errLib.New("Internal server error", http.StatusInternalServerError)
// 	}

// 	if row == 0 {
// 		return errLib.New("Course or facility not found", http.StatusNotFound)
// 	}
// 	return nil
// }

// func (r *CustomersRepository) DeleteEvent(c context.Context, id uuid.UUID) *errLib.CommonError {
// 	row, err := r.Queries.DeleteEvent(c, id)

// 	if err != nil {
// 		log.Printf("Failed to delete event with ID: %s. Error: %s", id, err.Error())
// 		return errLib.New("Internal server error", http.StatusInternalServerError)
// 	}

// 	if row == 0 {
// 		return errLib.New("Event not found", http.StatusNotFound)
// 	}

// 	return nil
// }

// func (r *CustomersRepository) GetEventDetails(ctx context.Context, id uuid.UUID) (*entity.Event, *errLib.CommonError) {

// 	eventDetails, err := r.Queries.GetEventById(ctx, id)

// 	if err != nil {
// 		log.Println("Failed to get event details: ", err.Error())
// 		return nil, errLib.New("Internal server error", http.StatusInternalServerError)
// 	}

// 	event := &entity.Event{
// 		ID:        eventDetails.ID,
// 		Course:    eventDetails.Course,
// 		Facility:  eventDetails.Facility,
// 		BeginTime: eventDetails.BeginTime,
// 		EndTime:   eventDetails.EndTime,
// 		Day:       string(eventDetails.Day),
// 	}

// 	return event, nil
// }
