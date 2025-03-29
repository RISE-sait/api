package payment

import (
	"api/internal/di"
	membership "api/internal/domains/membership/persistence/repositories"
	repository "api/internal/domains/payment/persistence/repositories"
	errLib "api/internal/libs/errors"
	contextUtils "api/utils/context"
	"context"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	squareClient "github.com/square/square-go-sdk/client"
	"github.com/stripe/stripe-go/v81"
	"github.com/stripe/stripe-go/v81/checkout/session"
	"log"
	"net/http"
	"strings"
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
		return createOneTimePayment(ctx, requirements.Name, 1, requirements.Price)
	}

	if !IsFrequencyValid(Frequency(requirements.PaymentFrequency)) {

		log.Printf("Invalid payment frequency when checking out membership plan. Plan ID: %v", membershipPlanID)
		return "", errLib.New("Internal Server Error when checking out membership plan ", http.StatusInternalServerError)
	}

	if link, err := createSubscription(ctx, requirements.Name, requirements.Price, Frequency(requirements.PaymentFrequency), *requirements.AmtPeriods); err != nil {
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

	if link, err := createOneTimePayment(ctx, joinProgramRequirements.ProgramName, 1, joinProgramRequirements.Price.Decimal); err != nil {
		return "", err
	} else {
		return link, nil
	}
}

func createOneTimePayment(
	ctx context.Context,
	itemName string,
	quantity int,
	price decimal.Decimal,
) (string, *errLib.CommonError) {

	if strings.ReplaceAll(stripe.Key, " ", "") == "" {
		return "", errLib.New("Stripe not initialized", http.StatusInternalServerError)
	}

	if itemName == "" {
		return "", errLib.New("item name cannot be empty", http.StatusBadRequest)
	}

	if quantity <= 0 {
		return "", errLib.New("quantity must be positive", http.StatusBadRequest)
	}

	userID, err := contextUtils.GetUserID(ctx)

	if err != nil {
		return "", err
	}

	priceInCents := price.Mul(decimal.NewFromInt(100)).IntPart()

	params := &stripe.CheckoutSessionParams{
		PaymentIntentData: &stripe.CheckoutSessionPaymentIntentDataParams{
			Metadata: map[string]string{"userID": userID.String()},
		},
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				PriceData: &stripe.CheckoutSessionLineItemPriceDataParams{
					Currency: stripe.String("cad"),
					ProductData: &stripe.CheckoutSessionLineItemPriceDataProductDataParams{
						Name: stripe.String(itemName),
					},
					UnitAmount: stripe.Int64(priceInCents),
				},
				Quantity: stripe.Int64(int64(quantity)),
			},
		},
		Mode:       stripe.String("payment"),
		SuccessURL: stripe.String("https://example.com/success"),
	}

	s, sessionErr := session.New(params)
	if sessionErr != nil {
		return "", errLib.New("Payment session failed: "+sessionErr.Error(), http.StatusInternalServerError)
	}
	return s.URL, nil
}

func createSubscription(
	ctx context.Context,
	planName string,
	price decimal.Decimal,
	frequency Frequency,
	periods int32,
) (string, *errLib.CommonError) {

	if planName == "" {
		return "", errLib.New("plan name cannot be empty", http.StatusBadRequest)
	}

	if price.LessThanOrEqual(decimal.Zero) {
		return "", errLib.New("price must be positive", http.StatusBadRequest)
	}

	if periods < 2 {
		return "", errLib.New("periods must be at least 2 for subscriptions. Use create one time payment if its not recurring", http.StatusBadRequest)
	}

	userID, err := contextUtils.GetUserID(ctx)

	if err != nil {
		return "", err
	}

	if strings.ReplaceAll(stripe.Key, " ", "") == "" {
		return "", errLib.New("Stripe not initialized", http.StatusInternalServerError)
	}

	interval := string(frequency)

	intervalCount := 1

	if frequency == Biweekly {
		interval = "week"
		intervalCount = 2
	}

	priceInCents := price.Mul(decimal.NewFromInt(100)).IntPart()

	params := &stripe.CheckoutSessionParams{
		SubscriptionData: &stripe.CheckoutSessionSubscriptionDataParams{
			Metadata: map[string]string{
				"userID":  userID.String(), // Accessible in subscription.Metadata
				"periods": string(periods),
			},
		},
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				PriceData: &stripe.CheckoutSessionLineItemPriceDataParams{
					Currency: stripe.String("cad"),
					ProductData: &stripe.CheckoutSessionLineItemPriceDataProductDataParams{
						Name: stripe.String(planName),
					},
					Recurring: &stripe.CheckoutSessionLineItemPriceDataRecurringParams{
						Interval:      stripe.String(interval),
						IntervalCount: stripe.Int64(int64(intervalCount)),
					},
					UnitAmount: stripe.Int64(priceInCents),
				},
				Quantity: stripe.Int64(1),
			},
		},
		Mode:       stripe.String("subscription"),
		SuccessURL: stripe.String("https://example.com/success"),
	}

	s, sessionErr := session.New(params)
	if sessionErr != nil {
		return "", errLib.New("Subscription setup failed: "+sessionErr.Error(), http.StatusInternalServerError)
	}

	return s.URL, nil
}
