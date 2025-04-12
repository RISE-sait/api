package payment

import (
	"api/internal/di"
	enrollment "api/internal/domains/enrollment/service"
	membership "api/internal/domains/membership/persistence/repositories"
	repository "api/internal/domains/payment/persistence/repositories"
	"api/internal/domains/payment/services/stripe"
	errLib "api/internal/libs/errors"
	contextUtils "api/utils/context"
	"context"
	"github.com/google/uuid"
	squareClient "github.com/square/square-go-sdk/client"
	"log"
	"net/http"
)

type Service struct {
	CheckoutRepo        *repository.CheckoutRepository
	MembershipPlansRepo *membership.PlansRepository
	SquareClient        *squareClient.Client
	EnrollmentService   *enrollment.CustomerEnrollmentService
}

func NewPurchaseService(container *di.Container) *Service {
	return &Service{
		CheckoutRepo:        repository.NewCheckoutRepository(container),
		MembershipPlansRepo: membership.NewMembershipPlansRepository(container),
		EnrollmentService:   enrollment.NewCustomerEnrollmentService(container),
		SquareClient:        container.SquareClient,
	}
}

func (s *Service) CheckoutMembershipPlan(ctx context.Context, membershipPlanID uuid.UUID) (string, *errLib.CommonError) {

	requirements, err := s.CheckoutRepo.GetMembershipPlanJoiningRequirement(ctx, membershipPlanID)

	if err != nil {
		return "", err
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

	err = s.EnrollmentService.ReserveSeatInProgram(ctx, programID, customerID)

	if err != nil {
		log.Println("Failed to reserve seat in program:", err)
		return "", err
	}

	return stripe.CreateOneTimePayment(ctx, priceID, 1)
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

	if err = s.EnrollmentService.ReserveSeatInEvent(ctx, eventID, customerID); err != nil {
		return "", err
	}

	return stripe.CreateOneTimePayment(ctx, priceID, 1)
}
