package team

import (
	"api/internal/di"
	staffActivityLogs "api/internal/domains/audit/staff_activity_logs/service"
	repo "api/internal/domains/team/persistence"
	values "api/internal/domains/team/values"
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
		repo:                     repo.NewTeamRepository(container),
		staffActivityLogsService: staffActivityLogs.NewService(container),
		db:                       container.DB,
	}
}

func (s *Service) executeInTx(ctx context.Context, fn func(repo *repo.Repository) *errLib.CommonError) *errLib.CommonError {
	return txUtils.ExecuteInTx(ctx, s.db, func(tx *sql.Tx) *errLib.CommonError {
		return fn(s.repo.WithTx(tx))
	})
}

func (s *Service) GetTeamByID(ctx context.Context, id uuid.UUID) (values.GetTeamValues, *errLib.CommonError) {

	return s.repo.GetByID(ctx, id)
}

func (s *Service) GetTeams(ctx context.Context) ([]values.GetTeamValues, *errLib.CommonError) {

	return s.repo.List(ctx)
}

func (s *Service) GetTeamsByCoach(ctx context.Context, coachID uuid.UUID) ([]values.GetTeamValues, *errLib.CommonError) {

	return s.repo.ListByCoach(ctx, coachID)
}

func (s *Service) Create(ctx context.Context, details values.CreateTeamValues) *errLib.CommonError {

	var (
		err     *errLib.CommonError
		staffID uuid.UUID
	)

	return s.executeInTx(ctx, func(txRepo *repo.Repository) *errLib.CommonError {
		if err = txRepo.Create(ctx, details); err != nil {
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
			fmt.Sprintf("Created team with details: %+v", details),
		); err != nil {
			return err
		}

		return nil
	})
}

func (s *Service) UpdateTeam(ctx context.Context, details values.UpdateTeamValues) *errLib.CommonError {

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
			fmt.Sprintf("Updated team with ID and new details: %+v", details),
		)
	})
}

func (s *Service) DeleteTeam(ctx context.Context, id uuid.UUID) *errLib.CommonError {

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
			fmt.Sprintf("Deleted team with ID: %s", id),
		)
	})
}
