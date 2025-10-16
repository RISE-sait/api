package payment

import (
	"context"
	"log"
	"net/http"

	"api/internal/di"
	enrollment "api/internal/domains/enrollment/service"
	dbEnrollment "api/internal/domains/enrollment/persistence/sqlc/generated"
	eventService "api/internal/domains/event/service"
	membership "api/internal/domains/membership/persistence/repositories"
	repository "api/internal/domains/payment/persistence/repositories"
	"api/internal/domains/payment/services/stripe"
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
	EnrollmentService   *enrollment.CustomerEnrollmentService
	EventService        *eventService.Service
	CreditService       *userServices.CustomerCreditService
}

func NewPurchaseService(container *di.Container) *Service {
	return &Service{
		CheckoutRepo:        repository.NewCheckoutRepository(container),
		MembershipPlansRepo: membership.NewMembershipPlansRepository(container),
		DiscountService:     discountService.NewService(container),
		EnrollmentService:   enrollment.NewCustomerEnrollmentService(container),
		EventService:        eventService.NewEventService(container),
		CreditService:       userServices.NewCustomerCreditService(container),
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
		if discountCode != nil {
		applied, err := s.DiscountService.ApplyDiscount(ctx, *discountCode, &membershipPlanID)
		if err != nil {
			return "", err
		}
		return stripe.CreateSubscriptionWithDiscountPercent(ctx, requirements.StripePriceID, requirements.StripeJoiningFeeID, applied.DiscountPercent, successURL)
	}

	// Check if membership has any joining fee
	if requirements.StripeJoiningFeeID != "" {
		// Use recurring joining fee (annual/monthly) - existing function handles this
		return stripe.CreateSubscription(ctx, requirements.StripePriceID, requirements.StripeJoiningFeeID, successURL)
	} else if requirements.JoiningFee > 0 {
		// Use one-time setup fee
		return stripe.CreateSubscriptionWithSetupFee(ctx, requirements.StripePriceID, requirements.JoiningFee, successURL)
	} else {
		// No joining fee - just regular subscription
		return stripe.CreateSubscription(ctx, requirements.StripePriceID, "", successURL)
	}
}

func (s *Service) CheckoutProgram(ctx context.Context, programID uuid.UUID, successURL string) (string, *errLib.CommonError) {
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

	programIDStr := programID.String()
	return stripe.CreateOneTimePayment(ctx, priceID, 1, &programIDStr, successURL)
}

func (s *Service) CheckoutEvent(ctx context.Context, eventID uuid.UUID, successURL string) (string, *errLib.CommonError) {
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

	// reserve seat so that the database can assume that the customer is enrolled
	// this is important for the enrollment process, as multiple users may try to enroll one after another,
	// and each successfully getting the stripe checkout link
	// and cause overbooking cuz the database is not aware of the amount of people that are able to pay

	// and the reservation is only valid for 10 minutes
	// so if the customer does not pay in 10 minutes, the reservation will be deemed cancelled
	if err = s.EnrollmentService.ReserveSeatInEvent(ctx, eventID, customerID); err != nil {
		return "", err
	}

	eventIDStr := eventID.String()
	return stripe.CreateOneTimePayment(ctx, priceID, 1, &eventIDStr, successURL)
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
	
	// Check if event has required membership
	if event.RequiredMembershipPlanID != nil {
		membershipPlanID := event.RequiredMembershipPlanID.String()
		options.MembershipPlanID = &membershipPlanID
		
		// Check if customer has this membership
		hasActiveMembership, err := s.CheckoutRepo.CheckCustomerHasActiveMembership(ctx, customerID, *event.RequiredMembershipPlanID)
		if err != nil {
			return nil, err
		}
		
		if hasActiveMembership {
			// Customer has required membership, can enroll for free
			options.CanEnrollFree = true
			return options, nil
		}
		// Customer doesn't have required membership, need to check payment options
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
func (s *Service) CheckoutEventEnhanced(ctx context.Context, eventID uuid.UUID, successURL string) (string, *errLib.CommonError) {
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

	// Proceed with existing Stripe checkout logic
	if err := s.EnrollmentService.ReserveSeatInEvent(ctx, eventID, customerID); err != nil {
		return "", err
	}

	eventIDStr := eventID.String()
	return stripe.CreateOneTimePayment(ctx, *options.StripePriceID, 1, &eventIDStr, successURL)
}
