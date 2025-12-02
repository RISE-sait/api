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

func (s *Service) GetExternalTeams(ctx context.Context) ([]values.GetTeamValues, *errLib.CommonError) {
	return s.repo.ListExternal(ctx)
}

func (s *Service) GetInternalTeams(ctx context.Context) ([]values.GetTeamValues, *errLib.CommonError) {
	return s.repo.ListInternal(ctx)
}

func (s *Service) SearchTeamsByName(ctx context.Context, query string, limit int32) ([]values.GetTeamValues, *errLib.CommonError) {
	if limit <= 0 || limit > 50 {
		limit = 20 // Default limit
	}
	return s.repo.SearchByName(ctx, query, limit)
}

func (s *Service) CheckTeamNameExists(ctx context.Context, name string) (bool, *errLib.CommonError) {
	return s.repo.CheckNameExists(ctx, name)
}

func (s *Service) Create(ctx context.Context, details values.CreateTeamValues) *errLib.CommonError {

	var (
		err     *errLib.CommonError
		staffID uuid.UUID
	)

	// Get user role and ID from context
	userID, err := contextUtils.GetUserID(ctx)
	if err != nil {
		return err
	}

	role, err := contextUtils.GetUserRole(ctx)
	if err != nil {
		return err
	}

	// For coaches, ensure they are creating a team for themselves
	if role == contextUtils.RoleCoach {
		if details.Details.CoachID != userID {
			return errLib.New("Coaches can only create teams for themselves", 403)
		}
	}

	// Validate team name doesn't already exist (case-insensitive)
	exists, err := s.CheckTeamNameExists(ctx, details.Details.Name)
	if err != nil {
		return err
	}
	if exists {
		return errLib.New(fmt.Sprintf("A team with the name '%s' already exists", details.Details.Name), 409)
	}

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
			fmt.Sprintf("Created team '%s'", details.Name),
		); err != nil {
			return err
		}

		return nil
	})
}

func (s *Service) CreateExternalTeam(ctx context.Context, details values.CreateExternalTeamValues) *errLib.CommonError {

	var (
		err     *errLib.CommonError
		staffID uuid.UUID
	)

	// Validate team name doesn't already exist (case-insensitive)
	exists, err := s.CheckTeamNameExists(ctx, details.Name)
	if err != nil {
		return err
	}
	if exists {
		return errLib.New(fmt.Sprintf("A team with the name '%s' already exists. Please search for existing teams before creating a new one", details.Name), 409)
	}

	return s.executeInTx(ctx, func(txRepo *repo.Repository) *errLib.CommonError {
		if err = txRepo.CreateExternal(ctx, details); err != nil {
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
			fmt.Sprintf("Created external team: %s", details.Name),
		); err != nil {
			return err
		}

		return nil
	})
}

func (s *Service) UpdateTeam(ctx context.Context, details values.UpdateTeamValues) *errLib.CommonError {

	// Get user role and ID from context
	userID, err := contextUtils.GetUserID(ctx)
	if err != nil {
		return err
	}

	role, err := contextUtils.GetUserRole(ctx)
	if err != nil {
		return err
	}

	// For coaches, validate they own the team they're trying to update
	if role == contextUtils.RoleCoach {
		existingTeam, err := s.repo.GetByID(ctx, details.ID)
		if err != nil {
			return err
		}

		if existingTeam.TeamDetails.CoachID != userID {
			return errLib.New("Coaches can only update their own teams", 403)
		}
	}

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
			fmt.Sprintf("Updated team '%s'", details.TeamDetails.Name),
		)
	})
}

func (s *Service) DeleteTeam(ctx context.Context, id uuid.UUID) *errLib.CommonError {

	// Get user role and ID from context
	userID, err := contextUtils.GetUserID(ctx)
	if err != nil {
		return err
	}

	role, err := contextUtils.GetUserRole(ctx)
	if err != nil {
		return err
	}

	// Fetch team before deletion for audit log and coach validation
	team, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// For coaches, validate they own the team they're trying to delete
	if role == contextUtils.RoleCoach {
		if team.TeamDetails.CoachID != userID {
			return errLib.New("Coaches can only delete their own teams", 403)
		}
	}

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
			fmt.Sprintf("Deleted team '%s'", team.TeamDetails.Name),
		)
	})
}
