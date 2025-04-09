package enrollment

import (
	databaseErrors "api/internal/constants"
	db "api/internal/domains/enrollment/persistence/sqlc/generated"
	errLib "api/internal/libs/errors"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/lib/pq"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
)

type CustomerEnrollmentRepository struct {
	Queries *db.Queries
	Db      *sql.DB
}

func NewEnrollmentRepository(db *sql.DB, dbQueries *db.Queries) *CustomerEnrollmentRepository {
	return &CustomerEnrollmentRepository{
		Db:      db,
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

	tx, err := r.Db.BeginTx(ctx, &sql.TxOptions{
		Isolation: sql.LevelSerializable,
	})

	if err != nil {
		return errLib.New("Failed to begin transaction", http.StatusInternalServerError)
	}

	defer func() {
		if err = tx.Rollback(); err != nil && !errors.Is(sql.ErrTxDone, err) {
			fmt.Println("Failed to rollback transaction:", err)
		}
	}()

	qtx := db.New(tx)

	isFull, err := qtx.GetProgramIsFull(ctx, programID)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// rollback the transaction if the program is not found
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				log.Println("Failed to rollback transaction:", rollbackErr)
				return errLib.New("Failed to rollback transaction", http.StatusInternalServerError)
			}
			return errLib.New("Program or capacity not found for the program", http.StatusNotFound)
		}
		return errLib.New(fmt.Sprintf("error checking program availability: %v", err), http.StatusInternalServerError)
	}

	if isFull {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			log.Println("Failed to rollback transaction:", rollbackErr)
			return errLib.New("Failed to rollback transaction", http.StatusInternalServerError)
		}
		return errLib.New("Program is full", http.StatusConflict)
	}

	if err = qtx.EnrollCustomerInProgramEvents(ctx, db.EnrollCustomerInProgramEventsParams{
		CustomerID: customerID,
		ProgramID:  programID,
	}); err != nil {

		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			log.Println("Failed to rollback transaction:", rollbackErr)
			return errLib.New("Failed to rollback transaction", http.StatusInternalServerError)
		}
		return errLib.New(fmt.Sprintf("Internal server error when enrolling customer in program events: %v", err), http.StatusInternalServerError)
	}

	if err = tx.Commit(); err != nil {
		log.Println("Failed to commit transaction:", err)

		if isSerializationError(err) {

			if err = tx.Rollback(); err != nil {
				log.Println("Failed to rollback transaction:", err)
				return errLib.New("Failed to rollback transaction", http.StatusInternalServerError)
			}
			return errLib.New("Transaction serialization error", http.StatusConflict)
		}

		return errLib.New("Failed to commit enrollment", http.StatusInternalServerError)
	}

	return nil
}

func isSerializationError(err error) bool {

	log.Println("Transaction serialization error 1:", err)

	if err == nil {
		return false
	}

	log.Println("Transaction serialization error 2:", err)

	// SQL driver might wrap the error, so check the underlying error
	var pqErr *pq.Error
	if errors.As(err, &pqErr) {
		return pqErr.Code == databaseErrors.TxSerializationError
	}
	return false
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
