package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"time"

	"api/internal/di"
	"api/internal/security"
)

type HealthHandler struct {
	Container *di.Container
}

type HealthStatus struct {
	Status      string            `json:"status"`
	Timestamp   time.Time         `json:"timestamp"`
	Version     string            `json:"version,omitempty"`
	Checks      map[string]Check  `json:"checks"`
	Duration    string            `json:"duration"`
}

type Check struct {
	Status    string        `json:"status"`
	Duration  string        `json:"duration"`
	Error     string        `json:"error,omitempty"`
	Details   interface{}   `json:"details,omitempty"`
}

const (
	StatusHealthy   = "healthy"
	StatusUnhealthy = "unhealthy"
	StatusDegraded  = "degraded"
)

func NewHealthHandler(container *di.Container) *HealthHandler {
	return &HealthHandler{
		Container: container,
	}
}

// HealthCheck provides a comprehensive health check endpoint
// @Summary Health check endpoint for load balancer integration
// @Description Returns the health status of the application and its dependencies
// @Tags health
// @Accept json
// @Produce json
// @Success 200 {object} HealthStatus "Service is healthy"
// @Success 503 {object} HealthStatus "Service is unhealthy"
// @Router /health [get]
func (h *HealthHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	
	// Create context with timeout for health checks
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()
	
	checks := make(map[string]Check)
	overallStatus := StatusHealthy
	
	// Check database connectivity
	dbCheck := h.checkDatabase(ctx)
	checks["database"] = dbCheck
	if dbCheck.Status != StatusHealthy {
		overallStatus = StatusUnhealthy
	}
	
	// Check application readiness
	appCheck := h.checkApplication(ctx)
	checks["application"] = appCheck
	if appCheck.Status != StatusHealthy && overallStatus == StatusHealthy {
		overallStatus = StatusDegraded
	}
	
	// Check external dependencies (Stripe)
	stripeCheck := h.checkStripe(ctx)
	checks["stripe"] = stripeCheck
	if stripeCheck.Status != StatusHealthy && overallStatus == StatusHealthy {
		overallStatus = StatusDegraded
	}
	
	duration := time.Since(startTime)
	
	response := HealthStatus{
		Status:    overallStatus,
		Timestamp: time.Now(),
		Version:   getVersion(),
		Checks:    checks,
		Duration:  duration.String(),
	}
	
	// Set appropriate HTTP status code
	statusCode := http.StatusOK
	if overallStatus == StatusUnhealthy {
		statusCode = http.StatusServiceUnavailable
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.WriteHeader(statusCode)
	
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode health response", http.StatusInternalServerError)
	}
}

// ReadinessCheck provides a simple readiness check for Kubernetes
// @Summary Readiness check endpoint
// @Description Returns whether the service is ready to accept traffic
// @Tags health
// @Accept json
// @Produce json
// @Success 200 {object} map[string]string "Service is ready"
// @Success 503 {object} map[string]string "Service is not ready"
// @Router /ready [get]
func (h *HealthHandler) ReadinessCheck(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	
	// Check if database is accessible
	if err := h.Container.DB.PingContext(ctx); err != nil {
		http.Error(w, `{"status":"not ready","reason":"database unavailable"}`, http.StatusServiceUnavailable)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status": "ready",
	})
}

// LivenessCheck provides a simple liveness check for Kubernetes
// @Summary Liveness check endpoint
// @Description Returns whether the service is alive
// @Tags health
// @Accept json
// @Produce json
// @Success 200 {object} map[string]string "Service is alive"
// @Router /live [get]
func (h *HealthHandler) LivenessCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status": "alive",
	})
}

func (h *HealthHandler) checkDatabase(ctx context.Context) Check {
	startTime := time.Now()
	
	// Test database connection
	err := h.Container.DB.PingContext(ctx)
	duration := time.Since(startTime)
	
	if err != nil {
		return Check{
			Status:   StatusUnhealthy,
			Duration: duration.String(),
			Error:    err.Error(),
		}
	}
	
	// Test a simple query to ensure database is functional
	var result int
	err = h.Container.DB.QueryRowContext(ctx, "SELECT 1").Scan(&result)
	if err != nil {
		return Check{
			Status:   StatusUnhealthy,
			Duration: duration.String(),
			Error:    "Database query failed: " + err.Error(),
		}
	}
	
	return Check{
		Status:   StatusHealthy,
		Duration: duration.String(),
		Details: map[string]interface{}{
			"connection": "ok",
			"query":      "ok",
		},
	}
}

