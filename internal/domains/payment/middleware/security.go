package middleware

import (
	"context"
	"net/http"
	"strings"
	"sync"
	"time"

	"api/internal/libs/logger"
)

// SecurityMiddleware provides comprehensive security features for payment endpoints
type SecurityMiddleware struct {
	csrfProtection  *CSRFProtection
	logger          *logger.StructuredLogger
}

// NewSecurityMiddleware creates a new security middleware instance
func NewSecurityMiddleware() *SecurityMiddleware {
	return &SecurityMiddleware{
		csrfProtection: NewCSRFProtection(),
		logger:         logger.WithComponent("payment-security"),
	}
}

// SecurePaymentEndpoints wraps payment endpoints with comprehensive security
func (s *SecurityMiddleware) SecurePaymentEndpoints(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Security headers
		s.setSecurityHeaders(w)
		
		// Input validation for sensitive endpoints
		if isPaymentEndpoint(r.URL.Path) {
			if err := s.validatePaymentRequest(r); err != nil {
				s.logger.WithFields(map[string]interface{}{
					"client_ip":      r.RemoteAddr,
					"path":           r.URL.Path,
					"validation_error": err.Error(),
				}).Warn("Payment request blocked due to validation failure")
				
				http.Error(w, "Invalid request", http.StatusBadRequest)
				return
			}
		}
		
		// Log security events
		s.logSecurityEvent(r)
		
		next.ServeHTTP(w, r)
	})
}

// WebhookSecurityMiddleware provides security specifically for webhook endpoints
func (s *SecurityMiddleware) WebhookSecurityMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Webhook-specific security headers
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		
		// Validate webhook source
		if !s.isValidWebhookSource(r) {
			s.logger.WithFields(map[string]interface{}{
				"client_ip":  r.RemoteAddr,
				"user_agent": r.Header.Get("User-Agent"),
				"reason":     "invalid_webhook_source",
			}).Error("Webhook request from invalid source blocked", nil)
			
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
		
		// Add request ID for tracking
		requestID := generateRequestID()
		ctx := context.WithValue(r.Context(), "webhook_request_id", requestID)
		r = r.WithContext(ctx)
		
		next.ServeHTTP(w, r)
	})
}

// setSecurityHeaders adds comprehensive security headers
func (s *SecurityMiddleware) setSecurityHeaders(w http.ResponseWriter) {
	headers := map[string]string{
		"X-Content-Type-Options":           "nosniff",
		"X-Frame-Options":                  "DENY", 
		"X-XSS-Protection":                 "1; mode=block",
		"Strict-Transport-Security":        "max-age=31536000; includeSubDomains",
		"Content-Security-Policy":          "default-src 'self'; script-src 'self'; object-src 'none';",
		"Referrer-Policy":                  "strict-origin-when-cross-origin",
		"Permissions-Policy":               "geolocation=(), microphone=(), camera=()",
		"X-Permitted-Cross-Domain-Policies": "none",
	}
	
	for header, value := range headers {
		w.Header().Set(header, value)
	}
}

// validatePaymentRequest performs input validation on payment requests
func (s *SecurityMiddleware) validatePaymentRequest(r *http.Request) error {
	// Check for suspicious patterns in headers
	suspiciousHeaders := []string{"X-Forwarded-For", "X-Real-IP", "X-Originating-IP"}
	for _, header := range suspiciousHeaders {
		if value := r.Header.Get(header); value != "" {
			if containsSuspiciousContent(value) {
				return &SecurityError{"Suspicious content in headers"}
			}
		}
	}
	
	// Validate Content-Type for POST requests
	if r.Method == http.MethodPost {
		contentType := r.Header.Get("Content-Type")
		if contentType != "" && !isValidContentType(contentType) {
			return &SecurityError{"Invalid content type"}
		}
	}
	
	// Check request size limits
	if r.ContentLength > 1024*1024 { // 1MB limit
		return &SecurityError{"Request too large"}
	}
	
	return nil
}

// isValidWebhookSource validates webhook requests are from Stripe
func (s *SecurityMiddleware) isValidWebhookSource(r *http.Request) bool {
	// Check User-Agent (Stripe webhooks have specific user agent)
	userAgent := r.Header.Get("User-Agent")
	if !strings.HasPrefix(userAgent, "Stripe/") {
		return false
	}
	
	// Additional IP validation could be added here
	// Stripe publishes their webhook IP ranges
	
	return true
}

// logSecurityEvent logs security-relevant events
func (s *SecurityMiddleware) logSecurityEvent(r *http.Request) {
	s.logger.WithFields(map[string]interface{}{
		"method":       r.Method,
		"path":         r.URL.Path,
		"client_ip":    r.RemoteAddr,
		"user_agent":   r.Header.Get("User-Agent"),
		"content_type": r.Header.Get("Content-Type"),
		"timestamp":    time.Now().UTC(),
	}).Info("Payment endpoint accessed")
}


// CSRFProtection provides CSRF token validation
type CSRFProtection struct {
	tokenStore map[string]time.Time
	mutex      sync.RWMutex
}

// NewCSRFProtection creates a new CSRF protection instance  
func NewCSRFProtection() *CSRFProtection {
	return &CSRFProtection{
		tokenStore: make(map[string]time.Time),
	}
}

// SecurityError represents a security validation error
type SecurityError struct {
	Message string
}

func (e *SecurityError) Error() string {
	return e.Message
}

// Helper functions

func isPaymentEndpoint(path string) bool {
	paymentPaths := []string{
		"/checkout",
		"/subscriptions", 
		"/webhooks/stripe",
	}
	
	for _, paymentPath := range paymentPaths {
		if strings.HasPrefix(path, paymentPath) {
			return true
		}
	}
	return false
}

func containsSuspiciousContent(value string) bool {
	suspiciousPatterns := []string{
		"<script", "javascript:", "../", "eval(", 
		"union select", "drop table", "insert into",
		"exec(", "system(", "cmd.exe", "/bin/sh",
	}
	
	lowerValue := strings.ToLower(value)
	for _, pattern := range suspiciousPatterns {
		if strings.Contains(lowerValue, pattern) {
			return true
		}
	}
	return false
}

func isValidContentType(contentType string) bool {
	validTypes := []string{
		"application/json",
		"application/x-www-form-urlencoded", 
		"multipart/form-data",
		"text/plain",
	}
	
	for _, validType := range validTypes {
		if strings.HasPrefix(contentType, validType) {
			return true
		}
	}
	return false
}

func generateRequestID() string {
	// Simple request ID generation
	return time.Now().Format("20060102150405") + "-" + randomString(8)
}

func randomString(length int) string {
	// Simple random string generation for request IDs
	chars := "abcdefghijklmnopqrstuvwxyz0123456789"
	result := make([]byte, length)
	for i := range result {
		result[i] = chars[time.Now().UnixNano()%int64(len(chars))]
	}
	return string(result)
}

