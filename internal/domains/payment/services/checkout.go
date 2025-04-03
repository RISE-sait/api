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
	PurchaseRepo        *repository.Repository
	MembershipPlansRepo *membership.PlansRepository
	SquareClient        *squareClient.Client
}

func NewPurchaseService(container *di.Container) *Service {
	return &Service{
		PurchaseRepo:        repository.NewRepository(container),
		MembershipPlansRepo: membership.NewMembershipPlansRepository(container),
		SquareClient:        container.SquareClient,
	}
}

func (s *Service) CheckoutMembershipPlan(ctx context.Context, membershipPlanID uuid.UUID) (string, *errLib.CommonError) {

	requirements, err := s.PurchaseRepo.GetMembershipPlanJoiningRequirement(ctx, membershipPlanID)

	if err != nil {
		return "", err
	}

	if link, err := stripe.CreateSubscription(ctx, requirements.StripePriceID, requirements.StripeJoiningFeeID); err != nil {
		return "", err
	} else {
		return link, nil
	}
}

func (s *Service) CheckoutProgram(ctx context.Context, userID, programID uuid.UUID) (string, *errLib.CommonError) {

	priceID, err := s.PurchaseRepo.GetProgramRegistrationPriceIDByCustomer(ctx, userID, programID)

	if err != nil {
		return "", err
	}

	if link, err := stripe.CreateOneTimePayment(ctx, priceID, 1); err != nil {
		return "", err
	} else {
		return link, nil
	}
}
