package program

import (
	db "api/cmd/seed/sqlc/generated"
	"api/internal/di"
	staffActivityLogs "api/internal/domains/audit/staff_activity_logs/service"
	repo "api/internal/domains/program/persistence"
	"api/internal/domains/program/values"
	errLib "api/internal/libs/errors"
	contextUtils "api/utils/context"
	txUtils "api/utils/db"
	"context"
	"database/sql"
	"fmt"
	"net/http"

	"github.com/google/uuid"
)

type Service struct {
	repo                     *repo.Repository
	staffActivityLogsService *staffActivityLogs.Service
	db                       *sql.DB
}

func NewProgramService(container *di.Container) *Service {

	return &Service{
		repo:                     repo.NewProgramRepository(container),
		staffActivityLogsService: staffActivityLogs.NewService(container),
		db:                       container.DB,
	}
}

func (s *Service) executeInTx(ctx context.Context, fn func(repo *repo.Repository) *errLib.CommonError) *errLib.CommonError {
	return txUtils.ExecuteInTx(ctx, s.db, func(tx *sql.Tx) *errLib.CommonError {
		return fn(s.repo.WithTx(tx))
	})
}

func (s *Service) GetProgram(ctx context.Context, programID uuid.UUID) (values.GetProgramValues, *errLib.CommonError) {

	return s.repo.GetProgramByID(ctx, programID)
}

func (s *Service) GetPrograms(ctx context.Context, programType string) ([]values.GetProgramValues, *errLib.CommonError) {

	return s.repo.List(ctx, programType)
}

func (s *Service) validateProgramType(inputType string) *errLib.CommonError {

	if !db.ProgramProgramType(inputType).Valid() {
		validTypes := db.AllProgramProgramTypeValues()
		return errLib.New(fmt.Sprintf("Invalid program type. Valid types are: %v", validTypes), http.StatusBadRequest)
	}

	return nil
}

func (s *Service) CreateProgram(ctx context.Context, details values.CreateProgramValues) *errLib.CommonError {

	if err := s.validateProgramType(details.Type); err != nil {
		return err
	}

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
			fmt.Sprintf("Created program '%s' (type: %s)", details.Name, details.Type),
		)
	})
}

func (s *Service) UpdateProgram(ctx context.Context, details values.UpdateProgramValues) *errLib.CommonError {

	if err := s.validateProgramType(details.Type); err != nil {
		return err
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
			fmt.Sprintf("Updated program '%s' (type: %s)", details.Name, details.Type),
		)
	})
}

func (s *Service) DeleteProgram(ctx context.Context, id uuid.UUID) *errLib.CommonError {

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
			fmt.Sprintf("Deleted program with ID: %s", id),
		)
	})
}
