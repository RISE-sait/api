package middleware

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	errLib "api/internal/libs/errors"
	responseHandlers "api/internal/libs/responses"
	contextUtils "api/utils/context"
)

// TokenBucket represents a rate limiter using token bucket algorithm
type TokenBucket struct {
	tokens       int
	maxTokens    int
	refillRate   int           // tokens per second
	lastRefill   time.Time
	mutex        sync.Mutex
}

// RateLimiter manages rate limiting for different endpoints
type RateLimiter struct {
	buckets map[string]*TokenBucket
	mutex   sync.RWMutex
}

// NewRateLimiter creates a new rate limiter instance
func NewRateLimiter() *RateLimiter {
	return &RateLimiter{
		buckets: make(map[string]*TokenBucket),
	}
}

// Global rate limiter instance
var globalRateLimiter = NewRateLimiter()

// newTokenBucket creates a new token bucket with specified parameters
func newTokenBucket(maxTokens, refillRate int) *TokenBucket {
	return &TokenBucket{
		tokens:     maxTokens,
		maxTokens:  maxTokens,
		refillRate: refillRate,
		lastRefill: time.Now(),
	}
}

// allowRequest checks if a request should be allowed based on rate limiting
func (tb *TokenBucket) allowRequest() bool {
	tb.mutex.Lock()
	defer tb.mutex.Unlock()

	now := time.Now()
	timePassed := now.Sub(tb.lastRefill).Seconds()
	
	// Refill tokens based on time passed
	tokensToAdd := int(timePassed * float64(tb.refillRate))
	if tokensToAdd > 0 {
		tb.tokens = min(tb.maxTokens, tb.tokens+tokensToAdd)
		tb.lastRefill = now
	}

	// Check if we have tokens available
	if tb.tokens > 0 {
		tb.tokens--
		return true
	}
	
	return false
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// getRateLimitKey generates a unique key for rate limiting based on user ID and endpoint
func getRateLimitKey(ctx context.Context, endpoint string) (string, error) {
	userID, err := contextUtils.GetUserID(ctx)
	if err != nil {
		// For unauthenticated requests, we could use IP address
		// but for now return error since all our endpoints require auth
		return "", fmt.Errorf("rate limiting requires authenticated user")
	}
	
	return fmt.Sprintf("user:%s:endpoint:%s", userID.String(), endpoint), nil
}

// getBucket gets or creates a token bucket for a specific key
func (rl *RateLimiter) getBucket(key string, maxTokens, refillRate int) *TokenBucket {
	rl.mutex.RLock()
	bucket, exists := rl.buckets[key]
	rl.mutex.RUnlock()
	
	if exists {
		return bucket
	}
	
	// Create new bucket
	rl.mutex.Lock()
	defer rl.mutex.Unlock()
	
	// Double-check after acquiring write lock
	if bucket, exists := rl.buckets[key]; exists {
		return bucket
	}
	
	bucket = newTokenBucket(maxTokens, refillRate)
	rl.buckets[key] = bucket
	return bucket
}

// RateLimit creates a rate limiting middleware
func RateLimit(maxRequests, refillRate int, endpoint string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Generate rate limit key
			key, err := getRateLimitKey(r.Context(), endpoint)
			if err != nil {
				log.Printf("[RATE_LIMIT] Failed to generate rate limit key: %v", err)
				// Allow the request but log the issue
				next.ServeHTTP(w, r)
				return
			}
			
			// Get or create token bucket for this key
			bucket := globalRateLimiter.getBucket(key, maxRequests, refillRate)
			
			// Check if request is allowed
			if !bucket.allowRequest() {
				log.Printf("[RATE_LIMIT] Rate limit exceeded for %s", key)
				responseHandlers.RespondWithError(w, errLib.New("Rate limit exceeded", http.StatusTooManyRequests))
				return
			}
			
			// Request is allowed, proceed
			next.ServeHTTP(w, r)
		})
	}
}

// Predefined rate limiters for different endpoint types

// SubscriptionRateLimit applies rate limiting for subscription management endpoints
func SubscriptionRateLimit() func(http.Handler) http.Handler {
	// 10 requests per minute for subscription management
	return RateLimit(10, 1, "subscription")
}

// CheckoutRateLimit applies rate limiting for checkout endpoints  
func CheckoutRateLimit() func(http.Handler) http.Handler {
	// 5 requests per minute for checkout to prevent abuse
	return RateLimit(5, 1, "checkout")
}

// WebhookRateLimit applies rate limiting for webhook endpoints
func WebhookRateLimit() func(http.Handler) http.Handler {
	// 100 requests per minute for webhooks (Stripe can send many)
	return RateLimit(100, 10, "webhook")
}

// PortalRateLimit applies rate limiting for customer portal access
func PortalRateLimit() func(http.Handler) http.Handler {
	// 3 requests per minute for portal access
	return RateLimit(3, 1, "portal")
}

// SecurityHeaders adds comprehensive security headers
func SecurityHeaders() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Security headers
			w.Header().Set("X-Content-Type-Options", "nosniff")
			w.Header().Set("X-Frame-Options", "DENY")
			w.Header().Set("X-XSS-Protection", "1; mode=block")
			w.Header().Set("Strict-Transport-Security", "max-age=63072000; includeSubDomains; preload")
			w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
			w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'; img-src 'self' data: https:; font-src 'self'; connect-src 'self' https://api.stripe.com; frame-src 'none'; object-src 'none';")
			
			// Prevent caching of sensitive endpoints
			if r.URL.Path == "/webhooks/stripe" || 
			   r.URL.Path == "/subscriptions/portal" ||
			   r.URL.Path == "/checkout" {
				w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
				w.Header().Set("Pragma", "no-cache")
				w.Header().Set("Expires", "0")
			}
			
			next.ServeHTTP(w, r)
		})
	}
}

// LogRequest logs request details for monitoring
func LogRequest() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			
			// Capture user ID if available for logging
			var userID string
			if uid, err := contextUtils.GetUserID(r.Context()); err == nil {
				userID = uid.String()
			}
			
			// Create a response writer wrapper to capture status code
			wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
			
			next.ServeHTTP(wrapped, r)
			
			duration := time.Since(start)
			log.Printf("[REQUEST] %s %s - User: %s - Status: %d - Duration: %v", 
				r.Method, r.URL.Path, userID, wrapped.statusCode, duration)
		})
	}
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// ValidateStripeSignature validates Stripe webhook signatures
func ValidateStripeSignature() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Only apply to Stripe webhook endpoints
			if r.URL.Path != "/webhooks/stripe" {
				next.ServeHTTP(w, r)
				return
			}
			
			// Check for Stripe signature header
			signature := r.Header.Get("Stripe-Signature")
			if signature == "" {
				log.Printf("[SECURITY] Missing Stripe signature for webhook")
				responseHandlers.RespondWithError(w, errLib.New("Missing Stripe signature", http.StatusBadRequest))
				return
			}
			
			// Additional validation can be added here
			next.ServeHTTP(w, r)
		})
	}
}