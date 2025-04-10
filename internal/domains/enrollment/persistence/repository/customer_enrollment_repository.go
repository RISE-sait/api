package enrollment

import (
	databaseErrors "api/internal/constants"
	dbEnrollment "api/internal/domains/enrollment/persistence/sqlc/generated"
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
	Queries *dbEnrollment.Queries
	Db      *sql.DB
}

func NewEnrollmentRepository(db *sql.DB) *CustomerEnrollmentRepository {
	return &CustomerEnrollmentRepository{
		Db:      db,
		Queries: dbEnrollment.New(db),
	}
}

func (r *CustomerEnrollmentRepository) WithTx(tx *sql.Tx) *CustomerEnrollmentRepository {
	return &CustomerEnrollmentRepository{
		Queries: r.Queries.WithTx(tx),
		Db:      r.Db,
	}
}

func (r *CustomerEnrollmentRepository) UnEnrollCustomerFromEvent(c context.Context, eventID, customerID uuid.UUID) *errLib.CommonError {
	row, err := r.Queries.UnEnrollCustomerFromEvent(c, dbEnrollment.UnEnrollCustomerFromEventParams{
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

func (r *CustomerEnrollmentRepository) EnrollCustomerInMembershipPlan(ctx context.Context, customerID, planID uuid.UUID, cancelAtDateTime time.Time) *errLib.CommonError {

	if err := r.Queries.EnrollCustomerInMembershipPlan(ctx, dbEnrollment.EnrollCustomerInMembershipPlanParams{
		CustomerID:       customerID,
		MembershipPlanID: planID,
		Status:           dbEnrollment.MembershipMembershipStatusActive,
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

func (r *CustomerEnrollmentRepository) checkIfProgramExists(ctx context.Context, programID uuid.UUID) (bool, *errLib.CommonError) {

	isExist, err := r.Queries.CheckProgramExists(ctx, programID)

	if err != nil {
		log.Println("error checking program existence: ", err)
		return false, errLib.New("Internal server error while finding program", http.StatusInternalServerError)
	}

	return isExist, nil
}

func (r *CustomerEnrollmentRepository) checkIfProgramCapacityExist(ctx context.Context, programID uuid.UUID) (bool, *errLib.CommonError) {

	isExist, err := r.Queries.CheckProgramCapacityExists(ctx, programID)

	if err != nil {
		log.Println("error checking program capacity: ", err)
		return false, errLib.New("Internal server error while finding program capacity", http.StatusInternalServerError)
	}

	return isExist, nil
}

func (r *CustomerEnrollmentRepository) GetProgramIsFull(ctx context.Context, programID uuid.UUID) (bool, *errLib.CommonError) {

	if isExist, err := r.checkIfProgramExists(ctx, programID); err != nil {
		return false, err
	} else if !isExist {
		return false, errLib.New("Program not found", http.StatusNotFound)
	}

	if isExist, err := r.checkIfProgramCapacityExist(ctx, programID); err != nil {
		return false, err
	} else if !isExist {
		return false, errLib.New("Capacity for program not found", http.StatusNotFound)
	}

	isFull, err := r.Queries.CheckProgramIsFull(ctx, programID)

	if err != nil {
		return true, errLib.New(fmt.Sprintf("error checking program availability: %v", err), http.StatusInternalServerError)
	}

	return isFull, nil
}

func (r *CustomerEnrollmentRepository) GetEventIsFull(ctx context.Context, eventID uuid.UUID) (bool, *errLib.CommonError) {

	isFull, err := r.Queries.CheckEventIsFull(ctx, eventID)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return true, errLib.New("Event or capacity for event not found", http.StatusNotFound)
		}

		return true, errLib.New(fmt.Sprintf("error checking event availability: %v", err), http.StatusInternalServerError)
	}

	return isFull, nil
}

func (r *CustomerEnrollmentRepository) ReserveSeatInProgram(ctx context.Context, programID, customerID uuid.UUID) *errLib.CommonError {

	if isExist, err := r.checkIfProgramExists(ctx, programID); err != nil {
		return err
	} else if !isExist {
		return errLib.New("Program not found", http.StatusNotFound)
	}

	if isExist, err := r.checkIfProgramCapacityExist(ctx, programID); err != nil {
		return err
	} else if !isExist {
		return errLib.New("Capacity for program not found", http.StatusNotFound)
	}

	affectedRows, err := r.Queries.ReserveSeatInProgram(ctx, dbEnrollment.ReserveSeatInProgramParams{
		ProgramID:  programID,
		CustomerID: customerID,
	})

	if err != nil {

		var pqErr *pq.Error
		if errors.As(err, &pqErr) {

			switch pqErr.Code {
			case databaseErrors.UniqueViolation:
				return errLib.New("Customer is already enrolled in the program", http.StatusConflict)
			case databaseErrors.TxSerializationError:
				return errLib.New("Too many people enrolled at the same time. Please try again.", http.StatusConflict)
			}
		}

		log.Println("error reserving seat in program: ", err)
		return errLib.New("error reserving seat in program", http.StatusInternalServerError)
	}

	if affectedRows == 0 {
		return errLib.New("Failed to book program for unknown reason. Please try again or contact support.", http.StatusInternalServerError)
	}

	return nil
}

func (r *CustomerEnrollmentRepository) UpdateReservationStatusInProgram(ctx context.Context, programID, customerID uuid.UUID, status dbEnrollment.PaymentStatus) *errLib.CommonError {

	if isExist, err := r.checkIfProgramExists(ctx, programID); err != nil {
		return err
	} else if !isExist {
		return errLib.New("Program not found", http.StatusNotFound)
	}

	if isExist, err := r.checkIfProgramCapacityExist(ctx, programID); err != nil {
		return err
	} else if !isExist {
		return errLib.New("Program not found", http.StatusNotFound)
	}

	affectedRows, err := r.Queries.UpdateSeatReservationStatusInProgram(ctx, dbEnrollment.UpdateSeatReservationStatusInProgramParams{
		ProgramID:     programID,
		CustomerID:    customerID,
		PaymentStatus: status,
	})

	if err != nil {

		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == databaseErrors.UniqueViolation {
			return errLib.New("Customer is already enrolled in the program", http.StatusConflict)
		}

		return errLib.New(fmt.Sprintf("error updating reservation status for program: %v", err), http.StatusInternalServerError)
	}

	if affectedRows == 0 {
		return errLib.New("Event is full, or someone is booking at the same time, please try again.", http.StatusConflict)
	}

	return nil
}

func (r *CustomerEnrollmentRepository) ReserveSeatInEvent(ctx context.Context, eventID, customerID uuid.UUID) *errLib.CommonError {

	affectedRows, err := r.Queries.ReserveSeatInEvent(ctx, dbEnrollment.ReserveSeatInEventParams{
		EventID:    eventID,
		CustomerID: customerID,
	})

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errLib.New("Event not found", http.StatusNotFound)
		}

		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == databaseErrors.UniqueViolation {
			return errLib.New("Customer is already enrolled in the event", http.StatusConflict)
		}

		return errLib.New(fmt.Sprintf("error checking event availability: %v", err), http.StatusInternalServerError)
	}

	if affectedRows == 0 {
		return errLib.New("Event is full, or someone is booking at the same time, please try again.", http.StatusConflict)
	}

	return nil
}

func (r *CustomerEnrollmentRepository) EnrollCustomerInProgram(ctx context.Context, programID, customerID uuid.UUID) *errLib.CommonError {

	if isExist, err := r.checkIfProgramExists(ctx, programID); err != nil {
		return err
	} else if !isExist {
		return errLib.New("Program not found", http.StatusNotFound)
	}

	params := dbEnrollment.EnrollCustomerInProgramParams{
		CustomerID: customerID,
		ProgramID:  programID,
	}

	if err := r.Queries.EnrollCustomerInProgram(ctx, params); err != nil {
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

func (r *CustomerEnrollmentRepository) EnrollCustomerInEvent(ctx context.Context, eventID, customerID uuid.UUID) *errLib.CommonError {

	params := dbEnrollment.EnrollCustomerInEventParams{
		CustomerID: customerID,
		EventID:    eventID,
	}

	if err := r.Queries.EnrollCustomerInEvent(ctx, params); err != nil {
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
