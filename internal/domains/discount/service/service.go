package discount

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"

	"api/internal/di"
	staffActivityLogs "api/internal/domains/audit/staff_activity_logs/service"
	repo "api/internal/domains/discount/persistence/repository"
	values "api/internal/domains/discount/values"
	userRepo "api/internal/domains/user/persistence/repository"
	errLib "api/internal/libs/errors"
	contextUtils "api/utils/context"
	txUtils "api/utils/db"

	"github.com/google/uuid"
	"github.com/stripe/stripe-go/v81"
	"github.com/stripe/stripe-go/v81/coupon"
)

type Service struct {
	repo                     *repo.Repository
	staffActivityLogsService *staffActivityLogs.Service
	db                       *sql.DB
	customerRepo             *userRepo.CustomerRepository
}

func NewService(container *di.Container) *Service {
	return &Service{
		repo:                     repo.NewRepository(container),
		staffActivityLogsService: staffActivityLogs.NewService(container),
		db:                       container.DB,
		customerRepo:             userRepo.NewCustomerRepository(container),
	}
}

func (s *Service) executeInTx(ctx context.Context, fn func(repo *repo.Repository) *errLib.CommonError) *errLib.CommonError {
	return txUtils.ExecuteInTx(ctx, s.db, func(tx *sql.Tx) *errLib.CommonError {
		return fn(s.repo.WithTx(tx))
	})
}

// createStripeCoupon creates a coupon in Stripe based on the discount details
func (s *Service) createStripeCoupon(details values.CreateValues) (*string, *errLib.CommonError) {
	params := &stripe.CouponParams{}

	// Set discount value based on type
	if details.DiscountType == values.TypePercentage {
		params.PercentOff = stripe.Float64(float64(details.DiscountPercent))
	} else if details.DiscountType == values.TypeFixedAmount {
		if details.DiscountAmount == nil {
			return nil, errLib.New("discount_amount is required for fixed_amount type", http.StatusBadRequest)
		}
		// Stripe expects amount in cents
		params.AmountOff = stripe.Int64(int64(*details.DiscountAmount * 100))
		params.Currency = stripe.String("cad") // TODO: Make currency configurable
	}

	// Set duration
	switch details.DurationType {
	case values.DurationOnce:
		params.Duration = stripe.String(string(stripe.CouponDurationOnce))
	case values.DurationRepeating:
		params.Duration = stripe.String(string(stripe.CouponDurationRepeating))
		if details.DurationMonths != nil {
			params.DurationInMonths = stripe.Int64(int64(*details.DurationMonths))
		}
	case values.DurationForever:
		params.Duration = stripe.String(string(stripe.CouponDurationForever))
	}

	// Set max redemptions if specified
	if details.MaxRedemptions != nil {
		params.MaxRedemptions = stripe.Int64(int64(*details.MaxRedemptions))
	}

	// Set name and metadata
	params.Name = stripe.String(details.Name)
	params.Metadata = map[string]string{
		"description": details.Description,
		"applies_to":  string(details.AppliesTo),
	}

	// Create the coupon in Stripe
	c, err := coupon.New(params)
	if err != nil {
		log.Printf("Failed to create Stripe coupon for discount '%s': %v", details.Name, err)
		return nil, errLib.New("Failed to create coupon in Stripe: "+err.Error(), http.StatusInternalServerError)
	}

	log.Printf("Successfully created Stripe coupon %s for discount '%s'", c.ID, details.Name)
	return &c.ID, nil
}

func (s *Service) GetDiscount(ctx context.Context, id uuid.UUID) (values.ReadValues, *errLib.CommonError) {
	return s.repo.GetByID(ctx, id)
}

func (s *Service) GetDiscounts(ctx context.Context) ([]values.ReadValues, *errLib.CommonError) {
	return s.repo.List(ctx)
}

func (s *Service) GetDiscountByNameActive(ctx context.Context, name string) (values.ReadValues, *errLib.CommonError) {
	return s.repo.GetByNameActive(ctx, name)
}

func (s *Service) GetDiscountByName(ctx context.Context, name string) (values.ReadValues, *errLib.CommonError) {
	return s.repo.GetByName(ctx, name)
}

