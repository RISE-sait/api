package payment

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"

	"api/internal/di"
	enrollment "api/internal/domains/enrollment/service"
	dbEnrollment "api/internal/domains/enrollment/persistence/sqlc/generated"
	eventService "api/internal/domains/event/service"
	membership "api/internal/domains/membership/persistence/repositories"
	repository "api/internal/domains/payment/persistence/repositories"
	"api/internal/domains/payment/services/stripe"
	subsidyService "api/internal/domains/subsidy/service"
	userServices "api/internal/domains/user/services"
	errLib "api/internal/libs/errors"
	contextUtils "api/utils/context"
	discountService "api/internal/domains/discount/service"
	"github.com/google/uuid"
)

type Service struct {
	CheckoutRepo        *repository.CheckoutRepository
	MembershipPlansRepo *membership.PlansRepository
	DiscountService     *discountService.Service
	SubsidyService      *subsidyService.SubsidyService
	EnrollmentService   *enrollment.CustomerEnrollmentService
	EventService        *eventService.Service
	CreditService       *userServices.CustomerCreditService
	DB                  *sql.DB
}

func NewPurchaseService(container *di.Container) *Service {
	return &Service{
		CheckoutRepo:        repository.NewCheckoutRepository(container),
		MembershipPlansRepo: membership.NewMembershipPlansRepository(container),
		DiscountService:     discountService.NewService(container),
		SubsidyService:      subsidyService.NewSubsidyService(container),
		EnrollmentService:   enrollment.NewCustomerEnrollmentService(container),
		EventService:        eventService.NewEventService(container),
		CreditService:       userServices.NewCustomerCreditService(container),
		DB:                  container.DB,
	}
}

// getExistingStripeCustomerID retrieves the existing Stripe customer ID for a user from the database
func (s *Service) getExistingStripeCustomerID(ctx context.Context, userID uuid.UUID) *string {
	var stripeCustomerID sql.NullString
	query := "SELECT stripe_customer_id FROM users.users WHERE id = $1"
	err := s.DB.QueryRowContext(ctx, query, userID).Scan(&stripeCustomerID)
	if err != nil || !stripeCustomerID.Valid || stripeCustomerID.String == "" {
		return nil
	}
	return &stripeCustomerID.String
}

// queueFailedRefund inserts a failed refund into the recovery queue for manual resolution
func (s *Service) queueFailedRefund(ctx context.Context, customerID, eventID uuid.UUID, creditAmount int32, refundErr *errLib.CommonError) {
	log.Printf("[REFUND_QUEUE] Queueing failed refund: customer=%s, event=%s, amount=%d, error=%s",
		customerID, eventID, creditAmount, refundErr.Error())

	query := `INSERT INTO payment.failed_refunds (customer_id, event_id, credit_amount, error_message, status)
		VALUES ($1, $2, $3, $4, 'pending')`

	_, dbErr := s.DB.ExecContext(ctx, query, customerID, eventID, creditAmount, refundErr.Error())
	if dbErr != nil {
		// If we can't even queue the refund, log it prominently for manual intervention
		log.Printf("[REFUND_QUEUE] CRITICAL: Failed to queue refund for recovery: customer=%s, event=%s, amount=%d, queue_error=%v",
			customerID, eventID, creditAmount, dbErr)
	} else {
		log.Printf("[REFUND_QUEUE] Successfully queued failed refund for recovery: customer=%s, event=%s, amount=%d",
			customerID, eventID, creditAmount)
	}
}

