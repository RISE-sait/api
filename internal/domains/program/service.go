package program

import (
	db "api/cmd/seed/sqlc/generated"
	"api/internal/di"
	staffActivityLogs "api/internal/domains/audit/staff_activity_logs/service"
	repo "api/internal/domains/program/persistence"
	"api/internal/domains/program/values"
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

func NewProgramService(container *di.Container) *Service {

	return &Service{
		repo:                     repo.NewProgramRepository(container),
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
		log.Printf("Failed to commit transaction for program: %v", err)
		return errLib.New("Failed to commit transaction", http.StatusInternalServerError)
	}
	return nil
}

func (s *Service) GetProgram(ctx context.Context, programID uuid.UUID) (values.GetProgramValues, *errLib.CommonError) {

	return s.repo.GetProgramByID(ctx, programID)
}

func (s *Service) GetPrograms(ctx context.Context, programType string) ([]values.GetProgramValues, *errLib.CommonError) {

	return s.repo.List(ctx, programType)
}

func (s *Service) GetProgramLevels() []string {

	return s.repo.GetProgramLevels()
}

func (s *Service) validateProgramLevels(inputLevel string) *errLib.CommonError {

	programLevels := s.repo.GetProgramLevels()

	isValidLevel := false

	for _, validLevel := range programLevels {

		log.Println("Validating program level:", validLevel)
		if inputLevel == validLevel {
			isValidLevel = true
			break
		}
	}

	if !isValidLevel {
		return errLib.New(fmt.Sprintf("Invalid program level. Valid levels are: %v", programLevels), http.StatusBadRequest)
	}

	return nil
}

func (s *Service) validateProgramType(inputType string) *errLib.CommonError {

	if !db.ProgramProgramType(inputType).Valid() {
		validTypes := db.AllProgramProgramTypeValues()
		return errLib.New(fmt.Sprintf("Invalid program type. Valid types are: %v", validTypes), http.StatusBadRequest)
	}

	return nil
}

func (s *Service) CreateProgram(ctx context.Context, details values.CreateProgramValues) *errLib.CommonError {

	if err := s.validateProgramLevels(details.Level); err != nil {
		return err
	}

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
			fmt.Sprintf("Created program with details: %+v", details),
		)
	})
}

func (s *Service) UpdateProgram(ctx context.Context, details values.UpdateProgramValues) *errLib.CommonError {

	if err := s.validateProgramLevels(details.Level); err != nil {
		return err
	}

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
			fmt.Sprintf("Updated program with ID and new details: %+v", details),
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
