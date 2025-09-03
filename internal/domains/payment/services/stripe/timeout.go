package stripe

import (
	"context"
	"time"
)

const (
	// DefaultStripeTimeout is the default timeout for Stripe operations
	DefaultStripeTimeout = 30 * time.Second
	
	// CriticalStripeTimeout is timeout for critical operations like payments
	CriticalStripeTimeout = 60 * time.Second
	
	// QuickStripeTimeout is timeout for quick operations like retrievals
	QuickStripeTimeout = 15 * time.Second
)

// withStripeTimeout creates a context with timeout for Stripe operations
func withStripeTimeout(ctx context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
	if timeout <= 0 {
		timeout = DefaultStripeTimeout
	}
	return context.WithTimeout(ctx, timeout)
}

// withDefaultTimeout creates a context with default Stripe timeout
func withDefaultTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	return withStripeTimeout(ctx, DefaultStripeTimeout)
}

// withCriticalTimeout creates a context with extended timeout for critical operations
func withCriticalTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	return withStripeTimeout(ctx, CriticalStripeTimeout)
}

// withQuickTimeout creates a context with shorter timeout for quick operations
func withQuickTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	return withStripeTimeout(ctx, QuickStripeTimeout)
}