func (s *Service) CheckoutMembershipPlan(ctx context.Context, membershipPlanID uuid.UUID, discountCode *string, successURL string) (string, *errLib.CommonError) {
	// Get customer ID from context
	customerID, ctxErr := contextUtils.GetUserID(ctx)
	if ctxErr != nil {
		return "", ctxErr
	}

	// SECURITY: Check if customer already has an active membership for this plan
	hasActiveMembership, err := s.CheckoutRepo.CheckCustomerHasActiveMembership(ctx, customerID, membershipPlanID)
	if err != nil {
		return "", err
	}

	if hasActiveMembership {
		return "", errLib.New("Customer already has an active membership for this plan", http.StatusConflict)
	}

	requirements, err := s.CheckoutRepo.GetMembershipPlanJoiningRequirement(ctx, membershipPlanID)
	if err != nil {
		return "", err
	}

	// Handle discount code if provided
	var stripeCouponID *string
	if discountCode != nil {
		applied, err := s.DiscountService.ApplyDiscount(ctx, *discountCode, &membershipPlanID)
		if err != nil {
			return "", err
		}

		// Validate that discount applies to subscriptions
		if applied.AppliesTo != "subscription" && applied.AppliesTo != "both" {
			return "", errLib.New("This discount code does not apply to subscriptions", http.StatusBadRequest)
		}

		stripeCouponID = applied.StripeCouponID
	}

	// Check if customer has active subsidy
	subsidy, subsidyErr := s.SubsidyService.GetActiveSubsidy(ctx, customerID)
	if subsidyErr != nil {
		// Log error but don't fail checkout - subsidy is optional
		log.Printf("Warning: Failed to check subsidy for customer %s: %v", customerID, subsidyErr)
	}

	// Add metadata for webhook processing
	metadata := map[string]string{
		"userID":           customerID.String(),
		"membershipPlanID": membershipPlanID.String(),
	}

	if subsidy != nil && subsidy.RemainingBalance > 0 {
		metadata["has_subsidy"] = "true"
		metadata["subsidy_id"] = subsidy.ID.String()
		metadata["subsidy_balance"] = fmt.Sprintf("%.2f", subsidy.RemainingBalance)
		log.Printf("Customer %s has active subsidy: $%.2f remaining", customerID, subsidy.RemainingBalance)

		// Create a Stripe coupon for the subsidy amount to apply at checkout time
		// This avoids race conditions with webhooks
		subsidyCouponID, couponErr := stripe.CreateSubsidyCoupon(ctx, subsidy.RemainingBalance)
		if couponErr != nil {
			log.Printf("Warning: Failed to create subsidy coupon: %v", couponErr)
		} else if subsidyCouponID != "" {
			// Apply subsidy coupon (overrides any discount code if both exist)
			stripeCouponID = &subsidyCouponID
			log.Printf("Created subsidy coupon: %s for $%.2f", subsidyCouponID, subsidy.RemainingBalance)
		}
	}

	// Get existing Stripe customer ID or recover if deleted (industry standard: one user = one Stripe customer)
	existingCustomerID, recoverErr := s.getOrRecreateStripeCustomer(ctx, customerID)
	if recoverErr != nil {
		return "", recoverErr
	}

	// Check if membership has any joining fee
	if requirements.StripeJoiningFeeID != "" {
		// Use recurring joining fee (annual/monthly) - existing function handles this
		return stripe.CreateSubscriptionWithMetadata(ctx, requirements.StripePriceID, requirements.StripeJoiningFeeID, stripeCouponID, metadata, successURL, existingCustomerID)
	} else if requirements.JoiningFee > 0 {
		// Use one-time setup fee
		return stripe.CreateSubscriptionWithSetupFeeAndMetadata(ctx, requirements.StripePriceID, requirements.JoiningFee, metadata, successURL, existingCustomerID)
	} else {
		// No joining fee - just regular subscription
		return stripe.CreateSubscriptionWithMetadata(ctx, requirements.StripePriceID, "", stripeCouponID, metadata, successURL, existingCustomerID)
	}
}

func (s *Service) CheckoutProgram(ctx context.Context, programID uuid.UUID, discountCode *string, successURL string) (string, *errLib.CommonError) {
	customerID, ctxErr := contextUtils.GetUserID(ctx)
	if ctxErr != nil {
		return "", ctxErr
	}

	isPayPerEvent, priceID, err := s.CheckoutRepo.GetRegistrationPriceIdForCustomerByProgramID(ctx, programID)
	if err != nil {
		return "", err
	}

	if isPayPerEvent {
		return "", errLib.New("program is not pay-per-event", http.StatusBadRequest)
	}

	// Handle discount code if provided
	var stripeCouponID *string
	if discountCode != nil {
		applied, err := s.DiscountService.ApplyDiscount(ctx, *discountCode, nil)
		if err != nil {
			return "", err
		}

		// Validate that discount applies to one-time payments
		if applied.AppliesTo != "one_time" && applied.AppliesTo != "both" {
			return "", errLib.New("This discount code does not apply to one-time payments", http.StatusBadRequest)
		}

		stripeCouponID = applied.StripeCouponID
	}

	// reserve seat so that the database can assume that the customer is enrolled
	// this is important for the enrollment process, as multiple users may try to enroll one after another,
	// and each successfully getting the stripe checkout link
	// and cause overbooking cuz the database is not aware of the amount of people that are able to pay

	// and the reservation is only valid for 10 minutes
	// so if the customer does not pay in 10 minutes, the reservation will be deemed cancelled
	err = s.EnrollmentService.ReserveSeatInProgram(ctx, programID, customerID)
	if err != nil {
		log.Println("Failed to reserve seat in program:", err)
		return "", err
	}

	// Get existing Stripe customer ID to reuse (industry standard: one user = one Stripe customer)
	existingCustomerID := s.getExistingStripeCustomerID(ctx, customerID)

	programIDStr := programID.String()
	return stripe.CreateOneTimePayment(ctx, priceID, 1, &programIDStr, stripeCouponID, successURL, existingCustomerID)
}

