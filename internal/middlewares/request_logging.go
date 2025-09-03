package middlewares

import (
	"context"
	"net/http"
	"time"

	"api/internal/libs/logger"
	"github.com/google/uuid"
)

// RequestLoggingMiddleware creates a middleware that logs HTTP requests with structured logging
func RequestLoggingMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			
			// Generate request ID for tracing
			requestID := uuid.New().String()
			
			// Add request ID to context
			ctx := context.WithValue(r.Context(), "request_id", requestID)
			r = r.WithContext(ctx)
			
			// Create a response writer wrapper to capture status code and size
			wrappedWriter := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
			
			// Create structured logger with request information
			reqLogger := logger.WithFields(map[string]interface{}{
				"request_id":     requestID,
				"method":         r.Method,
				"path":           r.URL.Path,
				"query":          r.URL.RawQuery,
				"remote_addr":    r.RemoteAddr,
				"user_agent":     r.Header.Get("User-Agent"),
				"content_length": r.ContentLength,
				"host":           r.Host,
			})
			
			// Log incoming request
			reqLogger.Info("HTTP request started")
			
			// Process request
			next.ServeHTTP(wrappedWriter, r)
			
			// Calculate duration
			duration := time.Since(start)
			
			// Log completed request
			reqLogger.WithFields(map[string]interface{}{
				"status_code":    wrappedWriter.statusCode,
				"response_size":  wrappedWriter.bytesWritten,
				"duration_ms":    duration.Milliseconds(),
				"duration":       duration.String(),
			}).Info("HTTP request completed")
			
			// Log errors and slow requests specially
			if wrappedWriter.statusCode >= 400 {
				reqLogger.WithFields(map[string]interface{}{
					"status_code": wrappedWriter.statusCode,
					"duration":    duration.String(),
				}).Warn("HTTP request resulted in error status")
			}
			
			if duration > 5*time.Second {
				reqLogger.WithFields(map[string]interface{}{
					"duration":    duration.String(),
					"status_code": wrappedWriter.statusCode,
				}).Warn("Slow HTTP request detected")
			}
		})
	}
}

// responseWriter wraps http.ResponseWriter to capture status code and response size
type responseWriter struct {
	http.ResponseWriter
	statusCode    int
	bytesWritten  int64
	headerWritten bool
}

func (rw *responseWriter) WriteHeader(statusCode int) {
	if !rw.headerWritten {
		rw.statusCode = statusCode
		rw.headerWritten = true
		rw.ResponseWriter.WriteHeader(statusCode)
	}
}

func (rw *responseWriter) Write(data []byte) (int, error) {
	if !rw.headerWritten {
		rw.WriteHeader(http.StatusOK)
	}
	
	n, err := rw.ResponseWriter.Write(data)
	rw.bytesWritten += int64(n)
	return n, err
}

// Implement other interfaces if needed
func (rw *responseWriter) Header() http.Header {
	return rw.ResponseWriter.Header()
}

// Implement http.Flusher if the underlying ResponseWriter supports it
func (rw *responseWriter) Flush() {
	if f, ok := rw.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

// SecurityLoggingMiddleware logs security-related events
func SecurityLoggingMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Log potentially suspicious activities
			securityLogger := logger.WithComponent("security").WithFields(map[string]interface{}{
				"method":      r.Method,
				"path":        r.URL.Path,
				"remote_addr": r.RemoteAddr,
				"user_agent":  r.Header.Get("User-Agent"),
			})
			
			// Check for suspicious patterns
			if isSuspiciousRequest(r) {
				securityLogger.WithFields(map[string]interface{}{
					"reason": "suspicious_patterns_detected",
				}).Warn("Potentially malicious request detected")
			}
			
			// Log failed authentication attempts (this would be enhanced based on your auth system)
			if r.Header.Get("Authorization") != "" && isAuthenticationEndpoint(r.URL.Path) {
				securityLogger.WithField("has_auth_header", true).Info("Authentication attempt")
			}
			
			next.ServeHTTP(w, r)
		})
	}
}

// isSuspiciousRequest checks for common attack patterns
func isSuspiciousRequest(r *http.Request) bool {
	suspiciousPatterns := []string{
		"../",
		"<script",
		"javascript:",
		"eval(",
		"union select",
		"drop table",
		"insert into",
		"update set",
		"delete from",
	}
	
	// Check URL path and query parameters
	fullURL := r.URL.Path + "?" + r.URL.RawQuery
	for _, pattern := range suspiciousPatterns {
		if containsIgnoreCase(fullURL, pattern) {
			return true
		}
	}
	
	// Check headers
	for _, values := range r.Header {
		for _, value := range values {
			for _, pattern := range suspiciousPatterns {
				if containsIgnoreCase(value, pattern) {
					return true
				}
			}
		}
	}
	
	return false
}

// isAuthenticationEndpoint checks if the path is an authentication endpoint
func isAuthenticationEndpoint(path string) bool {
	authEndpoints := []string{
		"/auth/login",
		"/auth/register",
		"/auth/refresh",
		"/login",
		"/register",
	}
	
	for _, endpoint := range authEndpoints {
		if path == endpoint {
			return true
		}
	}
	
	return false
}

// containsIgnoreCase performs case-insensitive substring search
func containsIgnoreCase(s, substr string) bool {
	return len(s) >= len(substr) && 
		(s == substr || 
		 len(s) > len(substr) && 
		 (containsIgnoreCase(s[1:], substr) || 
		  (len(substr) > 0 && (s[0]|32) == (substr[0]|32) && 
		   containsIgnoreCase(s[1:], substr[1:]))))
}