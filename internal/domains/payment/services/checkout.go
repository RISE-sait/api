package payment

import (
	"context"
	"log"
	"net/http"

	"api/internal/di"
	enrollment "api/internal/domains/enrollment/service"
	membership "api/internal/domains/membership/persistence/repositories"
	repository "api/internal/domains/payment/persistence/repositories"
	"api/internal/domains/payment/services/stripe"
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
}

func NewPurchaseService(container *di.Container) *Service {
	return &Service{
		CheckoutRepo:        repository.NewCheckoutRepository(container),
		MembershipPlansRepo: membership.NewMembershipPlansRepository(container),
		DiscountService:     discountService.NewService(container),
		EnrollmentService:   enrollment.NewCustomerEnrollmentService(container),
	}
}

func (s *Service) CheckoutMembershipPlan(ctx context.Context, membershipPlanID uuid.UUID, discountCode *string) (string, *errLib.CommonError) {
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
		return stripe.CreateSubscriptionWithDiscountPercent(ctx, requirements.StripePriceID, requirements.StripeJoiningFeeID, applied.DiscountPercent)
	}

	return stripe.CreateSubscription(ctx, requirements.StripePriceID, requirements.StripeJoiningFeeID)
}

func (s *Service) CheckoutProgram(ctx context.Context, programID uuid.UUID) (string, *errLib.CommonError) {
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
	return stripe.CreateOneTimePayment(ctx, priceID, 1, &programIDStr)
}

func (s *Service) CheckoutEvent(ctx context.Context, eventID uuid.UUID) (string, *errLib.CommonError) {
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
	return stripe.CreateOneTimePayment(ctx, priceID, 1, &eventIDStr)
}