func (s *Service) CreateDiscount(ctx context.Context, details values.CreateValues) (values.ReadValues, *errLib.CommonError) {
	// Create Stripe coupon first
	stripeCouponID, err := s.createStripeCoupon(details)
	if err != nil {
		return values.ReadValues{}, err
	}

	// Store the Stripe coupon ID in the discount details
	details.StripeCouponID = stripeCouponID

	var created values.ReadValues
	txErr := s.executeInTx(ctx, func(txRepo *repo.Repository) *errLib.CommonError {
		var err2 *errLib.CommonError
		created, err2 = txRepo.Create(ctx, details)
		if err2 != nil {
			return err2
		}
		staffID, err := contextUtils.GetUserID(ctx)
		if err != nil {
			return err
		}
		return s.staffActivityLogsService.InsertStaffActivity(
			ctx,
			txRepo.GetTx(),
			staffID,
			fmt.Sprintf("Created discount with details: %+v", details),
		)
	})
	if txErr != nil {
		// If database transaction fails, we should delete the Stripe coupon
		// to avoid orphaned coupons
		if stripeCouponID != nil {
			if _, delErr := coupon.Del(*stripeCouponID, nil); delErr != nil {
				log.Printf("WARNING: Failed to cleanup Stripe coupon %s after database error: %v", *stripeCouponID, delErr)
			} else {
				log.Printf("Cleaned up Stripe coupon %s after database transaction failure", *stripeCouponID)
			}
		}
		return values.ReadValues{}, txErr
	}
	return created, nil
}

func (s *Service) UpdateDiscount(ctx context.Context, details values.UpdateValues) (values.ReadValues, *errLib.CommonError) {
	var updated values.ReadValues
	err := s.executeInTx(ctx, func(txRepo *repo.Repository) *errLib.CommonError {
		var err2 *errLib.CommonError
		updated, err2 = txRepo.Update(ctx, details)
		if err2 != nil {
			return err2
		}
		staffID, err := contextUtils.GetUserID(ctx)
		if err != nil {
			return err
		}
		return s.staffActivityLogsService.InsertStaffActivity(
			ctx,
			txRepo.GetTx(),
			staffID,
			fmt.Sprintf("Updated discount with ID and new details: %+v", details),
		)
	})
	if err != nil {
		return values.ReadValues{}, err
	}
	return updated, nil
}

func (s *Service) DeleteDiscount(ctx context.Context, id uuid.UUID) *errLib.CommonError {
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
			fmt.Sprintf("Deleted discount with ID: %s", id),
		)
	})
}

func (s *Service) GetUsageCount(ctx context.Context, customerID, discountID uuid.UUID) (int32, *errLib.CommonError) {
	return s.repo.GetUsageCount(ctx, customerID, discountID)
}

func (s *Service) IncrementUsage(ctx context.Context, customerID, discountID uuid.UUID) *errLib.CommonError {
	return s.repo.IncrementUsage(ctx, customerID, discountID)
}

// ApplyDiscount validates and records usage of a discount code for the current customer
func (s *Service) ApplyDiscount(ctx context.Context, name string, membershipPlanID *uuid.UUID) (values.ReadValues, *errLib.CommonError) {
	discount, err := s.repo.GetByNameActive(ctx, name)
	if err != nil {
		return values.ReadValues{}, err
	}

	customerID, ctxErr := contextUtils.GetUserID(ctx)
	if ctxErr != nil {
		return values.ReadValues{}, ctxErr
	}

	if membershipPlanID == nil {
		if info, err := s.customerRepo.GetActiveMembershipInfo(ctx, customerID); err == nil {
			membershipPlanID = &info.MembershipPlanID
		}
	}

	if !discount.IsUseUnlimited && discount.UsePerClient > 0 {
		usage, err := s.repo.GetUsageCount(ctx, customerID, discount.ID)
		if err != nil {
			return values.ReadValues{}, err
		}
		if usage >= int32(discount.UsePerClient) {
			return values.ReadValues{}, errLib.New("discount usage limit reached", http.StatusForbidden)
		}
	}

	if membershipPlanID != nil {
		restricted, err := s.repo.GetRestrictedPlans(ctx, discount.ID)
		if err != nil {
			return values.ReadValues{}, err
		}
		if len(restricted) > 0 {
			allowed := false
			for _, pid := range restricted {
				if pid == *membershipPlanID {
					allowed = true
					break
				}
			}
			if !allowed {
				return values.ReadValues{}, errLib.New("discount not valid for this membership plan", http.StatusForbidden)
			}
		}
	}

	if err := s.repo.IncrementUsage(ctx, customerID, discount.ID); err != nil {
		return values.ReadValues{}, err
	}

	return discount, nil
}
