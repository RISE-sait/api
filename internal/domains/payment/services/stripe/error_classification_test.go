package stripe

import (
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	stripe "github.com/stripe/stripe-go/v81"
)

// ============================================================
// Fix 11: classifyStripeError — maps Stripe errors to HTTP codes
// ============================================================

func TestClassifyStripeError_400BadRequest(t *testing.T) {
	err := &stripe.Error{HTTPStatusCode: 400, Msg: "Invalid parameter"}
	status, msg := classifyStripeError(err)
	assert.Equal(t, http.StatusBadRequest, status)
	assert.Contains(t, msg, "Invalid request")
	assert.Contains(t, msg, "Invalid parameter")
}

func TestClassifyStripeError_401Unauthorized(t *testing.T) {
	err := &stripe.Error{HTTPStatusCode: 401, Msg: "Invalid API Key"}
	status, msg := classifyStripeError(err)
	assert.Equal(t, http.StatusInternalServerError, status) // Don't expose auth errors to customer
	assert.Contains(t, msg, "authentication error")
}

func TestClassifyStripeError_402PaymentRequired(t *testing.T) {
	err := &stripe.Error{HTTPStatusCode: 402, Msg: "Your card was declined"}
	status, msg := classifyStripeError(err)
	assert.Equal(t, http.StatusPaymentRequired, status)
	assert.Contains(t, msg, "Payment failed")
	assert.Contains(t, msg, "Your card was declined")
}

func TestClassifyStripeError_403Forbidden(t *testing.T) {
	err := &stripe.Error{HTTPStatusCode: 403, Msg: ""}
	status, _ := classifyStripeError(err)
	assert.Equal(t, http.StatusForbidden, status)
}

func TestClassifyStripeError_404NotFound(t *testing.T) {
	err := &stripe.Error{HTTPStatusCode: 404, Msg: "No such price: 'price_xxx'"}
	status, msg := classifyStripeError(err)
	assert.Equal(t, http.StatusNotFound, status)
	assert.Contains(t, msg, "not found")
}

func TestClassifyStripeError_409Conflict(t *testing.T) {
	err := &stripe.Error{HTTPStatusCode: 409, Msg: "conflict"}
	status, _ := classifyStripeError(err)
	assert.Equal(t, http.StatusConflict, status)
}

func TestClassifyStripeError_429RateLimit(t *testing.T) {
	err := &stripe.Error{HTTPStatusCode: 429, Msg: "Rate limit"}
	status, msg := classifyStripeError(err)
	assert.Equal(t, http.StatusTooManyRequests, status)
	assert.Contains(t, msg, "rate limit")
}

func TestClassifyStripeError_500ServerError(t *testing.T) {
	err := &stripe.Error{HTTPStatusCode: 500, Msg: "Internal error"}
	status, msg := classifyStripeError(err)
	assert.Equal(t, http.StatusBadGateway, status) // We return 502 — it's Stripe's fault, not ours
	assert.Contains(t, msg, "temporarily unavailable")
}

func TestClassifyStripeError_502BadGateway(t *testing.T) {
	err := &stripe.Error{HTTPStatusCode: 502, Msg: ""}
	status, _ := classifyStripeError(err)
	assert.Equal(t, http.StatusBadGateway, status)
}

func TestClassifyStripeError_503ServiceUnavailable(t *testing.T) {
	err := &stripe.Error{HTTPStatusCode: 503, Msg: ""}
	status, _ := classifyStripeError(err)
	assert.Equal(t, http.StatusBadGateway, status)
}

func TestClassifyStripeError_NonStripeError(t *testing.T) {
	err := errors.New("network timeout connecting to Stripe")
	status, msg := classifyStripeError(err)
	assert.Equal(t, http.StatusInternalServerError, status)
	assert.Contains(t, msg, "Payment processing error")
	assert.Contains(t, msg, "network timeout")
}

// ============================================================
// Fix 3: idempotencyKey — deterministic key generation
// ============================================================

func TestIdempotencyKey_Single(t *testing.T) {
	key := idempotencyKey("checkout")
	assert.NotNil(t, key)
	assert.Equal(t, "checkout", *key)
}

func TestIdempotencyKey_Multiple(t *testing.T) {
	key := idempotencyKey("checkout", "user-123", "price_abc")
	assert.Equal(t, "checkout:user-123:price_abc", *key)
}

func TestIdempotencyKey_Deterministic(t *testing.T) {
	key1 := idempotencyKey("a", "b", "c")
	key2 := idempotencyKey("a", "b", "c")
	assert.Equal(t, *key1, *key2, "same inputs should produce same key")
}

func TestIdempotencyKey_DifferentInputsDifferentKeys(t *testing.T) {
	key1 := idempotencyKey("a", "b")
	key2 := idempotencyKey("b", "a")
	assert.NotEqual(t, *key1, *key2, "different order should produce different keys")
}

func TestIdempotencyKey_Empty(t *testing.T) {
	key := idempotencyKey()
	assert.Equal(t, "", *key)
}
