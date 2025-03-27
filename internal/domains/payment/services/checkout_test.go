package payment_test

import (
	payment "api/internal/domains/payment/services"
	"api/internal/middlewares"
	"context"
	"github.com/shopspring/decimal"
	"log"
	"net/http"
	"strings"
	"testing"
)

func TestCreateOneTimePayment(t *testing.T) {
	ctx := context.WithValue(context.Background(), middlewares.UserIDKey, "test_user_123")
	tests := []struct {
		name     string
		itemName string
		quantity int
		price    decimal.Decimal
		wantErr  bool
		errMsg   string
		httpCode int
	}{
		{
			name:     "successful payment",
			itemName: "Test Product",
			quantity: 1,
			price:    decimal.NewFromFloat(19.99),
			wantErr:  false,
		},
		{
			name:     "empty item name",
			itemName: "",
			quantity: 1,
			price:    decimal.NewFromFloat(19.99),
			wantErr:  true,
			errMsg:   "item name cannot be empty",
			httpCode: http.StatusBadRequest,
		},
		{
			name:     "zero quantity",
			itemName: "Test Product",
			quantity: 0,
			price:    decimal.NewFromFloat(19.99),
			wantErr:  true,
			errMsg:   "quantity must be positive",
			httpCode: http.StatusBadRequest,
		},
		{
			name:     "zero price",
			itemName: "Test Product",
			quantity: 1,
			price:    decimal.Zero,
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			paymentLink, err := payment.CreateOneTimePayment(ctx, tt.itemName, tt.quantity, tt.price)

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
	ctx := context.WithValue(context.Background(), middlewares.UserIDKey, "test_user_123")
	basePlan := "Premium Membership"
	basePrice := decimal.NewFromFloat(9.99)

	tests := []struct {
		name     string
		planName string
		price    decimal.Decimal
		interval payment.Frequency
		periods  int32
		wantErr  bool
		errMsg   string
	}{
		{
			name:     "successful monthly subscription",
			planName: basePlan,
			price:    basePrice,
			interval: payment.Month,
			periods:  12,
			wantErr:  false,
		},
		{
			name:     "successful biweekly subscription",
			planName: basePlan,
			price:    basePrice,
			interval: payment.Biweekly,
			periods:  12,
			wantErr:  false,
		},
		{
			name:     "empty plan name",
			planName: "",
			price:    basePrice,
			interval: payment.Month,
			periods:  12,
			wantErr:  true,
			errMsg:   "plan name cannot be empty",
		},
		{
			name:     "single period subscription",
			planName: basePlan,
			price:    basePrice,
			interval: payment.Month,
			periods:  1,
			wantErr:  true,
			errMsg:   "totalBillingPeriods must be at least 2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			subscriptionLink, err := payment.CreateSubscription(
				ctx,
				tt.planName,
				tt.price,
				tt.interval,
				tt.periods,
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
			ctx:     context.WithValue(context.Background(), middlewares.UserIDKey, "test_user_123"),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			subscriptionLink, err := payment.CreateSubscription(
				tt.ctx,
				"Test Plan",
				decimal.NewFromFloat(9.99),
				payment.Month,
				12,
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