func (h *HealthHandler) checkApplication(ctx context.Context) Check {
	startTime := time.Now()
	
	// Check if all required environment variables are present
	requiredEnvVars := []string{
		"STRIPE_SECRET_KEY",
		"STRIPE_WEBHOOK_SECRET",
		"DATABASE_URL",
	}
	
	missing := []string{}
	for _, envVar := range requiredEnvVars {
		if value := os.Getenv(envVar); value == "" {
			missing = append(missing, envVar)
		}
	}
	
	duration := time.Since(startTime)
	
	if len(missing) > 0 {
		return Check{
			Status:   StatusUnhealthy,
			Duration: duration.String(),
			Error:    "Missing required environment variables",
			Details: map[string]interface{}{
				"missing_env_vars": missing,
			},
		}
	}
	
	return Check{
		Status:   StatusHealthy,
		Duration: duration.String(),
		Details: map[string]interface{}{
			"config": "ok",
		},
	}
}

func (h *HealthHandler) checkStripe(ctx context.Context) Check {
	startTime := time.Now()
	
	// Check if Stripe is configured
	stripeKey := os.Getenv("STRIPE_SECRET_KEY")
	stripeWebhookSecret := os.Getenv("STRIPE_WEBHOOK_SECRET")
	
	duration := time.Since(startTime)
	
	if stripeKey == "" {
		return Check{
			Status:   StatusUnhealthy,
			Duration: duration.String(),
			Error:    "Stripe API key not configured",
		}
	}
	
	if stripeWebhookSecret == "" {
		return Check{
			Status:   StatusDegraded,
			Duration: duration.String(),
			Error:    "Stripe webhook secret not configured",
		}
	}
	
	// Note: We don't make actual Stripe API calls in health checks
	// to avoid rate limiting and unnecessary external dependencies
	return Check{
		Status:   StatusHealthy,
		Duration: duration.String(),
		Details: map[string]interface{}{
			"api_key": "configured",
			"webhook_secret": "configured",
		},
	}
}

// WebhookRetryStats provides webhook retry statistics endpoint
// @Summary Webhook retry statistics
// @Description Returns statistics about webhook retry attempts
// @Tags health
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{} "Retry statistics"
// @Router /health/webhook-retries [get]
func (h *HealthHandler) WebhookRetryStats(w http.ResponseWriter, r *http.Request) {
	// This would need to be injected if we want real retry stats
	// For now, return a placeholder response
	stats := map[string]interface{}{
		"status": "webhook retry system operational",
		"note":   "detailed stats require webhook retry service integration",
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	
	if err := json.NewEncoder(w).Encode(stats); err != nil {
		http.Error(w, "Failed to encode retry stats", http.StatusInternalServerError)
	}
}

// SecurityAudit provides comprehensive security assessment endpoint
// @Summary Security audit
// @Description Performs comprehensive security audit and returns results
// @Tags security
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{} "Security audit results"
// @Router /health/security-audit [get]
func (h *HealthHandler) SecurityAudit(w http.ResponseWriter, r *http.Request) {
	audit := security.NewSecurityAudit()
	result := audit.PerformComprehensiveAudit(r.Context())
	
	// Set appropriate status code based on security level
	statusCode := http.StatusOK
	if result.Overall == security.SecurityLevelCritical {
		statusCode = http.StatusServiceUnavailable
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	
	if err := json.NewEncoder(w).Encode(result); err != nil {
		http.Error(w, "Failed to encode security audit results", http.StatusInternalServerError)
	}
}

func getVersion() string {
	// This could be set at build time using ldflags
	// go build -ldflags "-X main.Version=1.0.0"
	version := "development"
	return version
}