package court

import (
	"context"
	"database/sql"
	"fmt"

	"api/internal/di"
	staffActivityLogs "api/internal/domains/audit/staff_activity_logs/service"
	repo "api/internal/domains/court/persistence"
	values "api/internal/domains/court/values"
	errLib "api/internal/libs/errors"
	contextUtils "api/utils/context"
	txUtils "api/utils/db"

	"github.com/google/uuid"
)

type Service struct {
	repo                     *repo.Repository
	staffActivityLogsService *staffActivityLogs.Service
	db                       *sql.DB
}

func NewService(container *di.Container) *Service {
	return &Service{
		repo:                     repo.NewRepository(container),
		staffActivityLogsService: staffActivityLogs.NewService(container),
		db:                       container.DB,
	}
}

func (s *Service) executeInTx(ctx context.Context, fn func(r *repo.Repository) *errLib.CommonError) *errLib.CommonError {
	return txUtils.ExecuteInTx(ctx, s.db, func(tx *sql.Tx) *errLib.CommonError {
		return fn(s.repo.WithTx(tx))
	})
}

func (s *Service) CreateCourt(ctx context.Context, d values.CreateDetails) (values.ReadValues, *errLib.CommonError) {
	var created values.ReadValues
	err := s.executeInTx(ctx, func(r *repo.Repository) *errLib.CommonError {
		var err2 *errLib.CommonError
		created, err2 = r.Create(ctx, d)
		if err2 != nil {
			return err2
		}
		staffID, err2 := contextUtils.GetUserID(ctx)
		if err2 != nil {
			return err2
		}
		return s.staffActivityLogsService.InsertStaffActivity(ctx, r.GetTx(), staffID, fmt.Sprintf("Created court %s", d.Name))
	})
	if err != nil {
		return values.ReadValues{}, err
	}
	return created, nil
}

func (s *Service) GetCourt(ctx context.Context, id uuid.UUID) (values.ReadValues, *errLib.CommonError) {
	return s.repo.Get(ctx, id)
}

func (s *Service) GetCourts(ctx context.Context) ([]values.ReadValues, *errLib.CommonError) {
	return s.repo.List(ctx)
}

func (s *Service) UpdateCourt(ctx context.Context, d values.UpdateDetails) *errLib.CommonError {
	return s.executeInTx(ctx, func(r *repo.Repository) *errLib.CommonError {
		if err := r.Update(ctx, d); err != nil {
			return err
		}
		staffID, err := contextUtils.GetUserID(ctx)
		if err != nil {
			return err
		}
		return s.staffActivityLogsService.InsertStaffActivity(ctx, r.GetTx(), staffID, fmt.Sprintf("Updated court '%s'", d.Name))
	})
}

func (s *Service) DeleteCourt(ctx context.Context, id uuid.UUID) *errLib.CommonError {
	// Get court name before deletion for audit log
	court, getErr := s.repo.Get(ctx, id)
	courtName := id.String() // fallback to ID if court not found
	if getErr == nil {
		courtName = court.Name
	}

	return s.executeInTx(ctx, func(r *repo.Repository) *errLib.CommonError {
		if err := r.Delete(ctx, id); err != nil {
			return err
		}
		staffID, err := contextUtils.GetUserID(ctx)
		if err != nil {
			return err
		}
		return s.staffActivityLogsService.InsertStaffActivity(ctx, r.GetTx(), staffID, fmt.Sprintf("Deleted court '%s'", courtName))
	})
}