package payment

import (
	"api/internal/di"
	membership "api/internal/domains/membership/persistence/repositories"
	repository "api/internal/domains/payment/persistence/repositories"
	"api/internal/domains/payment/services/stripe"
	errLib "api/internal/libs/errors"
	"context"
	"github.com/google/uuid"
	squareClient "github.com/square/square-go-sdk/client"
)

type Service struct {
	CheckoutRepo        *repository.CheckoutRepository
	MembershipPlansRepo *membership.PlansRepository
	SquareClient        *squareClient.Client
}

func NewPurchaseService(container *di.Container) *Service {
	return &Service{
		CheckoutRepo:        repository.NewCheckoutRepository(container),
		MembershipPlansRepo: membership.NewMembershipPlansRepository(container),
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

	priceID, err := s.CheckoutRepo.GetProgramRegistrationPriceIdForCustomer(ctx, programID)

	if err != nil {
		return "", err
	}

	return stripe.CreateOneTimePayment(ctx, priceID, 1)
}
