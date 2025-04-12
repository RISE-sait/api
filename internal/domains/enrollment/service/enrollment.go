package enrollment

import (
	database_errors "api/internal/constants"
	"api/internal/di"
	repo "api/internal/domains/enrollment/persistence/repository"
	dbEnrollment "api/internal/domains/enrollment/persistence/sqlc/generated"
	errLib "api/internal/libs/errors"
	"context"
	"database/sql"
	"errors"
	"github.com/google/uuid"
	"github.com/lib/pq"
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

	UpdateReservationStatusInEvent(ctx context.Context, eventID, customerID uuid.UUID, status dbEnrollment.PaymentStatus) *errLib.CommonError

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
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == database_errors.TxSerializationError {
			return errLib.New(
				"Too many people enrolled at the same time. Please try again.",
				http.StatusConflict,
			)
		}
		return errLib.New("Failed to commit transaction", http.StatusInternalServerError)
	}
	return nil
}

func (s *CustomerEnrollmentService) EnrollCustomerInProgram(ctx context.Context, customerID, programID uuid.UUID) *errLib.CommonError {

	return s.repo.EnrollCustomerInProgram(ctx, programID, customerID)
}

func (s *CustomerEnrollmentService) EnrollCustomerInEvent(ctx context.Context, customerID, eventID uuid.UUID) *errLib.CommonError {
	return s.repo.EnrollCustomerInEvent(ctx, customerID, eventID)
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
	return s.repo.UpdateReservationStatusInProgram(ctx, programID, customerID, status)
}

func (s *CustomerEnrollmentService) UpdateReservationStatusInEvent(ctx context.Context, eventID, customerID uuid.UUID, status dbEnrollment.PaymentStatus) *errLib.CommonError {
	return s.repo.UpdateReservationStatusInEvent(ctx, eventID, customerID, status)
}
