package enrollment

import (
	"api/internal/di"
	repo "api/internal/domains/enrollment/persistence/repository"
	dbEnrollment "api/internal/domains/enrollment/persistence/sqlc/generated"
	errLib "api/internal/libs/errors"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"log"
	"net/http"
	"time"
)

type ICustomerEnrollmentService interface {
	EnrollCustomerInProgram(
		ctx context.Context,
		customerID uuid.UUID,
		programID uuid.UUID,
	) *errLib.CommonError

	EnrollCustomerInMembershipPlan(
		ctx context.Context,
		customerID uuid.UUID,
		planID uuid.UUID,
		cancelAtDateTime time.Time,
	) *errLib.CommonError

	GetProgramIsFull(ctx context.Context, programID uuid.UUID) (bool, *errLib.CommonError)

	GetEventIsFull(ctx context.Context, eventID uuid.UUID) (bool, *errLib.CommonError)

	ReserveSeatInProgram(ctx context.Context, programID, customerID uuid.UUID) *errLib.CommonError

	UpdateReservationStatusInProgram(ctx context.Context, programID, customerID uuid.UUID, status dbEnrollment.PaymentStatus) *errLib.CommonError

	ReserveSeatInEvent(ctx context.Context, eventID, customerID uuid.UUID) *errLib.CommonError

	UnEnrollCustomerFromEvent(ctx context.Context, eventID, customerID uuid.UUID) *errLib.CommonError
}

var _ ICustomerEnrollmentService = (*CustomerEnrollmentService)(nil)

type CustomerEnrollmentService struct {
	repo *repo.CustomerEnrollmentRepository
	db   *sql.DB
}

func NewCustomerEnrollmentService(container *di.Container) *CustomerEnrollmentService {
	return &CustomerEnrollmentService{
		repo: repo.NewEnrollmentRepository(container.DB),
		db:   container.DB,
	}
}

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
		log.Printf("Transaction error: %v", err)
		return errLib.New("Failed to commit transaction", http.StatusInternalServerError)
	}
	return nil
}

func (s *CustomerEnrollmentService) EnrollCustomerInProgram(ctx context.Context, customerID, programID uuid.UUID) *errLib.CommonError {
	return s.executeInTx(ctx, func(r *repo.CustomerEnrollmentRepository) *errLib.CommonError {

		isFull, err := r.GetProgramIsFull(ctx, programID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return errLib.New("Program or capacity for program not found", http.StatusNotFound)
			}
			return err
		}
		if isFull {
			return errLib.New("Program is full", http.StatusConflict)
		}
		if err = r.EnrollCustomerInProgram(ctx, customerID, programID); err != nil {
			return handleDatabaseError(err)
		}

		return nil
	})
}

func (s *CustomerEnrollmentService) EnrollCustomerInEvent(ctx context.Context, customerID, eventID uuid.UUID) *errLib.CommonError {
	return s.executeInTx(ctx, func(r *repo.CustomerEnrollmentRepository) *errLib.CommonError {
		isFull, err := r.GetEventIsFull(ctx, eventID)
		if err != nil {
			return err
		}
		if isFull {
			return errLib.New("Event is full", http.StatusConflict)
		}

		if err = r.EnrollCustomerInEvent(ctx, customerID, eventID); err != nil {
			return handleDatabaseError(err)
		}
		return nil
	})
}

func (s *CustomerEnrollmentService) EnrollCustomerInMembershipPlan(ctx context.Context, customerID, planID uuid.UUID, cancelAtDateTime time.Time) *errLib.CommonError {
	return s.repo.EnrollCustomerInMembershipPlan(ctx, customerID, planID, cancelAtDateTime)
}

func (s *CustomerEnrollmentService) UnEnrollCustomerFromEvent(ctx context.Context, eventID, customerID uuid.UUID) *errLib.CommonError {
	return s.repo.UnEnrollCustomerFromEvent(ctx, eventID, customerID)
}

func (s *CustomerEnrollmentService) GetProgramIsFull(ctx context.Context, programID uuid.UUID) (bool, *errLib.CommonError) {
	return s.repo.GetProgramIsFull(ctx, programID)
}

func (s *CustomerEnrollmentService) GetEventIsFull(ctx context.Context, eventID uuid.UUID) (bool, *errLib.CommonError) {
	return s.repo.GetEventIsFull(ctx, eventID)
}

func (s *CustomerEnrollmentService) ReserveSeatInEvent(ctx context.Context, eventID, customerID uuid.UUID) *errLib.CommonError {
	return s.executeInTx(ctx, func(r *repo.CustomerEnrollmentRepository) *errLib.CommonError {
		isFull, err := r.GetEventIsFull(ctx, eventID)
		if err != nil {
			return err
		}
		if isFull {
			return errLib.New("Event is full", http.StatusConflict)
		}
		if err = r.ReserveSeatInEvent(ctx, customerID, eventID); err != nil {
			return handleDatabaseError(err)
		}
		return nil
	})
}

func (s *CustomerEnrollmentService) ReserveSeatInProgram(ctx context.Context, programID, customerID uuid.UUID) *errLib.CommonError {
	return s.executeInTx(ctx, func(r *repo.CustomerEnrollmentRepository) *errLib.CommonError {
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
	return s.executeInTx(ctx, func(r *repo.CustomerEnrollmentRepository) *errLib.CommonError {

		if err := r.UpdateReservationStatusInProgram(ctx, customerID, programID, status); err != nil {
			return handleDatabaseError(err)
		}
		return nil
	})
}

func handleDatabaseError(err error) *errLib.CommonError {
	if errors.Is(err, sql.ErrNoRows) {
		return errLib.New("Resource not found", http.StatusNotFound)
	}
	return errLib.New(fmt.Sprintf("Database error: %v", err), http.StatusInternalServerError)
}
