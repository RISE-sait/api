package stripe_test

import (
	payment "api/internal/domains/payment/services/stripe"
	contextUtils "api/utils/context"
	"context"
	"github.com/google/uuid"
	"log"
	"net/http"
	"os"
	"strings"
	"testing"

	_ "api/internal/di"
	stripeAPI "github.com/stripe/stripe-go/v81"
)

func TestCreateOneTimePayment(t *testing.T) {
	// Initialize Stripe if API key is available, otherwise test error handling
	if apiKey := os.Getenv("STRIPE_SECRET_KEY"); apiKey != "" {
		stripeAPI.Key = apiKey
	}
	
	ctx := context.WithValue(context.Background(), contextUtils.UserIDKey, uuid.New())
	
	// Determine if we expect Stripe not initialized error
	expectStripeError := stripeAPI.Key == ""
	
	tests := []struct {
		name     string
		priceID  string
		quantity int
		wantErr  bool
		errMsg   string
		httpCode int
	}{
		{
			name:     "successful payment or stripe not initialized",
			priceID:  "price_1R9srEAB1pU7EbknzAO7IVi8",
			quantity: 1,
			wantErr:  expectStripeError,
			errMsg:   func() string { if expectStripeError { return "Stripe not initialized" } else { return "" } }(),
		},
		{
			name:     "empty price ID",
			quantity: 1,
			wantErr:  true,
			errMsg:   func() string { if expectStripeError { return "Stripe not initialized" } else { return "item stripe price ID cannot be empty" } }(),
			httpCode: http.StatusBadRequest,
		},
		{
			name:     "zero quantity",
			priceID:  "Test Product",
			quantity: 0,
			wantErr:  true,
			errMsg:   func() string { if expectStripeError { return "Stripe not initialized" } else { return "quantity must be positive" } }(),
			httpCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			paymentLink, err := payment.CreateOneTimePayment(ctx, tt.priceID, tt.quantity, nil)

			if tt.wantErr {
				if err == nil {
					t.Error("Expected error but got none")
				} else if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("Expected error to contain %q, got %q", tt.errMsg, err.Error())
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if paymentLink == "" {
				t.Error("Expected a payment link, got empty string")
			}

			if !strings.HasPrefix(paymentLink, "https://") {
				t.Errorf("Expected URL to start with https://, got %q", paymentLink)
			}

			log.Printf("Successfully generated payment link: %s", paymentLink)
		})
	}
}

func TestCreateSubscription(t *testing.T) {
	// Initialize Stripe if API key is available, otherwise test error handling
	if apiKey := os.Getenv("STRIPE_SECRET_KEY"); apiKey != "" {
		stripeAPI.Key = apiKey
	}
	
	// Common test setup
	ctx := context.WithValue(context.Background(), contextUtils.UserIDKey, uuid.New())
	
	// Determine if we expect Stripe not initialized error
	expectStripeError := stripeAPI.Key == ""

	tests := []struct {
		name          string
		priceID       string
		joiningFeesID string
		wantErr       bool
		errMsg        string
	}{
		{
			name:          "successful subscription or stripe not initialized",
			priceID:       "price_1RAJEOAB1pU7EbknIH4e3bBu",
			joiningFeesID: "price_1RA7MAAB1pU7EbknpkvwLmyp",
			wantErr:       expectStripeError,
			errMsg:        func() string { if expectStripeError { return "Stripe not initialized" } else { return "" } }(),
		},
		{
			name:          "missing price ID",
			joiningFeesID: "price_1RA7MAAB1pU7EbknpkvwLmyp",
			wantErr:       true,
			errMsg:        func() string { if expectStripeError { return "Stripe not initialized" } else { return "item stripe price ID cannot be empty" } }(),
		},
		{
			name:    "missing joining fees ID or stripe not initialized",
			priceID: "price_1RAJEOAB1pU7EbknIH4e3bBu",
			wantErr: expectStripeError,
			errMsg:  func() string { if expectStripeError { return "Stripe not initialized" } else { return "" } }(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			subscriptionLink, err := payment.CreateSubscription(
				ctx,
				tt.priceID,
				tt.joiningFeesID,
			)

			if tt.wantErr {
				if err == nil {
					t.Error("Expected error but got none")
				} else if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("Expected error to contain %q, got %q", tt.errMsg, err.Error())
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if subscriptionLink == "" {
				t.Error("Expected a subscription link, got empty string")
			}

			// Verify the URL is a valid stripeService URL
			if !strings.HasPrefix(subscriptionLink, "https://") {
				t.Errorf("Expected URL to start with https://, got %q", subscriptionLink)
			}

			t.Logf("Successfully generated subscription link: %s", subscriptionLink)
		})
	}
}

func TestCreateSubscription_Context(t *testing.T) {
	// Initialize Stripe if API key is available, otherwise test error handling
	if apiKey := os.Getenv("STRIPE_SECRET_KEY"); apiKey != "" {
		stripeAPI.Key = apiKey
	}
	
	// Determine if we expect Stripe not initialized error
	expectStripeError := stripeAPI.Key == ""
	
	tests := []struct {
		name     string
		ctx      context.Context
		wantErr  bool
		errMsg   string // Expected error message substring
		httpCode int    // Expected HTTP status code
	}{
		{
			name:     "nil context",
			ctx:      nil,
			wantErr:  true,
			errMsg:   "context cannot be nil",
			httpCode: http.StatusBadRequest,
		},
		{
			name:     "missing user ID",
			ctx:      context.Background(),
			wantErr:  true,
			errMsg:   "user ID not found in context",
			httpCode: http.StatusUnauthorized,
		},
		{
			name:    "valid context",
			ctx:     context.WithValue(context.Background(), contextUtils.UserIDKey, uuid.New()),
			wantErr: expectStripeError,
			errMsg:  func() string { if expectStripeError { return "Stripe not initialized" } else { return "" } }(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			subscriptionLink, err := payment.CreateSubscription(
				tt.ctx,
				"price_1RAJEOAB1pU7EbknIH4e3bBu",
				"price_1RA7MAAB1pU7EbknpkvwLmyp",
			)

			if tt.wantErr {
				if err == nil {
					t.Error("Expected error but got none")
					return
				}

				// Check error message
				if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("Expected error to contain %q, got %q", tt.errMsg, err.Error())
					return
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
					return
				}

				log.Printf("Successfully created subscription with url: %v", subscriptionLink)
			}
		})
	}
}
