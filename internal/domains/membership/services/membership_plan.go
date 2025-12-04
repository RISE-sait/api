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
	productService           *stripeService.ProductService
	db                       *sql.DB
}

func NewPlanService(container *di.Container) *PlanService {

	return &PlanService{
		repo:                     repo.NewMembershipPlansRepository(container),
		staffActivityLogsService: staffActivityLogs.NewService(container),
		stripeService:            stripeService.NewPriceService(),
		productService:           stripeService.NewProductService(),
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

// GetAllPlansWithStripeData retrieves ALL membership plans (including hidden) and enriches them with live Stripe price data
func (s *PlanService) GetAllPlansWithStripeData(ctx context.Context, membershipID uuid.UUID) ([]values.PlanReadValues, *errLib.CommonError) {
	plans, err := s.repo.GetAllMembershipPlansAdmin(ctx, membershipID)
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

	// If no StripePriceID provided, create product/price in Stripe
	if details.StripePriceID == "" {
		// Set defaults
		currency := details.Currency
		if currency == "" {
			currency = "cad"
		}
		intervalCount := int64(1)
		if details.IntervalCount != nil {
			intervalCount = *details.IntervalCount
		}

		// Create Stripe Product + recurring Price
		priceID, productID, err := s.productService.CreateProductWithRecurringPrice(
			details.Name,
			"",
			*details.UnitAmount,
			currency,
			details.BillingInterval,
			intervalCount,
		)
		if err != nil {
			return err
		}

		details.StripePriceID = priceID
		log.Printf("[STRIPE] Created Stripe product %s with price %s for plan '%s'", productID, priceID, details.Name)

		// If joining fee provided, create one-time price
		if details.JoiningFeeAmount != nil && *details.JoiningFeeAmount > 0 {
			joiningFeePriceID, err := s.productService.CreateOneTimePrice(
				productID,
				*details.JoiningFeeAmount,
				currency,
				"Joining Fee",
			)
			if err != nil {
				return err
			}
			details.StripeJoiningFeesID = joiningFeePriceID
			details.JoiningFee = int(*details.JoiningFeeAmount)
			log.Printf("[STRIPE] Created Stripe joining fee price %s for plan '%s'", joiningFeePriceID, details.Name)
		}
	}

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
			fmt.Sprintf("Created membership plan '%s'", details.Name),
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
			fmt.Sprintf("Updated membership plan '%s'", details.Name),
		)
	})
}

func (s *PlanService) DeleteMembershipPlan(ctx context.Context, id uuid.UUID) *errLib.CommonError {

	// First, get the plan to retrieve Stripe IDs before deleting
	plan, err := s.repo.GetMembershipPlanById(ctx, id)
	if err != nil {
		return err
	}

	// Check if there are any active customer memberships for this plan
	var activeCount int
	checkActiveQuery := `
		SELECT COUNT(*) FROM users.customer_membership_plans
		WHERE membership_plan_id = $1 AND status = 'active'
	`
	if dbErr := s.db.QueryRowContext(ctx, checkActiveQuery, id).Scan(&activeCount); dbErr != nil {
		log.Printf("Failed to check active memberships for plan %s: %v", id, dbErr)
		return errLib.New("Failed to check active memberships", 500)
	}

	if activeCount > 0 {
		return errLib.New(fmt.Sprintf("Cannot delete plan: %d active customer memberships exist. Wait for them to expire or cancel them first.", activeCount), 400)
	}

	// Deactivate Stripe product and prices (don't fail if Stripe fails)
	if plan.StripePriceID != "" {
		s.productService.DeactivatePrice(plan.StripePriceID)
		s.productService.DeactivateProductFromPrice(plan.StripePriceID)
	}
	if plan.StripeJoiningFeesID != "" {
		s.productService.DeactivatePrice(plan.StripeJoiningFeesID)
	}

	return s.executeInTx(ctx, func(txRepo *repo.PlansRepository) *errLib.CommonError {
		// Delete expired customer memberships for this plan first
		deleteExpiredQuery := `
			DELETE FROM users.customer_membership_plans
			WHERE membership_plan_id = $1 AND status != 'active'
		`
		if _, dbErr := txRepo.GetTx().ExecContext(ctx, deleteExpiredQuery, id); dbErr != nil {
			log.Printf("Failed to delete expired memberships for plan %s: %v", id, dbErr)
			return errLib.New("Failed to delete expired customer memberships", 500)
		}

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
			fmt.Sprintf("Deleted membership plan '%s' (Stripe product deactivated, expired enrollments removed)", plan.Name),
		)
	})
}

func (s *PlanService) ToggleMembershipPlanVisibility(ctx context.Context, id uuid.UUID, isVisible bool) (values.PlanReadValues, *errLib.CommonError) {

	var plan values.PlanReadValues

	err := s.executeInTx(ctx, func(txRepo *repo.PlansRepository) *errLib.CommonError {
		result, err := txRepo.ToggleMembershipPlanVisibility(ctx, id, isVisible)
		if err != nil {
			return err
		}

		plan = result

		staffID, err := contextUtils.GetUserID(ctx)
		if err != nil {
			return err
		}

		visibilityStatus := "hidden"
		if isVisible {
			visibilityStatus = "visible"
		}

		return s.staffActivityLogsService.InsertStaffActivity(
			ctx,
			txRepo.GetTx(),
			staffID,
			fmt.Sprintf("Set membership plan %s visibility to %s", id, visibilityStatus),
		)
	})

	return plan, err
}
