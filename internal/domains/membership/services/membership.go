package membership

import (
	"api/internal/di"
	staffActivityLogs "api/internal/domains/audit/staff_activity_logs/service"
	repo "api/internal/domains/membership/persistence/repositories"
	values "api/internal/domains/membership/values"
	errLib "api/internal/libs/errors"
	contextUtils "api/utils/context"
	txUtils "api/utils/db"

	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
)

type Service struct {
	repo                     *repo.Repository
	staffActivityLogsService *staffActivityLogs.Service
	db                       *sql.DB
}

func NewService(container *di.Container) *Service {

	return &Service{
		repo:                     repo.NewMembershipsRepository(container),
		staffActivityLogsService: staffActivityLogs.NewService(container),
		db:                       container.DB,
	}
}

func (s *Service) executeInTx(ctx context.Context, fn func(repo *repo.Repository) *errLib.CommonError) *errLib.CommonError {
	return txUtils.ExecuteInTx(ctx, s.db, func(tx *sql.Tx) *errLib.CommonError {
		return fn(s.repo.WithTx(tx))
	})
}

func (s *Service) GetMembership(ctx context.Context, id uuid.UUID) (values.ReadValues, *errLib.CommonError) {

	return s.repo.GetByID(ctx, id)
}

func (s *Service) GetMemberships(ctx context.Context) ([]values.ReadValues, *errLib.CommonError) {

	return s.repo.List(ctx)
}

func (s *Service) CreateMembership(ctx context.Context, details values.CreateValues) *errLib.CommonError {

	return s.executeInTx(ctx, func(txRepo *repo.Repository) *errLib.CommonError {
		if err := txRepo.Create(ctx, details); err != nil {
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
			fmt.Sprintf("Created membership '%s'", details.Name),
		)
	})
}

func (s *Service) UpdateMembership(ctx context.Context, details values.UpdateValues) *errLib.CommonError {

	return s.executeInTx(ctx, func(txRepo *repo.Repository) *errLib.CommonError {
		if err := txRepo.Update(ctx, details); err != nil {
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
			fmt.Sprintf("Updated membership '%s'", details.Name),
		)
	})
}

func (s *Service) DeleteMembership(ctx context.Context, id uuid.UUID) *errLib.CommonError {

	return s.executeInTx(ctx, func(txRepo *repo.Repository) *errLib.CommonError {
		if err := txRepo.Delete(ctx, id); err != nil {
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
			fmt.Sprintf("Deleted membership with ID: %s", id),
		)
	})
}
