package membership

import (
	"api/internal/di"
	staffActivityLogs "api/internal/domains/audit/staff_activity_logs/service"
	repo "api/internal/domains/location/persistence"
	values "api/internal/domains/location/values"
	errLib "api/internal/libs/errors"
	contextUtils "api/utils/context"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"log"
	"net/http"
)

type Service struct {
	repo                     *repo.Repository
	staffActivityLogsService *staffActivityLogs.Service
	db                       *sql.DB
}

func NewService(container *di.Container) *Service {

	return &Service{
		repo:                     repo.NewLocationRepository(container),
		staffActivityLogsService: staffActivityLogs.NewService(container),
		db:                       container.DB,
	}
}

func (s *Service) executeInTx(ctx context.Context, fn func(repo *repo.Repository) *errLib.CommonError) *errLib.CommonError {
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{})

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
		log.Printf("Failed to commit transaction for locations: %v", err)
		return errLib.New("Failed to commit transaction", http.StatusInternalServerError)
	}
	return nil
}

func (s *Service) GetLocation(ctx context.Context, id uuid.UUID) (values.ReadValues, *errLib.CommonError) {

	return s.repo.GetLocationByID(ctx, id)
}

func (s *Service) GetLocations(ctx context.Context) ([]values.ReadValues, *errLib.CommonError) {

	return s.repo.GetLocations(ctx)
}

func (s *Service) CreateLocation(ctx context.Context, details values.CreateDetails) (values.ReadValues, *errLib.CommonError) {

	var (
		createdLocation values.ReadValues
		err             *errLib.CommonError
		staffID         uuid.UUID
	)

	err = s.executeInTx(ctx, func(txRepo *repo.Repository) *errLib.CommonError {
		createdLocation, err = txRepo.CreateLocation(ctx, details)

		if err != nil {
			return err
		}

		staffID, err = contextUtils.GetUserID(ctx)

		if err != nil {
			return err
		}

		if err = s.staffActivityLogsService.InsertStaffActivity(
			ctx,
			txRepo.GetTx(),
			staffID,
			fmt.Sprintf("Created location with details: %+v", details),
		); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return values.ReadValues{}, err
	}
	return createdLocation, nil
}

func (s *Service) UpdateLocation(ctx context.Context, details values.UpdateDetails) *errLib.CommonError {

	return s.executeInTx(ctx, func(txRepo *repo.Repository) *errLib.CommonError {
		if err := txRepo.UpdateLocation(ctx, details); err != nil {
			return err
		}

		staffID, err := contextUtils.GetUserID(ctx)
		if err != nil {
			return err
		}

		return s.staffActivityLogsService.InsertStaffActivity(
			ctx,
			txRepo.GetTx(),
			staffID,
			fmt.Sprintf("Updated location with ID and new details: %+v", details),
		)
	})
}

func (s *Service) DeleteLocation(ctx context.Context, id uuid.UUID) *errLib.CommonError {

	return s.executeInTx(ctx, func(txRepo *repo.Repository) *errLib.CommonError {
		if err := txRepo.DeleteLocation(ctx, id); err != nil {
			return err
		}

		staffID, err := contextUtils.GetUserID(ctx)
		if err != nil {
			return err
		}

		return s.staffActivityLogsService.InsertStaffActivity(
			ctx,
			txRepo.GetTx(),
			staffID,
			fmt.Sprintf("Deleted location with ID: %s", id),
		)
	})
}
