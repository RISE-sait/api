package stripe

import (
	_ "api/internal/di"
	errLib "api/internal/libs/errors"
	contextUtils "api/utils/context"
	"context"
	_ "github.com/square/square-go-sdk/client"
	"github.com/stripe/stripe-go/v81"
	"github.com/stripe/stripe-go/v81/checkout/session"
	"net/http"
	"strings"
)

func CreateOneTimePayment(
	ctx context.Context,
	itemStripePriceID string,
	quantity int,
) (string, *errLib.CommonError) {

	if strings.ReplaceAll(stripe.Key, " ", "") == "" {
		return "", errLib.New("Stripe not initialized", http.StatusInternalServerError)
	}

	if itemStripePriceID == "" {
		return "", errLib.New("item stripe price ID cannot be empty", http.StatusBadRequest)
	}

	if quantity <= 0 {
		return "", errLib.New("quantity must be positive", http.StatusBadRequest)
	}

	userID, err := contextUtils.GetUserID(ctx)

	if err != nil {
		return "", err
	}

	params := &stripe.CheckoutSessionParams{
		Metadata: map[string]string{
			"userID": userID.String(), // Accessible in subscription.Metadata
		},
		PaymentIntentData: &stripe.CheckoutSessionPaymentIntentDataParams{
			Metadata: map[string]string{
				"userID": userID.String(),
			},
		},
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				Price:    stripe.String(itemStripePriceID), // Use pre-created Price ID
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

func CreateSubscription(
	ctx context.Context,
	stripePlanPriceID string,
	stripeJoiningFeesID string,
) (string, *errLib.CommonError) {

	userID, err := contextUtils.GetUserID(ctx)

	if err != nil {
		return "", err
	}

	if strings.ReplaceAll(stripe.Key, " ", "") == "" {
		return "", errLib.New("Stripe not initialized", http.StatusInternalServerError)
	}

	if stripePlanPriceID == "" {
		return "", errLib.New("item stripe price ID cannot be empty", http.StatusBadRequest)
	}

	params := &stripe.CheckoutSessionParams{
		Metadata: map[string]string{
			"userID": userID.String(), // Accessible in subscription.Metadata
		},
		SubscriptionData: &stripe.CheckoutSessionSubscriptionDataParams{
			Metadata: map[string]string{
				"userID": userID.String(), // Accessible in subscription.Metadata
			},
		},
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				Price:    stripe.String(stripePlanPriceID),
				Quantity: stripe.Int64(1),
			},
		},
		Mode:       stripe.String("subscription"),
		SuccessURL: stripe.String("https://example.com/success"),
	}

	params.AddExpand("line_items.data.price")
	params.AddExpand("subscription")

	if stripeJoiningFeesID != "" {
		params.LineItems = append(params.LineItems, &stripe.CheckoutSessionLineItemParams{
			Price:    stripe.String(stripeJoiningFeesID),
			Quantity: stripe.Int64(1),
		})
	}

	s, sessionErr := session.New(params)
	if sessionErr != nil {
		return "", errLib.New("Subscription setup failed: "+sessionErr.Error(), http.StatusInternalServerError)
	}

	return s.URL, nil
}
