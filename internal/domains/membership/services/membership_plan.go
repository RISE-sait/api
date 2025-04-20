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
	return txUtils.ExecuteInTx(ctx, s.db, func(tx *sql.Tx) *errLib.CommonError {
		return fn(s.repo.WithTx(tx))
	})
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
