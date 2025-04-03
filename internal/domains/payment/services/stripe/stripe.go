package stripe

import (
	_ "api/internal/di"
	errLib "api/internal/libs/errors"
	contextUtils "api/utils/context"
	"context"
	"github.com/google/uuid"
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
		PaymentIntentData: &stripe.CheckoutSessionPaymentIntentDataParams{
			Metadata: map[string]string{
				"userID":  userID.String(),
				"priceID": itemStripePriceID,
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
	stripePlanID string,
	stripeJoiningFeesID string,
) (string, *errLib.CommonError) {

	//userID, err := contextUtils.GetUserID(ctx)
	//
	//if err != nil {
	//	return "", err
	//}

	userID := uuid.MustParse("844047ca-e27d-42f4-b245-ba6ae28b090a")

	if strings.ReplaceAll(stripe.Key, " ", "") == "" {
		return "", errLib.New("Stripe not initialized", http.StatusInternalServerError)
	}

	params := &stripe.CheckoutSessionParams{
		SubscriptionData: &stripe.CheckoutSessionSubscriptionDataParams{
			Metadata: map[string]string{
				"userID": userID.String(), // Accessible in subscription.Metadata
				"planID": stripePlanID,
			},
		},
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				Price:    stripe.String(stripePlanID),
				Quantity: stripe.Int64(1),
			},
		},
		Mode:       stripe.String("subscription"),
		SuccessURL: stripe.String("https://example.com/success"),
	}

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
