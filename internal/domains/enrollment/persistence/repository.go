package enrollment

import (
	database_errors "api/internal/constants"
	db "api/internal/domains/enrollment/persistence/sqlc/generated"
	"api/internal/domains/enrollment/values"
	errLib "api/internal/libs/errors"
	"context"
	"database/sql"
	"errors"
	"log"
	"net/http"

	"github.com/lib/pq"

	"github.com/google/uuid"
)

type Repository struct {
	Queries *db.Queries
}

func NewEnrollmentRepository(dbQueries *db.Queries) *Repository {
	return &Repository{
		Queries: dbQueries,
	}
}

func (r *Repository) GetEnrollments(c context.Context, eventId, customerId uuid.UUID) ([]values.EnrollmentReadDetails, *errLib.CommonError) {

	var args db.GetCustomerEnrollmentsParams

	args.CustomerID = uuid.NullUUID{
		UUID:  customerId,
		Valid: customerId != uuid.Nil,
	}

	args.EventID = uuid.NullUUID{
		UUID:  eventId,
		Valid: eventId != uuid.Nil,
	}

	dbEnrollments, err := r.Queries.GetCustomerEnrollments(c, args)

	if err != nil {
		log.Printf("Error getting customerEvents: %v", err)
		return nil, errLib.New("Internal server error", http.StatusInternalServerError)
	}

	enrollments := make([]values.EnrollmentReadDetails, len(dbEnrollments))

	for i, enrollment := range dbEnrollments {
		response := values.EnrollmentReadDetails{
			ID:          enrollment.ID,
			CustomerID:  enrollment.CustomerID,
			EventID:     enrollment.EventID,
			CreatedAt:   enrollment.CreatedAt,
			UpdatedAt:   enrollment.UpdatedAt,
			IsCancelled: enrollment.IsCancelled,
		}

		if enrollment.CheckedInAt.Valid {
			response.CheckedInAt = &enrollment.CheckedInAt.Time
		}

		enrollments[i] = response
	}

	return enrollments, nil
}

func (r *Repository) UnEnrollCustomer(c context.Context, id uuid.UUID) *errLib.CommonError {
	row, err := r.Queries.UnEnrollCustomer(c, id)

	if err != nil {
		return errLib.New("Internal server error", http.StatusInternalServerError)
	}

	if row == 0 {
		return errLib.New("Enrollment not found", http.StatusNotFound)
	}

	return nil
}

func (r *Repository) EnrollCustomer(c context.Context, input values.EnrollmentCreateDetails) (values.EnrollmentReadDetails, *errLib.CommonError) {

	var returnedValues values.EnrollmentReadDetails

	params := db.EnrollCustomerParams{
		CustomerID: input.CustomerId,
		EventID:    input.EventId,
		CheckedInAt: sql.NullTime{
			Valid: false,
		},
		IsCancelled: false,
	}

	enrollment, err := r.Queries.EnrollCustomer(c, params)

	if err != nil {
		// Check if the error is a unique violation (error code 23505)
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == database_errors.UniqueViolation {
			// Return a custom error for unique violation
			return returnedValues, errLib.New("Duplicate info", http.StatusConflict)
		}

		// Return a generic internal server error for other cases
		log.Println("error creating enrollment: ", err)
		return returnedValues, errLib.New("Internal server error", http.StatusInternalServerError)
	}

	response := values.EnrollmentReadDetails{
		ID:         enrollment.ID,
		CustomerID: enrollment.CustomerID,
		EventID:    enrollment.EventID,
		CreatedAt:  enrollment.CreatedAt,
		UpdatedAt:  enrollment.UpdatedAt,
	}

	if enrollment.CheckedInAt.Valid {
		response.CheckedInAt = &enrollment.CheckedInAt.Time
	}

	return response, nil
}

func (r *Repository) GetEventIsFull(c context.Context, eventId uuid.UUID) (bool, *errLib.CommonError) {

	isFull, err := r.Queries.GetEventIsFull(c, eventId)

	if err != nil {
		log.Printf("Error getting info: %v", err)
		return true, errLib.New("Internal server error", http.StatusInternalServerError)
	}

	return isFull, nil
}
