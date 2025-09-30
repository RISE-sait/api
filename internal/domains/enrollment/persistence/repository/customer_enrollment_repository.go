package enrollment

import (
	databaseErrors "api/internal/constants"
	"api/internal/di"
	dbEnrollment "api/internal/domains/enrollment/persistence/sqlc/generated"
	"api/internal/domains/event/service"
	"api/internal/domains/program"
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
	Queries        *dbEnrollment.Queries
	ProgramService *program.Service
	EventService   *service.Service
	Db             *sql.DB // Add the Db field here
}

func NewEnrollmentRepository(container *di.Container) *CustomerEnrollmentRepository {
	return &CustomerEnrollmentRepository{
		Db:             container.DB,
		Queries:        dbEnrollment.New(container.DB),
		ProgramService: program.NewProgramService(container),
		EventService:   service.NewEventService(container),
	}
}

func (r *CustomerEnrollmentRepository) WithTx(tx *sql.Tx) *CustomerEnrollmentRepository {
	return &CustomerEnrollmentRepository{
		Queries:        r.Queries.WithTx(tx),
		Db:             r.Db,
		ProgramService: r.ProgramService,
		EventService:   r.EventService,
	}
}

func (r *CustomerEnrollmentRepository) UnEnrollCustomerFromEvent(c context.Context, eventID, customerID uuid.UUID) *errLib.CommonError {
	row, err := r.Queries.UnEnrollCustomerFromEvent(c, dbEnrollment.UnEnrollCustomerFromEventParams{
		CustomerID: customerID,
		EventID:    eventID,
	})

	if err != nil {
		log.Println("error unenrolling customer from event: ", err)
		return errLib.New("Internal server error", http.StatusInternalServerError)
	}

	if row == 0 {
		return errLib.New("Enrollment not found", http.StatusNotFound)
	}

	return nil
}

func (r *CustomerEnrollmentRepository) EnrollCustomerInMembershipPlan(ctx context.Context, customerID, planID uuid.UUID, cancelAtDateTime time.Time, startTime time.Time) *errLib.CommonError {

	if err := r.Queries.EnrollCustomerInMembershipPlan(ctx, dbEnrollment.EnrollCustomerInMembershipPlanParams{
		CustomerID:       customerID,
		MembershipPlanID: planID,
		Status:           dbEnrollment.MembershipMembershipStatusActive,
		StartDate:        startTime,
		RenewalDate: sql.NullTime{
			Time:  cancelAtDateTime,
			Valid: !cancelAtDateTime.IsZero(),
		},
		SubscriptionSource: sql.NullString{
			String: "subscription",
			Valid:  true,
		},
	}); err != nil {
		// Handle PostgreSQL constraint violations
		var pqErr *pq.Error
		if errors.As(err, &pqErr) {
			switch pqErr.Code {
			case databaseErrors.UniqueViolation:
				// Check if it's the unique_customer_membership_plan constraint
				if pqErr.Constraint == "unique_customer_membership_plan" {
					return errLib.New("Customer is already enrolled in this membership plan", http.StatusConflict)
				}
				return errLib.New("Duplicate enrollment detected", http.StatusConflict)
			case databaseErrors.TxSerializationError:
				return errLib.New("Too many enrollment requests at the same time. Please try again.", http.StatusConflict)
			case databaseErrors.ForeignKeyViolation:
				return errLib.New("Invalid customer or membership plan ID", http.StatusBadRequest)
			}
		}
		
		log.Printf("error enrolling customer %s in membership plan %s: %v", customerID, planID, err)
		return errLib.New("Failed to enroll in membership plan", http.StatusInternalServerError)
	}
	return nil
}

func (r *CustomerEnrollmentRepository) CheckIfEventCapacityExist(ctx context.Context, eventID uuid.UUID) (bool, *errLib.CommonError) {

	isExist, err := r.Queries.CheckEventCapacityExists(ctx, eventID)

	if err != nil {
		log.Println("error checking event capacity: ", err)
		return false, errLib.New("Internal server error while finding event capacity", http.StatusInternalServerError)
	}

	return isExist, nil
}

