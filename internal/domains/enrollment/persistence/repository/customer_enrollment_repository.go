package enrollment

import (
	db "api/internal/domains/enrollment/persistence/sqlc/generated"
	errLib "api/internal/libs/errors"
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"time"

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

func (r *CustomerEnrollmentRepository) EnrollCustomerInProgramEvents(ctx context.Context, customerID, programID uuid.UUID) *errLib.CommonError {

	if err := r.Queries.EnrollCustomerInProgramEvents(ctx, db.EnrollCustomerInProgramEventsParams{
		CustomerID: customerID,
		ProgramID:  programID,
	}); err != nil {
		return errLib.New(fmt.Sprintf("error enrolling customer in program events: %v", err), http.StatusBadRequest)
	}
	return nil
}

func (r *CustomerEnrollmentRepository) EnrollCustomerInMembershipPlan(ctx context.Context, customerID, planID uuid.UUID, cancelAtDateTime time.Time) *errLib.CommonError {

	if err := r.Queries.EnrollCustomerInMembershipPlan(ctx, db.EnrollCustomerInMembershipPlanParams{
		CustomerID:       customerID,
		MembershipPlanID: planID,
		Status:           db.MembershipMembershipStatusActive,
		StartDate:        time.Now(),
		RenewalDate: sql.NullTime{
			Time:  cancelAtDateTime,
			Valid: !cancelAtDateTime.IsZero(),
		},
	}); err != nil {
		return errLib.New(fmt.Sprintf("error enrolling customer in membership plan: %v", err), http.StatusBadRequest)
	}
	return nil
}

//
//func (r *CustomerEnrollmentRepository) EnrollCustomer(c context.Context, eventID, customerID uuid.UUID) *errLib.CommonError {
//
//	params := db.EnrollCustomerParams{
//		CustomerID: customerID,
//		EventID:    eventID,
//		CheckedInAt: sql.NullTime{
//			Valid: false,
//		},
//	}
//
//	_, err := r.Queries.EnrollCustomer(c, params)
//
//	if err != nil {
//		var pqErr *pq.Error
//		if errors.As(err, &pqErr) && pqErr.Code == databaseErrors.UniqueViolation {
//			// Return a custom error for unique violation
//			return errLib.New("Customer is already enrolled", http.StatusConflict)
//		}
//
//		// Return a generic internal server error for other cases
//		log.Println("error creating enrollment: ", err)
//		return errLib.New("Internal server error", http.StatusInternalServerError)
//	}
//
//	return nil
