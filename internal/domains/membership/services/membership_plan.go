package membership

import (
	"api/internal/di"
	staffActivityLogs "api/internal/domains/audit/staff_activity_logs/service"
	repo "api/internal/domains/membership/persistence/repositories"
	stripeService "api/internal/domains/payment/services/stripe"
	values "api/internal/domains/membership/values"
	errLib "api/internal/libs/errors"
	contextUtils "api/utils/context"
	txUtils "api/utils/db"
	"context"
	"database/sql"
	"fmt"
	"log"

	"github.com/google/uuid"
)

type PlanService struct {
	repo                     *repo.PlansRepository
	staffActivityLogsService *staffActivityLogs.Service
	stripeService            *stripeService.PriceService
	db                       *sql.DB
}

func NewPlanService(container *di.Container) *PlanService {

	return &PlanService{
		repo:                     repo.NewMembershipPlansRepository(container),
		staffActivityLogsService: staffActivityLogs.NewService(container),
		stripeService:            stripeService.NewPriceService(),
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

// GetPlansWithStripeData retrieves membership plans and enriches them with live Stripe price data
func (s *PlanService) GetPlansWithStripeData(ctx context.Context, membershipID uuid.UUID) ([]values.PlanReadValues, *errLib.CommonError) {
	plans, err := s.repo.GetMembershipPlans(ctx, membershipID)
	if err != nil {
		return nil, err
	}

	// Enrich plans with live Stripe price data
	for i, plan := range plans {
		if plan.StripePriceID != "" {
			stripePrice, stripeErr := s.stripeService.GetPrice(plan.StripePriceID)
			if stripeErr != nil {
				log.Printf("Warning: Failed to fetch Stripe price for plan %s (price_id: %s): %v", 
					plan.ID, plan.StripePriceID, stripeErr)
				// Continue with database values if Stripe fails
				continue
			}

			// Update plan with live Stripe data
			plans[i].UnitAmount = int(stripePrice.UnitAmount)
			plans[i].Currency = string(stripePrice.Currency)
			plans[i].Interval = string(stripePrice.Recurring.Interval)
		}

		// Fetch joining fee price from Stripe
		if plan.StripeJoiningFeesID != "" {
			joiningFeePrice, joiningFeeErr := s.stripeService.GetPrice(plan.StripeJoiningFeesID)
			if joiningFeeErr != nil {
				log.Printf("Warning: Failed to fetch Stripe joining fee price for plan %s (price_id: %s): %v", 
					plan.ID, plan.StripeJoiningFeesID, joiningFeeErr)
				// Continue with database values if Stripe fails
				continue
			}

			// Update plan with live Stripe joining fee data
			plans[i].JoiningFee = int(joiningFeePrice.UnitAmount)
		}
	}

	return plans, nil
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