func (s *Service) CheckoutEvent(ctx context.Context, eventID uuid.UUID, discountCode *string, successURL string) (string, *errLib.CommonError) {
	customerID, ctxErr := contextUtils.GetUserID(ctx)
	if ctxErr != nil {
		return "", ctxErr
	}

	programID, err := s.CheckoutRepo.GetProgramIDOfEvent(ctx, eventID)
	if err != nil {
		return "", err
	}

	isPayPerEvent, priceID, err := s.CheckoutRepo.GetRegistrationPriceIdForCustomerByProgramID(ctx, programID)
	if err != nil {
		return "", err
	}

	if !isPayPerEvent {
		return "", errLib.New("event is pay-per-program", http.StatusBadRequest)
	}

	// Handle discount code if provided
	var stripeCouponID *string
	if discountCode != nil {
		applied, err := s.DiscountService.ApplyDiscount(ctx, *discountCode, nil)
		if err != nil {
			return "", err
		}

		// Validate that discount applies to one-time payments
		if applied.AppliesTo != "one_time" && applied.AppliesTo != "both" {
			return "", errLib.New("This discount code does not apply to one-time payments", http.StatusBadRequest)
		}

		stripeCouponID = applied.StripeCouponID
	}

	// reserve seat so that the database can assume that the customer is enrolled
	// this is important for the enrollment process, as multiple users may try to enroll one after another,
	// and each successfully getting the stripe checkout link
	// and cause overbooking cuz the database is not aware of the amount of people that are able to pay

	// and the reservation is only valid for 10 minutes
	// so if the customer does not pay in 10 minutes, the reservation will be deemed cancelled
	if err = s.EnrollmentService.ReserveSeatInEvent(ctx, eventID, customerID); err != nil {
		return "", err
	}

	// Get existing Stripe customer ID to reuse (industry standard: one user = one Stripe customer)
	existingCustomerID := s.getExistingStripeCustomerID(ctx, customerID)

	eventIDStr := eventID.String()
	return stripe.CreateOneTimePayment(ctx, priceID, 1, &eventIDStr, stripeCouponID, successURL, existingCustomerID)
}

// CheckEventEnrollmentOptions returns available enrollment options for a customer and event
type EventEnrollmentOptions struct {
	CanEnrollFree    bool    `json:"can_enroll_free"`
	StripePriceID    *string `json:"stripe_price_id"`
	CreditCost       *int32  `json:"credit_cost"`
	MembershipPlanID *string `json:"membership_plan_id"`
	HasSufficientCredits bool `json:"has_sufficient_credits"`
}

func (s *Service) CheckEventEnrollmentOptions(ctx context.Context, eventID uuid.UUID) (*EventEnrollmentOptions, *errLib.CommonError) {
	customerID, ctxErr := contextUtils.GetUserID(ctx)
	if ctxErr != nil {
		return nil, ctxErr
	}

	// Get event details including membership requirements
	event, err := s.EventService.GetEvent(ctx, eventID)
	if err != nil {
		return nil, err
	}

	options := &EventEnrollmentOptions{}

	// Check if event has required membership plans
	if len(event.RequiredMembershipPlanIDs) > 0 {
		// Event requires membership - check if customer has ANY of the required memberships
		hasAccess, err := s.CheckoutRepo.CheckCustomerHasEventMembershipAccess(ctx, eventID, customerID)
		if err != nil {
			return nil, err
		}

		if hasAccess {
			// Customer has one of the required memberships, can enroll for free
			options.CanEnrollFree = true
			// Store the first membership plan ID for reference (optional)
			if len(event.RequiredMembershipPlanIDs) > 0 {
				membershipPlanID := event.RequiredMembershipPlanIDs[0].String()
				options.MembershipPlanID = &membershipPlanID
			}
			return options, nil
		}
		// Customer doesn't have any required membership, need to check payment options
	} else {
		// No membership requirement, check if event is free
		if event.PriceID == nil && event.CreditCost == nil {
			// Event is completely free
			options.CanEnrollFree = true
			return options, nil
		}
	}

	// Event requires payment - check available payment methods
	if event.PriceID != nil {
		options.StripePriceID = event.PriceID
	}
	
	if event.CreditCost != nil {
		options.CreditCost = event.CreditCost
		
		// Check if customer has sufficient credits
		err := s.CreditService.ValidateEventCreditPayment(ctx, eventID, customerID)
		options.HasSufficientCredits = (err == nil)
	}

	return options, nil
}

