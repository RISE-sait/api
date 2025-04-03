package stripe

import (
	types "api/internal/domains/payment/types"
	errLib "api/internal/libs/errors"
	contextUtils "api/utils/context"
	"context"
	"github.com/shopspring/decimal"

	_ "api/internal/di"
	_ "github.com/square/square-go-sdk/client"
	"github.com/stripe/stripe-go/v81"
	"github.com/stripe/stripe-go/v81/checkout/session"
	"net/http"
	"strings"
)

func CreateOneTimePayment(
	ctx context.Context,
	items []types.CheckoutItem,
	itemType types.OneTimePaymentCheckoutItemType,
) (string, *errLib.CommonError) {

	if strings.ReplaceAll(stripe.Key, " ", "") == "" {
		return "", errLib.New("Stripe not initialized", http.StatusInternalServerError)
	}

	if len(items) == 0 {
		return "", errLib.New("items cannot be empty", http.StatusBadRequest)
	}

	userID, err := contextUtils.GetUserID(ctx)

	if err != nil {
		return "", err
	}

	lineItems := make([]*stripe.CheckoutSessionLineItemParams, len(items))
	for i, item := range items {

		if item.Quantity <= 0 {
			return "", errLib.New("item quantity must be greater than 0", http.StatusBadRequest)
		}

		if item.Price.LessThanOrEqual(decimal.Zero) {
			return "", errLib.New("item price must be greater than 0", http.StatusBadRequest)
		}

		priceInCents := item.Price.Mul(decimal.NewFromInt(100)).IntPart() // Convert dollars to cents

		lineItems[i] = &stripe.CheckoutSessionLineItemParams{
			PriceData: &stripe.CheckoutSessionLineItemPriceDataParams{
				Currency: stripe.String("cad"),
				ProductData: &stripe.CheckoutSessionLineItemPriceDataProductDataParams{
					Name: stripe.String(item.Name),
					Metadata: map[string]string{
						"item_id": item.ID.String(),
					},
				},
				UnitAmount: stripe.Int64(priceInCents),
			},
			Quantity: stripe.Int64(int64(item.Quantity)),
		}
	}

	params := &stripe.CheckoutSessionParams{
		Metadata: map[string]string{
			"userID":   userID.String(),
			"itemType": string(itemType),
		},
		PaymentIntentData: &stripe.CheckoutSessionPaymentIntentDataParams{
			Metadata: map[string]string{
				"userID":   userID.String(),
				"itemType": string(itemType),
			},
		},
		LineItems:  lineItems,
		Mode:       stripe.String("payment"),
		SuccessURL: stripe.String("https://example.com/success"),
	}

	s, sessionErr := session.New(params)
	if sessionErr != nil {
		return "", errLib.New("Payment session failed: "+sessionErr.Error(), http.StatusInternalServerError)
	}
	return s.URL, nil
}

func CreateSubscription(
	ctx context.Context,
	planName string,
	price decimal.Decimal,
	frequency types.PaymentFrequency,
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

	if frequency == types.Biweekly {
		interval = "week"
		intervalCount = 2
	}

	priceInCents := price.Mul(decimal.NewFromInt(100)).IntPart()

	params := &stripe.CheckoutSessionParams{
		Metadata: map[string]string{"userID": userID.String()},
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