func (r *CustomerEnrollmentRepository) GetProgramIsFull(ctx context.Context, programID uuid.UUID) (bool, *errLib.CommonError) {

	isFull, err := r.Queries.CheckProgramIsFull(ctx, programID)

	if err != nil {
		log.Println("error checking program availability: ", err)
		return true, errLib.New("error checking program availability", http.StatusInternalServerError)
	}

	return isFull, nil
}

func (r *CustomerEnrollmentRepository) GetEventIsFull(ctx context.Context, eventID uuid.UUID) (bool, *errLib.CommonError) {

	isFull, err := r.Queries.CheckEventIsFull(ctx, eventID)

	if err != nil {

		log.Println("error checking event availability: ", err)
		return true, errLib.New("error checking event availability: %v", http.StatusInternalServerError)
	}

	return isFull, nil
}

func (r *CustomerEnrollmentRepository) ReserveSeatInProgram(ctx context.Context, programID, customerID uuid.UUID) *errLib.CommonError {

	if isEnrolled, err := r.Queries.GetCustomerIsEnrolledInProgram(ctx, dbEnrollment.GetCustomerIsEnrolledInProgramParams{
		CustomerID: customerID,
		ProgramID:  programID,
	}); err != nil {
		log.Println("error checking if customer is enrolled in program: ", err)
		return errLib.New("error checking if customer is enrolled in program", http.StatusInternalServerError)
	} else if isEnrolled {
		return errLib.New("Customer is already enrolled in the program", http.StatusConflict)
	}

	affectedRows, err := r.Queries.ReserveSeatInProgram(ctx, dbEnrollment.ReserveSeatInProgramParams{
		ProgramID:  programID,
		CustomerID: customerID,
	})

	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == databaseErrors.TxSerializationError {
			return errLib.New("Too many people enrolled at the same time. Please try again.", http.StatusConflict)
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

	affectedRows, err := r.Queries.UpdateSeatReservationStatusInProgram(ctx, dbEnrollment.UpdateSeatReservationStatusInProgramParams{
		ProgramID:     programID,
		CustomerID:    customerID,
		PaymentStatus: status,
	})

	if err != nil {

		log.Println("error updating reservation status for program: ", err)

		return errLib.New(fmt.Sprintf("error updating reservation status for program: %v", err), http.StatusInternalServerError)
	}

	if affectedRows == 0 {
		return errLib.New("Error confirming customer's reservation status for unknown reason, please try again.", http.StatusInternalServerError)
	}

	return nil
}

func (r *CustomerEnrollmentRepository) UpdateReservationStatusInEvent(ctx context.Context, eventID, customerID uuid.UUID, status dbEnrollment.PaymentStatus) *errLib.CommonError {

	affectedRows, err := r.Queries.UpdateSeatReservationStatusInEvent(ctx, dbEnrollment.UpdateSeatReservationStatusInEventParams{
		EventID:       eventID,
		CustomerID:    customerID,
		PaymentStatus: status,
	})

	if err != nil {

		log.Println("error updating reservation status for event: ", err)

		return errLib.New(fmt.Sprintf("error updating reservation status for event: %v", err), http.StatusInternalServerError)
	}

	if affectedRows == 0 {
		return errLib.New("Error confirming customer's reservation status for unknown reason, please try again.", http.StatusInternalServerError)
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
		if errors.As(err, &pqErr) {
			switch pqErr.Code {
			case databaseErrors.UniqueViolation:
				return errLib.New("Customer is already enrolled in the event", http.StatusConflict)
			case databaseErrors.TxSerializationError:
				return errLib.New("Too many people enrolled at the same time. Please try again.", http.StatusConflict)
			}
		}

		return errLib.New(fmt.Sprintf("error checking event availability: %v", err), http.StatusInternalServerError)
	}

	if affectedRows == 0 {
		return errLib.New("Failed to book event for unknown reason. Please try again or contact support.", http.StatusInternalServerError)
	}

	return nil
}

func (r *CustomerEnrollmentRepository) EnrollCustomerInProgram(ctx context.Context, programID, customerID uuid.UUID) *errLib.CommonError {

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
