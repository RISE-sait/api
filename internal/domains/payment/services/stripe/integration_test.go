package stripe_test

import (
	"context"
	"os"
	"testing"
	
	"api/internal/di"
	"api/internal/domains/payment/services/stripe"
	contextUtils "api/utils/context"
	errLib "api/internal/libs/errors"

	"github.com/google/uuid"
	stripeAPI "github.com/stripe/stripe-go/v81"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestStripeIntegration tests the complete Stripe integration
func TestStripeIntegration(t *testing.T) {
	// Skip integration tests if Stripe API key is not set
	if os.Getenv("STRIPE_SECRET_KEY") == "" {
		t.Skip("STRIPE_SECRET_KEY not set, skipping integration tests")
	}

	// Initialize Stripe with test API key
	stripeAPI.Key = os.Getenv("STRIPE_SECRET_KEY")

	// Create test context with user ID
	userID := uuid.New()
	ctx := context.WithValue(context.Background(), contextUtils.UserIDKey, userID)

	t.Run("CreateOneTimePayment", func(t *testing.T) {
		// Skip actual Stripe API calls in CI - would need real price IDs
		t.Skip("Skipping Stripe API call - requires valid price IDs in test account")
		
		// Test creating a one-time payment checkout session
		checkoutURL, err := stripe.CreateOneTimePayment(ctx, "price_test_example", 1, nil, nil, "https://www.rise-basketball.com/success")

		// Should succeed with valid inputs
		assert.NoError(t, err)
		assert.NotEmpty(t, checkoutURL)
		assert.Contains(t, checkoutURL, "checkout.stripe.com")
	})

	t.Run("CreateOneTimePayment_WithEventID", func(t *testing.T) {
		// Skip actual Stripe API calls in CI - would need real price IDs
		t.Skip("Skipping Stripe API call - requires valid price IDs in test account")
		
		// Test creating a one-time payment with event ID for event enrollment
		eventID := uuid.New().String()
		checkoutURL, err := stripe.CreateOneTimePayment(ctx, "price_test_example", 1, &eventID, nil, "https://www.rise-basketball.com/success")

		// Should succeed with valid inputs including event ID
		assert.NoError(t, err)
		assert.NotEmpty(t, checkoutURL)
		assert.Contains(t, checkoutURL, "checkout.stripe.com")
		// Note: We can't easily verify the metadata in the checkout URL, 
		// but the webhook processing will test that the eventID is properly passed
	})

	t.Run("CreateOneTimePayment_InvalidInputs", func(t *testing.T) {
		// Test with empty price ID
		_, err := stripe.CreateOneTimePayment(ctx, "", 1, nil, nil, "https://www.rise-basketball.com/success")
		assert.Error(t, err)
		assert.Equal(t, 400, err.HTTPCode)

		// Test with invalid quantity
		_, err = stripe.CreateOneTimePayment(ctx, "price_test_example", 0, nil, nil, "https://www.rise-basketball.com/success")
		assert.Error(t, err)
		assert.Equal(t, 400, err.HTTPCode)

		// Test with empty success URL
		_, err = stripe.CreateOneTimePayment(ctx, "price_test_example", 1, nil, nil, "")
		assert.Error(t, err)
		assert.Equal(t, 400, err.HTTPCode)
	})

	t.Run("CreateSubscription", func(t *testing.T) {
		// Skip actual Stripe API calls in CI - would need real price IDs
		t.Skip("Skipping Stripe API call - requires valid price IDs in test account")

		// Test creating a subscription checkout session
		checkoutURL, err := stripe.CreateSubscription(ctx, "price_test_subscription", "", nil, "https://www.rise-basketball.com/success")

		// Should succeed with valid inputs
		assert.NoError(t, err)
		assert.NotEmpty(t, checkoutURL)
		assert.Contains(t, checkoutURL, "checkout.stripe.com")
	})

	t.Run("CreateSubscription_WithJoiningFee", func(t *testing.T) {
		// Skip actual Stripe API calls in CI - would need real price IDs
		t.Skip("Skipping Stripe API call - requires valid price IDs in test account")

		// Test creating subscription with joining fee
		checkoutURL, err := stripe.CreateSubscription(ctx, "price_test_subscription", "price_test_joining_fee", nil, "https://www.rise-basketball.com/success")

		// Should succeed with valid inputs
		assert.NoError(t, err)
		assert.NotEmpty(t, checkoutURL)
		assert.Contains(t, checkoutURL, "checkout.stripe.com")
	})

	t.Run("CreateSubscriptionWithDiscount", func(t *testing.T) {
		// Skip actual Stripe API calls in CI - would need real price IDs
		t.Skip("Skipping Stripe API call - requires valid price IDs in test account")
		
		// Test creating subscription with discount
		checkoutURL, err := stripe.CreateSubscriptionWithDiscountPercent(ctx, "price_test_subscription", "", 20, "https://www.rise-basketball.com/success")

		// Should succeed with valid inputs
		assert.NoError(t, err)
		assert.NotEmpty(t, checkoutURL)
		assert.Contains(t, checkoutURL, "checkout.stripe.com")
	})

	t.Run("CreateSubscriptionWithDiscount_InvalidPercent", func(t *testing.T) {
		// Test with invalid discount percentage
		_, err := stripe.CreateSubscriptionWithDiscountPercent(ctx, "price_test_subscription", "", 0, "https://www.rise-basketball.com/success")
		assert.Error(t, err)
		assert.Equal(t, 400, err.HTTPCode)

		_, err = stripe.CreateSubscriptionWithDiscountPercent(ctx, "price_test_subscription", "", 101, "https://www.rise-basketball.com/success")
		assert.Error(t, err)
		assert.Equal(t, 400, err.HTTPCode)
	})
}

// TestSubscriptionService tests the subscription management service
func TestSubscriptionService(t *testing.T) {
	if os.Getenv("STRIPE_SECRET_KEY") == "" {
		t.Skip("STRIPE_SECRET_KEY not set, skipping integration tests")
	}

	stripeAPI.Key = os.Getenv("STRIPE_SECRET_KEY")
	// Create mock container for testing - tests don't need real DB
	container := &di.Container{DB: nil}
	service := stripe.NewSubscriptionService(container)

	userID := uuid.New()
	ctx := context.WithValue(context.Background(), contextUtils.UserIDKey, userID)

	t.Run("GetSubscription_NotFound", func(t *testing.T) {
		// Test getting non-existent subscription
		_, err := service.GetSubscription(ctx, "sub_nonexistent")
		assert.Error(t, err)
		// Should return 500 for Stripe API error, not 404
		assert.Equal(t, 500, err.HTTPCode)
	})

	t.Run("GetSubscription_EmptyID", func(t *testing.T) {
		// Test with empty subscription ID
		_, err := service.GetSubscription(ctx, "")
		assert.Error(t, err)
		assert.Equal(t, 400, err.HTTPCode)
	})

	t.Run("CancelSubscription_EmptyID", func(t *testing.T) {
		// Test canceling with empty subscription ID
		_, err := service.CancelSubscription(ctx, "", false)
		assert.Error(t, err)
		assert.Equal(t, 400, err.HTTPCode)
	})

	t.Run("PauseSubscription_EmptyID", func(t *testing.T) {
		// Test pausing with empty subscription ID
		_, err := service.PauseSubscription(ctx, "", nil)
		assert.Error(t, err)
		assert.Equal(t, 400, err.HTTPCode)
	})

	t.Run("ResumeSubscription_EmptyID", func(t *testing.T) {
		// Test resuming with empty subscription ID
		_, err := service.ResumeSubscription(ctx, "")
		assert.Error(t, err)
		assert.Equal(t, 400, err.HTTPCode)
	})

	t.Run("GetCustomerSubscriptions", func(t *testing.T) {
		// Skip actual Stripe API calls in CI - would need real customer setup
		t.Skip("Skipping Stripe API call - requires customer setup in test account")
		
		// Test getting customer subscriptions (should return empty list for new customer)
		subscriptions, err := service.GetCustomerSubscriptions(ctx)
		assert.NoError(t, err)
		assert.NotNil(t, subscriptions)
		assert.Equal(t, 0, len(subscriptions))
	})

	t.Run("CreateCustomerPortalSession_EmptyReturnURL", func(t *testing.T) {
		// Test portal session with empty return URL
		_, err := service.CreateCustomerPortalSession(ctx, "")
		assert.Error(t, err)
		assert.Equal(t, 400, err.HTTPCode)
	})
}

// TestWebhookValidation tests webhook signature validation
func TestWebhookValidation(t *testing.T) {
	t.Run("ValidateWebhookSignature_EmptyPayload", func(t *testing.T) {
		_, err := stripe.ValidateWebhookSignature([]byte{}, "test_signature", "test_secret")
		assert.Error(t, err)
		assert.Equal(t, 400, err.HTTPCode)
	})

	t.Run("ValidateWebhookSignature_EmptySignature", func(t *testing.T) {
		_, err := stripe.ValidateWebhookSignature([]byte("test payload"), "", "test_secret")
		assert.Error(t, err)
		assert.Equal(t, 400, err.HTTPCode)
	})

	t.Run("ValidateWebhookSignature_EmptySecret", func(t *testing.T) {
		_, err := stripe.ValidateWebhookSignature([]byte("test payload"), "test_signature", "")
		assert.Error(t, err)
		assert.Equal(t, 500, err.HTTPCode)
	})

	t.Run("ValidateWebhookSignature_InvalidSignature", func(t *testing.T) {
		payload := []byte(`{"type":"test.event","data":{"object":{}}}`)
		signature := "invalid_signature"
		secret := "whsec_test_secret"
		
		_, err := stripe.ValidateWebhookSignature(payload, signature, secret)
		assert.Error(t, err)
		assert.Equal(t, 400, err.HTTPCode)
	})
}

// TestStripeConfiguration tests Stripe configuration validation
func TestStripeConfiguration(t *testing.T) {
	// Save original key
	originalKey := stripeAPI.Key

	t.Run("CreateOneTimePayment_NoAPIKey", func(t *testing.T) {
		// Test with empty Stripe API key
		stripeAPI.Key = ""

		userID := uuid.New()
		ctx := context.WithValue(context.Background(), contextUtils.UserIDKey, userID)

		_, err := stripe.CreateOneTimePayment(ctx, "price_test", 1, nil, nil, "https://www.rise-basketball.com/success")
		assert.Error(t, err)
		assert.Equal(t, 500, err.HTTPCode)
		assert.Contains(t, err.Message, "Stripe not initialized")
	})

	t.Run("CreateSubscription_NoAPIKey", func(t *testing.T) {
		// Test with empty Stripe API key
		stripeAPI.Key = ""

		userID := uuid.New()
		ctx := context.WithValue(context.Background(), contextUtils.UserIDKey, userID)

		_, err := stripe.CreateSubscription(ctx, "price_test", "", nil, "https://www.rise-basketball.com/success")
		assert.Error(t, err)
		assert.Equal(t, 500, err.HTTPCode)
		assert.Contains(t, err.Message, "Stripe not initialized")
	})

	// Restore original key
	t.Cleanup(func() {
		stripeAPI.Key = originalKey
	})
}

// TestAuthenticationValidation tests authentication requirements
func TestAuthenticationValidation(t *testing.T) {
	if os.Getenv("STRIPE_SECRET_KEY") == "" {
		t.Skip("STRIPE_SECRET_KEY not set, skipping integration tests")
	}

	stripeAPI.Key = os.Getenv("STRIPE_SECRET_KEY")

	t.Run("CreateOneTimePayment_NoAuth", func(t *testing.T) {
		// Test without user ID in context
		ctx := context.Background()

		_, err := stripe.CreateOneTimePayment(ctx, "price_test", 1, nil, nil, "https://www.rise-basketball.com/success")
		assert.Error(t, err)
		// Should fail with authentication error
	})

	t.Run("CreateSubscription_NoAuth", func(t *testing.T) {
		// Test without user ID in context
		ctx := context.Background()

		_, err := stripe.CreateSubscription(ctx, "price_test", "", nil, "https://www.rise-basketball.com/success")
		assert.Error(t, err)
		// Should fail with authentication error
	})
}

// BenchmarkStripeOperations benchmarks critical Stripe operations
func BenchmarkStripeOperations(b *testing.B) {
	if os.Getenv("STRIPE_SECRET_KEY") == "" {
		b.Skip("STRIPE_SECRET_KEY not set, skipping benchmarks")
	}

	stripeAPI.Key = os.Getenv("STRIPE_SECRET_KEY")
	userID := uuid.New()
	ctx := context.WithValue(context.Background(), contextUtils.UserIDKey, userID)

	b.Run("CreateOneTimePayment", func(b *testing.B) {
		b.Skip("Skipping Stripe API benchmark - requires valid price IDs in test account")

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := stripe.CreateOneTimePayment(ctx, "price_test_benchmark", 1, nil, nil, "https://www.rise-basketball.com/success")
			if err != nil {
				b.Fatalf("CreateOneTimePayment failed: %v", err)
			}
		}
	})

	b.Run("CreateSubscription", func(b *testing.B) {
		b.Skip("Skipping Stripe API benchmark - requires valid price IDs in test account")

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := stripe.CreateSubscription(ctx, "price_test_benchmark", "", nil, "https://www.rise-basketball.com/success")
			if err != nil {
				b.Fatalf("CreateSubscription failed: %v", err)
			}
		}
	})
}

// Helper function to create test context with user ID
func createTestContext() context.Context {
	userID := uuid.New()
	return context.WithValue(context.Background(), contextUtils.UserIDKey, userID)
}

// Test helper to assert error properties
func assertError(t *testing.T, err *errLib.CommonError, expectedStatus int, expectedMessage string) {
	require.NotNil(t, err)
	assert.Equal(t, expectedStatus, err.HTTPCode)
	if expectedMessage != "" {
		assert.Contains(t, err.Message, expectedMessage)
	}
}

// Test helper to assert successful response
func assertSuccess(t *testing.T, result interface{}, err *errLib.CommonError) {
	assert.Nil(t, err)
	assert.NotNil(t, result)
}