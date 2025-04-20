package membership

import (
	"api/internal/di"
	staffActivityLogs "api/internal/domains/audit/staff_activity_logs/service"
	repo "api/internal/domains/membership/persistence/repositories"
	values "api/internal/domains/membership/values"
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

type PlanService struct {
	repo                     *repo.PlansRepository
	staffActivityLogsService *staffActivityLogs.Service
	db                       *sql.DB
}

func NewPlanService(container *di.Container) *PlanService {

	return &PlanService{
		repo:                     repo.NewMembershipPlansRepository(container),
		staffActivityLogsService: staffActivityLogs.NewService(container),
		db:                       container.DB,
	}
}

func (s *PlanService) executeInTx(ctx context.Context, fn func(repo *repo.PlansRepository) *errLib.CommonError) *errLib.CommonError {
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
		log.Printf("Failed to commit transaction for membership: %v", err)
		return errLib.New("Failed to commit transaction", http.StatusInternalServerError)
	}
	return nil
}

func (s *PlanService) GetPlan(ctx context.Context, planID uuid.UUID) (values.PlanReadValues, *errLib.CommonError) {

	return s.repo.GetMembershipPlanById(ctx, planID)
}

func (s *PlanService) GetPlans(ctx context.Context, membershipID uuid.UUID) ([]values.PlanReadValues, *errLib.CommonError) {

	return s.repo.GetMembershipPlans(ctx, membershipID)
}

func (s *PlanService) CreateMembershipPlan(ctx context.Context, details values.PlanCreateValues) *errLib.CommonError {

	return s.executeInTx(ctx, func(txRepo *repo.PlansRepository) *errLib.CommonError {
		if err := txRepo.CreateMembershipPlan(ctx, details); err != nil {
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
			fmt.Sprintf("Created membership plan with details: %+v", details),
		)
	})
}

func (s *PlanService) UpdateMembershipPlan(ctx context.Context, details values.PlanUpdateValues) *errLib.CommonError {

	return s.executeInTx(ctx, func(txRepo *repo.PlansRepository) *errLib.CommonError {
		if err := txRepo.UpdateMembershipPlan(ctx, details); err != nil {
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
			fmt.Sprintf("Updated membership plan with ID and new details: %+v", details),
		)
	})
}

func (s *PlanService) DeleteMembershipPlan(ctx context.Context, id uuid.UUID) *errLib.CommonError {

	return s.executeInTx(ctx, func(txRepo *repo.PlansRepository) *errLib.CommonError {
		if err := txRepo.DeleteMembershipPlan(ctx, id); err != nil {
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
			fmt.Sprintf("Deleted membership plan with ID: %s", id),
		)
	})
}
