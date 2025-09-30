package enrollment

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"net/http"
	"strings"
	"time"

	databaseErrors "api/internal/constants"
	"api/internal/di"
	repo "api/internal/domains/enrollment/persistence/repository"
	dbEnrollment "api/internal/domains/enrollment/persistence/sqlc/generated"
	event "api/internal/domains/event/service"
	"api/internal/domains/program"
	errLib "api/internal/libs/errors"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

type CustomerEnrollmentService struct {
	repo           *repo.CustomerEnrollmentRepository
	programService *program.Service
	eventService   *event.Service
	db             *sql.DB
}

func NewCustomerEnrollmentService(container *di.Container) *CustomerEnrollmentService {
	return &CustomerEnrollmentService{
		repo:           repo.NewEnrollmentRepository(container),
		programService: program.NewProgramService(container),
		eventService:   event.NewEventService(container),
		db:             container.DB,
	}
}

// transaction isolation level is set to Serializable to prevent write skew due to race conditions
// this is important for the enrollment process, as multiple users may try to enroll at the same time
// and cause overbooking
func (s *CustomerEnrollmentService) executeInTx(ctx context.Context, fn func(repo *repo.CustomerEnrollmentRepository) *errLib.CommonError) *errLib.CommonError {
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{
		Isolation: sql.LevelSerializable,
	})
	if err != nil {
		log.Printf("Failed to begin transaction: %v", err)
		return errLib.New("Failed to begin transaction", http.StatusInternalServerError)
	}

	defer func() {
		if err = tx.Rollback(); err != nil && !errors.Is(err, sql.ErrTxDone) {
			log.Printf("Rollback error (usually harmless): %v", err)
		}
	}()

	if txErr := fn(s.repo.WithTx(tx)); txErr != nil {
		return txErr
	}

	if err = tx.Commit(); err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == databaseErrors.TxSerializationError {
			return errLib.New(
				"Too many people enrolled at the same time. Please try again.",
				http.StatusConflict,
			)
		}
		return errLib.New("Failed to commit transaction", http.StatusInternalServerError)
	}
	return nil
}

// should only be used if want to force enrollment, default behavior is to reserve a seat, then update reservation status

//func (s *CustomerEnrollmentService) EnrollCustomerInProgram(ctx context.Context, customerID, programID uuid.UUID) *errLib.CommonError {
//
//	if _, err := s.programService.GetProgram(ctx, programID); err != nil {
//		return err
//	}
//
//	return s.repo.EnrollCustomerInProgram(ctx, programID, customerID)
//}
//
//func (s *CustomerEnrollmentService) EnrollCustomerInEvent(ctx context.Context, customerID, eventID uuid.UUID) *errLib.CommonError {
//
//	if _, err := s.eventService.GetEvent(ctx, eventID); err != nil {
//		return err
//	}
//
//	return s.repo.EnrollCustomerInEvent(ctx, customerID, eventID)
//}

func (s *CustomerEnrollmentService) EnrollCustomerInMembershipPlan(ctx context.Context, customerID, planID uuid.UUID, cancelAtDateTime time.Time, startTime time.Time) *errLib.CommonError {
	return s.repo.EnrollCustomerInMembershipPlan(ctx, customerID, planID, cancelAtDateTime, startTime)
}

func (s *CustomerEnrollmentService) UnEnrollCustomerFromEvent(ctx context.Context, eventID, customerID uuid.UUID) *errLib.CommonError {
	return s.repo.UnEnrollCustomerFromEvent(ctx, eventID, customerID)
}

func (s *CustomerEnrollmentService) ReserveSeatInEvent(ctx context.Context, eventID, customerID uuid.UUID) *errLib.CommonError {
	return s.executeInTx(ctx, func(r *repo.CustomerEnrollmentRepository) *errLib.CommonError {
		if _, err := s.eventService.GetEvent(ctx, eventID); err != nil {
			return err
		}

		capacityExist, err := r.CheckIfEventCapacityExist(ctx, eventID)
		if err != nil {
			return err
		}
		if !capacityExist {
			return errLib.New("Capacity for event not found", http.StatusNotFound)
		}

		if isFull, err := r.GetEventIsFull(ctx, eventID); err != nil {
			return err
		} else if isFull {
			return errLib.New("Event is full", http.StatusConflict)
		}
		return r.ReserveSeatInEvent(ctx, eventID, customerID)
	})
}

func (s *CustomerEnrollmentService) ReserveSeatInProgram(ctx context.Context, programID, customerID uuid.UUID) *errLib.CommonError {
	return s.executeInTx(ctx, func(r *repo.CustomerEnrollmentRepository) *errLib.CommonError {
		if getProgram, err := s.programService.GetProgram(ctx, programID); err != nil {
			return err
		} else if getProgram.Capacity == nil {
			return errLib.New("Program does not have a capacity", http.StatusBadRequest)
		}

		isFull, err := r.GetProgramIsFull(ctx, programID)
		if err != nil {
			return err
		}

		if isFull {
			return errLib.New("Program is full", http.StatusConflict)
		}
		return r.ReserveSeatInProgram(ctx, programID, customerID)
	})
}

func (s *CustomerEnrollmentService) UpdateReservationStatusInProgram(ctx context.Context, programID, customerID uuid.UUID, status dbEnrollment.PaymentStatus) *errLib.CommonError {
	if _, err := s.programService.GetProgram(ctx, programID); err != nil {
		return err
	}

	return s.repo.UpdateReservationStatusInProgram(ctx, programID, customerID, status)
}

func (s *CustomerEnrollmentService) UpdateReservationStatusInEvent(ctx context.Context, eventID, customerID uuid.UUID, status dbEnrollment.PaymentStatus) *errLib.CommonError {
	// Try to update the reservation first - if it fails due to foreign key constraint,
	// then we know the event doesn't exist. This avoids potential isolation level issues.
	err := s.repo.UpdateReservationStatusInEvent(ctx, eventID, customerID, status)
	if err != nil {
		// Check if this is a foreign key constraint error (event doesn't exist)
		if err.HTTPCode == 400 && (strings.Contains(err.Message, "event") || strings.Contains(err.Message, "foreign key")) {
			log.Printf("WARNING: Event %s not found when updating reservation for customer %s - event may have been deleted", eventID, customerID)
			return nil // Don't fail webhook for missing events
		}
		// Check if no reservation exists to update (0 affected rows)
		if err.HTTPCode == 500 && strings.Contains(err.Message, "Error confirming customer's reservation status") {
			log.Printf("WARNING: No reservation found for customer %s in event %s - creating reservation", customerID, eventID)
			
			// Try to create the reservation first, then update status
			if createErr := s.ReserveSeatInEvent(ctx, eventID, customerID); createErr != nil {
				log.Printf("ERROR: Failed to create reservation for customer %s in event %s: %v", customerID, eventID, createErr)
				return nil // Don't fail webhook even if creation fails
			}
			
			// Now try to update the status again
			if updateErr := s.repo.UpdateReservationStatusInEvent(ctx, eventID, customerID, status); updateErr != nil {
				log.Printf("ERROR: Failed to update reservation status after creation for customer %s in event %s: %v", customerID, eventID, updateErr)
				return nil // Don't fail webhook
			}
			
			log.Printf("SUCCESS: Created and updated reservation for customer %s in event %s", customerID, eventID)
			return nil
		}
		return err
	}
	
	return nil
}
