package stripe

import (
	"context"
	"net/http"
	"strings"

	_ "api/internal/di"
	errLib "api/internal/libs/errors"
	contextUtils "api/utils/context"

	_ "github.com/square/square-go-sdk/client"
	"github.com/stripe/stripe-go/v81"
	"github.com/stripe/stripe-go/v81/checkout/session"
)

// CreateOneTimePayment creates a Stripe Checkout Session for a one-time payment
func CreateOneTimePayment(
	ctx context.Context, // request-scoped context (for userID, cancellation, etc.)
	itemStripePriceID string, // Stripe Price ID of the item being purchased
	quantity int, // Number of items
) (string, *errLib.CommonError) {
	// Check if the Stripe API key is set
	if strings.ReplaceAll(stripe.Key, " ", "") == "" {
		return "", errLib.New("Stripe not initialized", http.StatusInternalServerError)
	}

	// Validate input
	if itemStripePriceID == "" {
		return "", errLib.New("item stripe price ID cannot be empty", http.StatusBadRequest)
	}

	if quantity <= 0 {
		return "", errLib.New("quantity must be positive", http.StatusBadRequest)
	}

	// Extract user ID from context (e.g. JWT or middleware-injected value)
	userID, err := contextUtils.GetUserID(ctx)
	if err != nil {
		return "", err
	}

	// Metadata includes userID to track who made the payment
	params := &stripe.CheckoutSessionParams{
		Metadata: map[string]string{
			"userID": userID.String(),
		},
		PaymentIntentData: &stripe.CheckoutSessionPaymentIntentDataParams{
			Metadata: map[string]string{
				"userID": userID.String(),
			},
		},
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				Price:    stripe.String(itemStripePriceID), // Price ID (pre-created in Stripe dashboard)
				Quantity: stripe.Int64(int64(quantity)),    // Number of items
			},
		},
		Mode:       stripe.String("payment"),                     // One-time payment mode
		SuccessURL: stripe.String("https://example.com/success"), // Redirect URL after success
	}

	// Create Stripe session
	s, sessionErr := session.New(params)
	if sessionErr != nil {
		return "", errLib.New("Payment session failed: "+sessionErr.Error(), http.StatusInternalServerError)
	}

	// Return session URL to redirect client to
	return s.URL, nil
}

// CreateSubscription creates a Stripe Checkout Session for a recurring subscription
func CreateSubscription(
	ctx context.Context,
	stripePlanPriceID string, // Stripe Price ID for the recurring plan
	stripeJoiningFeesID string, // Optional one-time joining fee
) (string, *errLib.CommonError) {
	// Extract user ID from context
	userID, err := contextUtils.GetUserID(ctx)
	if err != nil {
		return "", err
	}

	// Check if Stripe is initialized
	if strings.ReplaceAll(stripe.Key, " ", "") == "" {
		return "", errLib.New("Stripe not initialized", http.StatusInternalServerError)
	}

	// Validate input
	if stripePlanPriceID == "" {
		return "", errLib.New("item stripe price ID cannot be empty", http.StatusBadRequest)
	}

	// Set up Checkout session with subscription mode
	params := &stripe.CheckoutSessionParams{
		Metadata: map[string]string{
			"userID": userID.String(),
		},
		SubscriptionData: &stripe.CheckoutSessionSubscriptionDataParams{
			Metadata: map[string]string{
				"userID": userID.String(),
			},
		},
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				Price:    stripe.String(stripePlanPriceID), // Main subscription plan
				Quantity: stripe.Int64(1),
			},
		},
		Mode:       stripe.String("subscription"), // Subscription mode
		SuccessURL: stripe.String("https://example.com/success"),
	}

	// Ask Stripe to expand line item pricing and subscription in response
	params.AddExpand("line_items.data.price")
	params.AddExpand("subscription")

	// If there's a joining fee, add it as a second line item
	if stripeJoiningFeesID != "" {
		params.LineItems = append(params.LineItems, &stripe.CheckoutSessionLineItemParams{
			Price:    stripe.String(stripeJoiningFeesID),
			Quantity: stripe.Int64(1),
		})
	}

	// Create Stripe session
	s, sessionErr := session.New(params)
	if sessionErr != nil {
		return "", errLib.New("Subscription setup failed: "+sessionErr.Error(), http.StatusInternalServerError)
	}

	return s.URL, nil // Return URL to redirect client for payment
}
