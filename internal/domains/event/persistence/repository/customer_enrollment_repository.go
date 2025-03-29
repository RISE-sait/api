package event

import (
	databaseErrors "api/internal/constants"
	db "api/internal/domains/event/persistence/sqlc/generated"
	errLib "api/internal/libs/errors"
	"context"
	"database/sql"
	"errors"
	"log"
	"net/http"

	"github.com/lib/pq"

	"github.com/google/uuid"
)

type CustomerEnrollmentRepository struct {
	Queries *db.Queries
}

func NewEnrollmentRepository(dbQueries *db.Queries) *CustomerEnrollmentRepository {
	return &CustomerEnrollmentRepository{
		Queries: dbQueries,
	}
}

func (r *CustomerEnrollmentRepository) UnEnrollCustomer(c context.Context, eventID, customerID uuid.UUID) *errLib.CommonError {
	row, err := r.Queries.UnEnrollCustomer(c, db.UnEnrollCustomerParams{
		CustomerID: customerID,
		EventID:    eventID,
	})

	if err != nil {
		return errLib.New("Internal server error", http.StatusInternalServerError)
	}

	if row == 0 {
		return errLib.New("Enrollment not found", http.StatusNotFound)
	}

	return nil
}

func (r *CustomerEnrollmentRepository) EnrollCustomer(c context.Context, eventID, customerID uuid.UUID) *errLib.CommonError {

	params := db.EnrollCustomerParams{
		CustomerID: customerID,
		EventID:    eventID,
		CheckedInAt: sql.NullTime{
			Valid: false,
		},
	}

	_, err := r.Queries.EnrollCustomer(c, params)

	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == databaseErrors.UniqueViolation {
			// Return a custom error for unique violation
			return errLib.New("Customer is already enrolled", http.StatusConflict)
		}

		// Return a generic internal server error for other cases
		log.Println("error creating enrollment: ", err)
		return errLib.New("Internal server error", http.StatusInternalServerError)
	}

	return nil
}

func (r *CustomerEnrollmentRepository) GetEventIsFull(c context.Context, eventId uuid.UUID) (bool, *errLib.CommonError) {

	isFull, err := r.Queries.GetEventIsFull(c, eventId)

	if err != nil {
		log.Printf("Error getting info: %v", err)
		return true, errLib.New("Internal server error", http.StatusInternalServerError)
	}

	return isFull, nil
}
