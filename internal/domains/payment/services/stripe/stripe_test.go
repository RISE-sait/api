package stripe_test

import (
	payment "api/internal/domains/payment/services/stripe"
	contextUtils "api/utils/context"
	"context"
	"github.com/google/uuid"
	"log"
	"net/http"
	"strings"
	"testing"

	_ "api/internal/di"
	_ "github.com/square/square-go-sdk/client"
	_ "github.com/stripe/stripe-go/v81"
)

func TestCreateOneTimePayment(t *testing.T) {
	ctx := context.WithValue(context.Background(), contextUtils.UserIDKey, uuid.New())
	tests := []struct {
		name     string
		priceID  string
		quantity int
		wantErr  bool
		errMsg   string
		httpCode int
	}{
		{
			name:     "successful payment",
			priceID:  "price_1RAJEOAB1pU7EbknIH4e3bBu",
			quantity: 1,
			wantErr:  false,
		},
		{
			name:     "empty price ID",
			quantity: 1,
			wantErr:  true,
			errMsg:   "item stripe price ID cannot be empty",
			httpCode: http.StatusBadRequest,
		},
		{
			name:     "zero quantity",
			priceID:  "Test Product",
			quantity: 0,
			wantErr:  true,
			errMsg:   "quantity must be positive",
			httpCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			paymentLink, err := payment.CreateOneTimePayment(ctx, tt.priceID, tt.quantity)

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
	// Common test setup
	ctx := context.WithValue(context.Background(), contextUtils.UserIDKey, uuid.New())

	tests := []struct {
		name          string
		priceID       string
		joiningFeesID string
		periods       int32
		wantErr       bool
		errMsg        string
	}{
		{
			name:          "successful subscription",
			priceID:       "price_1RAJEOAB1pU7EbknIH4e3bBu",
			joiningFeesID: "price_1RA7MAAB1pU7EbknpkvwLmyp",
			periods:       12,
			wantErr:       false,
		},
		{
			name:          "single period subscription",
			priceID:       "price_1RAJEOAB1pU7EbknIH4e3bBu",
			joiningFeesID: "price_1RA7MAAB1pU7EbknpkvwLmyp",
			periods:       1,
			wantErr:       true,
			errMsg:        "periods must be at least 2 for subscriptions. Use create one time payment if its not recurring",
		},
		{
			name:          "missing price ID",
			joiningFeesID: "price_1RA7MAAB1pU7EbknpkvwLmyp",
			periods:       12,
			wantErr:       true,
			errMsg:        "item stripe price ID cannot be empty",
		},
		{
			name:    "missing joining fees ID",
			priceID: "price_1RAJEOAB1pU7EbknIH4e3bBu",
			periods: 12,
			wantErr: false,
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
			wantErr: false,
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
