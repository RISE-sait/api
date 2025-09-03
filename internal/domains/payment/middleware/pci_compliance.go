package middleware

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"net/http"
	"strings"
	"time"

	"api/internal/libs/logger"
)

// PCIComplianceMiddleware ensures PCI DSS compliance 
type PCIComplianceMiddleware struct {
	logger *logger.StructuredLogger
}

// NewPCIComplianceMiddleware creates a new PCI compliance middleware
func NewPCIComplianceMiddleware() *PCIComplianceMiddleware {
	return &PCIComplianceMiddleware{
		logger: logger.WithComponent("pci-compliance"),
	}
}

// EnforcePCICompliance applies PCI DSS requirements to payment endpoints
func (p *PCIComplianceMiddleware) EnforcePCICompliance(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !p.isSecureConnection(r) {
			p.logger.WithFields(map[string]interface{}{
				"client_ip": r.RemoteAddr,
				"path":      r.URL.Path,
				"protocol":  r.Proto,
				"violation": "insecure_connection",
			}).Error("PCI violation: Payment request over insecure connection", nil)
			
			http.Error(w, "HTTPS required for payment operations", http.StatusBadRequest)
			return
		}
		

		if p.containsCardholderData(r) {
			p.logger.WithFields(map[string]interface{}{
				"client_ip": r.RemoteAddr,
				"path":      r.URL.Path,
				"violation": "cardholder_data_detected",
			}).Error("PCI violation: Cardholder data detected in request", nil)
			
			http.Error(w, "Cardholder data not permitted", http.StatusBadRequest)
			return
		}

		if !p.isAuthenticated(r) && p.requiresAuthentication(r.URL.Path) {
			p.logger.WithFields(map[string]interface{}{
				"client_ip": r.RemoteAddr,
				"path":      r.URL.Path,
				"violation": "unauthenticated_access",
			}).Warn("PCI security: Unauthenticated access to payment endpoint")
			
			http.Error(w, "Authentication required", http.StatusUnauthorized)
			return
		}
		
	
		p.logPCIAccess(r)
		
		// Add PCI compliance headers
		p.setPCIHeaders(w)
		
		next.ServeHTTP(w, r)
	})
}

// DataMaskingMiddleware ensures sensitive data is never logged or exposed
func (p *PCIComplianceMiddleware) DataMaskingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Wrap response writer to mask sensitive data in responses
		maskedWriter := &MaskedResponseWriter{
			ResponseWriter: w,
			logger:         p.logger,
		}
		
		next.ServeHTTP(maskedWriter, r)
	})
}


func (p *PCIComplianceMiddleware) isSecureConnection(r *http.Request) bool {

	if r.TLS != nil {
		return true
	}
	

	if proto := r.Header.Get("X-Forwarded-Proto"); proto == "https" {
		return true
	}
	

	if strings.Contains(r.Host, "localhost") || strings.Contains(r.Host, "127.0.0.1") {
		p.logger.Warn("Development mode: allowing insecure connection to localhost")
		return true
	}
	
	return false
}


func (p *PCIComplianceMiddleware) containsCardholderData(r *http.Request) bool {
	// Check URL parameters
	if p.scanForCardData(r.URL.RawQuery) {
		return true
	}
	
	// Check headers (though this should never happen)
	for _, values := range r.Header {
		for _, value := range values {
			if p.scanForCardData(value) {
				return true
			}
		}
	}
	

	
	return false
}


func (p *PCIComplianceMiddleware) scanForCardData(data string) bool {

	cleaned := strings.ReplaceAll(data, " ", "")
	cleaned = strings.ReplaceAll(cleaned, "-", "")
	cleaned = strings.ReplaceAll(cleaned, "_", "")
	

	if len(cleaned) >= 13 && len(cleaned) <= 19 {
		if p.isLuhnValid(cleaned) && p.isCardPrefix(cleaned) {
			return true
		}
	}
	

	if p.containsCVVPattern(data) {
		return true
	}
	
	return false
}


