package payment

import (
	"api/internal/di"
	membership "api/internal/domains/membership/persistence/repositories"
	repository "api/internal/domains/payment/persistence/repositories"
	"api/internal/domains/payment/services/stripe"
	types "api/internal/domains/payment/types"
	errLib "api/internal/libs/errors"
	"context"
	"github.com/google/uuid"
	squareClient "github.com/square/square-go-sdk/client"
	"log"
	"net/http"
)

type CheckoutService struct {
	PurchaseRepo        *repository.CheckoutRepository
	MembershipPlansRepo *membership.PlansRepository
	SquareClient        *squareClient.Client
}

func NewCheckoutService(container *di.Container) *CheckoutService {
	return &CheckoutService{
		PurchaseRepo:        repository.NewCheckoutRepository(container),
		MembershipPlansRepo: membership.NewMembershipPlansRepository(container),
		SquareClient:        container.SquareClient,
	}
}

func (s *CheckoutService) CheckoutMembershipPlan(ctx context.Context, membershipPlanID uuid.UUID) (string, *errLib.CommonError) {

	requirements, err := s.PurchaseRepo.GetMembershipPlanJoiningRequirement(ctx, membershipPlanID)

	if err != nil {
		return "", err
	}

	lineItems := []types.CheckoutItem{
		{
			ID:       membershipPlanID,
			Name:     requirements.Name,
			Quantity: 1,
			Price:    requirements.Price,
		},
	}

	if requirements.IsOneTimePayment {
		return stripe.CreateOneTimePayment(ctx, lineItems, types.MembershipPlan)
	}

	if !types.IsPaymentFrequencyValid(types.PaymentFrequency(requirements.PaymentFrequency)) {

		log.Printf("Invalid payment frequency when checking out membership plan. Plan ID: %v", membershipPlanID)
		return "", errLib.New("Internal Server Error when checking out membership plan ", http.StatusInternalServerError)
	}

	if link, err := stripe.CreateSubscription(ctx, requirements.Name, requirements.Price, types.PaymentFrequency(requirements.PaymentFrequency), *requirements.AmtPeriods); err != nil {
		return "", err
	} else {
		return link, nil
	}
}

func (s *CheckoutService) CheckoutProgram(ctx context.Context, programID uuid.UUID) (string, *errLib.CommonError) {

	joinProgramRequirements, err := s.PurchaseRepo.GetProgramRegistrationInfoForCustomer(ctx, programID)

	if err != nil {
		return "", err
	}

	if !joinProgramRequirements.Price.Valid {
		return "", errLib.New("User is not eligible to join program", http.StatusBadRequest)
	}

	items := []types.CheckoutItem{
		{
			ID:       programID,
			Name:     joinProgramRequirements.ProgramName,
			Quantity: 1,
			Price:    joinProgramRequirements.Price.Decimal,
		},
	}

	if link, err := stripe.CreateOneTimePayment(ctx, items, types.Program); err != nil {
		return "", err
	} else {
		return link, nil
	}
}
