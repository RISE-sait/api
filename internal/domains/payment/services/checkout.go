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

	if requirements.IsOneTimePayment {
		return stripe.CreateOneTimePayment(ctx, requirements.Name, 1, requirements.Price)
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

func (s *Service) CheckoutProgram(ctx context.Context, userID, programID uuid.UUID) (string, *errLib.CommonError) {

	joinProgramRequirements, err := s.PurchaseRepo.GetProgramRegistrationInfoByCustomer(ctx, userID, programID)

	if err != nil {
		return "", err
	}

	if !joinProgramRequirements.Price.Valid {
		return "", errLib.New("User is not eligible to join program", http.StatusBadRequest)
	}

	if link, err := stripe.CreateOneTimePayment(ctx, joinProgramRequirements.ProgramName, 1, joinProgramRequirements.Price.Decimal); err != nil {
		return "", err
	} else {
		return link, nil
	}
}
