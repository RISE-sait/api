package stripe_test

import (
	stripe "api/internal/domains/payment/services/stripe"
	"errors"
	"testing"

	stripeAPI "github.com/stripe/stripe-go/v81"
)

func TestCentsToDollars(t *testing.T) {
	tests := []struct {
		cents    int64
		expected float64
	}{
		{0, 0.0},
		{100, 1.0},
		{199, 1.99},
		{1, 0.01},
		{999, 9.99},
		{10050, 100.50},
		{-500, -5.0},
	}

	for _, tc := range tests {
		result := stripe.CentsToDollars(tc.cents)
		if result != tc.expected {
			t.Errorf("CentsToDollars(%d) = %f, want %f", tc.cents, result, tc.expected)
		}
	}
}

func TestDollarsToCents(t *testing.T) {
	tests := []struct {
		dollars  float64
		expected int64
	}{
		{0.0, 0},
		{1.0, 100},
		{1.99, 199},
		{0.01, 1},
		{9.99, 999},
		{100.50, 10050},
		{-5.0, -500},
		// Edge case: float precision — 19.99 * 100 = 1998.9999... with naive float math
		{19.99, 1999},
	}

	for _, tc := range tests {
		result := stripe.DollarsToCents(tc.dollars)
		if result != tc.expected {
			t.Errorf("DollarsToCents(%f) = %d, want %d", tc.dollars, result, tc.expected)
		}
	}
}

func TestCentsToDollarsRoundTrip(t *testing.T) {
	// Verify round-trip: cents -> dollars -> cents preserves value
	testCents := []int64{0, 1, 99, 100, 199, 1999, 9999, 10050, 99999}
	for _, cents := range testCents {
		dollars := stripe.CentsToDollars(cents)
		back := stripe.DollarsToCents(dollars)
		if back != cents {
			t.Errorf("Round-trip failed: %d -> %f -> %d", cents, dollars, back)
		}
	}
}

func TestRetryStripeCall_SucceedsFirstAttempt(t *testing.T) {
	calls := 0
	err := stripe.RetryStripeCall("test", 3, func() error {
		calls++
		return nil
	})
	if err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
	if calls != 1 {
		t.Errorf("expected 1 call, got %d", calls)
	}
}

func TestRetryStripeCall_RetriesOnServerError(t *testing.T) {
	calls := 0
	err := stripe.RetryStripeCall("test", 3, func() error {
		calls++
		if calls < 3 {
			return &stripeAPI.Error{HTTPStatusCode: 500, Msg: "server error"}
		}
		return nil
	})
	if err != nil {
		t.Errorf("expected nil error after retries, got %v", err)
	}
	if calls != 3 {
		t.Errorf("expected 3 calls, got %d", calls)
	}
}

func TestRetryStripeCall_NoRetryOn4xx(t *testing.T) {
	calls := 0
	err := stripe.RetryStripeCall("test", 3, func() error {
		calls++
		return &stripeAPI.Error{HTTPStatusCode: 400, Msg: "bad request"}
	})
	if err == nil {
		t.Error("expected error, got nil")
	}
	if calls != 1 {
		t.Errorf("expected 1 call (no retry on 4xx), got %d", calls)
	}
}

func TestRetryStripeCall_RetriesOnNetworkError(t *testing.T) {
	calls := 0
	err := stripe.RetryStripeCall("test", 2, func() error {
		calls++
		return errors.New("connection refused")
	})
	if err == nil {
		t.Error("expected error, got nil")
	}
	if calls != 2 {
		t.Errorf("expected 2 calls, got %d", calls)
	}
}