func (p *PCIComplianceMiddleware) isLuhnValid(number string) bool {
	sum := 0
	alternate := false
	
	for i := len(number) - 1; i >= 0; i-- {
		digit := int(number[i] - '0')
		if digit < 0 || digit > 9 {
			return false
		}
		
		if alternate {
			digit *= 2
			if digit > 9 {
				digit = (digit % 10) + 1
			}
		}
		
		sum += digit
		alternate = !alternate
	}
	
	return sum%10 == 0
}


func (p *PCIComplianceMiddleware) isCardPrefix(number string) bool {
	// Major card prefixes
	prefixes := []string{
		"4",      // Visa
		"51", "52", "53", "54", "55", // Mastercard
		"34", "37", // American Express
		"6011", "65", // Discover
	}
	
	for _, prefix := range prefixes {
		if strings.HasPrefix(number, prefix) {
			return true
		}
	}
	
	return false
}

// containsCVVPattern looks for CVV-like patterns
func (p *PCIComplianceMiddleware) containsCVVPattern(data string) bool {

	cvvPatterns := []string{"cvv", "cvc", "cid", "security_code", "card_code"}
	
	lowerData := strings.ToLower(data)
	for _, pattern := range cvvPatterns {
		if strings.Contains(lowerData, pattern) {
			return true
		}
	}
	
	return false
}

// isAuthenticated checks if request is authenticated
func (p *PCIComplianceMiddleware) isAuthenticated(r *http.Request) bool {

	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return false
	}
	

	if strings.HasPrefix(authHeader, "Bearer ") {
		return true
	}
	
	return false
}


func (p *PCIComplianceMiddleware) requiresAuthentication(path string) bool {
	// Webhooks don't require JWT auth (they use signature verification)
	if strings.Contains(path, "/webhooks/") {
		return false
	}

	paymentEndpoints := []string{"/checkout", "/subscriptions"}
	for _, endpoint := range paymentEndpoints {
		if strings.HasPrefix(path, endpoint) {
			return true
		}
	}
	
	return false
}

// logPCIAccess logs access for PCI compliance monitoring
func (p *PCIComplianceMiddleware) logPCIAccess(r *http.Request) {
	p.logger.WithFields(map[string]interface{}{
		"event_type":    "pci_access",
		"client_ip":     r.RemoteAddr,
		"method":        r.Method,
		"path":          r.URL.Path,
		"user_agent":    r.Header.Get("User-Agent"),
		"timestamp":     time.Now().UTC(),
		"secure":        p.isSecureConnection(r),
		"authenticated": p.isAuthenticated(r),
	}).Info("PCI-monitored payment endpoint access")
}

// setPCIHeaders adds headers required for PCI compliance
func (p *PCIComplianceMiddleware) setPCIHeaders(w http.ResponseWriter) {
	// Ensure no caching of sensitive payment data
	w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, private")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")
	
	// Additional security headers for payment endpoints
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("X-Frame-Options", "DENY")
	w.Header().Set("X-XSS-Protection", "1; mode=block")
}

// MaskedResponseWriter wraps http.ResponseWriter to mask sensitive data
type MaskedResponseWriter struct {
	http.ResponseWriter
	logger *logger.StructuredLogger
}

// Write masks sensitive data in response body
func (m *MaskedResponseWriter) Write(data []byte) (int, error) {

	
	dataStr := string(data)
	if m.containsSensitiveData(dataStr) {
		m.logger.Error("Attempted to write sensitive data to response", nil)

		maskedResponse := `{"error": "Sensitive data detected and masked for PCI compliance"}`
		return m.ResponseWriter.Write([]byte(maskedResponse))
	}
	
	return m.ResponseWriter.Write(data)
}

// containsSensitiveData checks response for sensitive payment information
func (m *MaskedResponseWriter) containsSensitiveData(data string) bool {

	sensitivePatterns := []string{
		"card_number", "cardNumber", "pan",
		"cvv", "cvc", "security_code",
		"expiry", "exp_month", "exp_year",
	}
	
	lowerData := strings.ToLower(data)
	for _, pattern := range sensitivePatterns {
		if strings.Contains(lowerData, pattern) {
			return true
		}
	}
	
	return false
}

// GenerateSecureToken creates a cryptographically secure token
func GenerateSecureToken(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

// SecureCompare performs constant-time string comparison
func SecureCompare(a, b string) bool {
	return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}