// CheckoutEventWithCredits handles credit-based event enrollment
func (s *Service) CheckoutEventWithCredits(ctx context.Context, eventID uuid.UUID) *errLib.CommonError {
	customerID, ctxErr := contextUtils.GetUserID(ctx)
	if ctxErr != nil {
		return ctxErr
	}

	// Check enrollment options
	options, err := s.CheckEventEnrollmentOptions(ctx, eventID)
	if err != nil {
		return err
	}

	// Validate credit payment is available
	if options.CreditCost == nil {
		return errLib.New("Event does not accept credit payments", http.StatusBadRequest)
	}
	
	if !options.HasSufficientCredits {
		return errLib.New("Insufficient credits", http.StatusBadRequest)
	}

	// If customer can enroll for free (has membership), they shouldn't use credits
	if options.CanEnrollFree {
		return errLib.New("Event is free for your membership level", http.StatusBadRequest)
	}

	// CRITICAL FIX: Process credit payment FIRST (includes all validations like weekly limit)
	// This ensures we don't reserve a seat if payment will fail
	if err := s.CreditService.EnrollWithCredits(ctx, eventID, customerID); err != nil {
		log.Printf("Credit enrollment failed for customer %s, event %s: %v", customerID, eventID, err)
		return err
	}

	// Reserve seat AFTER credits are successfully deducted
	if err := s.EnrollmentService.ReserveSeatInEvent(ctx, eventID, customerID); err != nil {
		// Credits were deducted but seat reservation failed - need to refund
		log.Printf("Seat reservation failed after credit deduction, attempting refund for customer %s, event %s", customerID, eventID)
		if refundErr := s.CreditService.RefundCreditsForCancellation(ctx, eventID, customerID); refundErr != nil {
			log.Printf("CRITICAL: Failed to refund credits after seat reservation failure: %v", refundErr)
			// Queue the failed refund for manual recovery
			s.queueFailedRefund(ctx, customerID, eventID, *options.CreditCost, refundErr)
		}
		return err
	}

	// Update payment status to 'paid' since credit payment was successful
	if err := s.EnrollmentService.UpdateReservationStatusInEvent(ctx, eventID, customerID, dbEnrollment.PaymentStatusPaid); err != nil {
		log.Printf("Failed to update reservation status after credit payment: %v", err)
		// Don't return error here since the payment already went through
	}

	return nil
}

// Enhanced CheckoutEvent with membership validation
func (s *Service) CheckoutEventEnhanced(ctx context.Context, eventID uuid.UUID, discountCode *string, successURL string) (string, *errLib.CommonError) {
	customerID, ctxErr := contextUtils.GetUserID(ctx)
	if ctxErr != nil {
		return "", ctxErr
	}

	// Check enrollment options
	options, err := s.CheckEventEnrollmentOptions(ctx, eventID)
	if err != nil {
		return "", err
	}

	// If customer can enroll for free, complete enrollment without payment
	if options.CanEnrollFree {
		if err := s.EnrollmentService.ReserveSeatInEvent(ctx, eventID, customerID); err != nil {
			return "", err
		}

		// Mark as paid since it's free for this customer
		if err := s.EnrollmentService.UpdateReservationStatusInEvent(ctx, eventID, customerID, dbEnrollment.PaymentStatusPaid); err != nil {
			log.Printf("Failed to update reservation status for free enrollment: %v", err)
		}

		return "", nil // No payment URL needed
	}

	// Event requires payment - must have Stripe price ID
	if options.StripePriceID == nil {
		return "", errLib.New("Event requires membership or is not available for purchase", http.StatusBadRequest)
	}

	// Handle discount code if provided
	var stripeCouponID *string
	if discountCode != nil {
		applied, err := s.DiscountService.ApplyDiscount(ctx, *discountCode, nil)
		if err != nil {
			return "", err
		}

		// Validate that discount applies to one-time payments
		if applied.AppliesTo != "one_time" && applied.AppliesTo != "both" {
			return "", errLib.New("This discount code does not apply to one-time payments", http.StatusBadRequest)
		}

		stripeCouponID = applied.StripeCouponID
	}

	// Proceed with existing Stripe checkout logic
	if err := s.EnrollmentService.ReserveSeatInEvent(ctx, eventID, customerID); err != nil {
		return "", err
	}

	// Get existing Stripe customer ID to reuse (industry standard: one user = one Stripe customer)
	existingCustomerID := s.getExistingStripeCustomerID(ctx, customerID)

	eventIDStr := eventID.String()
	return stripe.CreateOneTimePayment(ctx, *options.StripePriceID, 1, &eventIDStr, stripeCouponID, successURL, existingCustomerID)
}
