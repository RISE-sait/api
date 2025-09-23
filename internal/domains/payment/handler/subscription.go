package payment

import (
	"net/http"
	"strings"
	"time"

	"api/internal/di"
	stripeService "api/internal/domains/payment/services/stripe"
	errLib "api/internal/libs/errors"
	responseHandlers "api/internal/libs/responses"

	"github.com/go-chi/chi"
)

type SubscriptionHandlers struct {
	StripeService *stripeService.SubscriptionService
}

func NewSubscriptionHandlers(container *di.Container) *SubscriptionHandlers {
	return &SubscriptionHandlers{
		StripeService: stripeService.NewSubscriptionService(container),
	}
}

// GetSubscription retrieves a subscription by ID with security validation
// @Description Get subscription details with ownership verification
// @Tags subscriptions
// @Accept json
// @Produce json
// @Param id path string true "Subscription ID"
// @Success 200 {object} map[string]interface{} "Subscription details"
// @Failure 400 {object} map[string]interface{} "Bad Request: Invalid subscription ID"
// @Failure 403 {object} map[string]interface{} "Forbidden: Access denied"
// @Failure 404 {object} map[string]interface{} "Not Found: Subscription not found"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Security Bearer
// @Router /subscriptions/{id} [get]
func (h *SubscriptionHandlers) GetSubscription(w http.ResponseWriter, r *http.Request) {
	subscriptionID := strings.TrimSpace(chi.URLParam(r, "id"))
	if subscriptionID == "" {
		responseHandlers.RespondWithError(w, errLib.New("subscription ID is required", http.StatusBadRequest))
		return
	}

	subscription, err := h.StripeService.GetSubscription(r.Context(), subscriptionID)
	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	// Convert Stripe subscription to response format
	response := map[string]interface{}{
		"id":                subscription.ID,
		"status":            subscription.Status,
		"current_period_start": subscription.CurrentPeriodStart,
		"current_period_end":   subscription.CurrentPeriodEnd,
		"cancel_at_period_end": subscription.CancelAtPeriodEnd,
		"canceled_at":          subscription.CanceledAt,
		"created":              subscription.Created,
		"items":                subscription.Items,
		"latest_invoice":       subscription.LatestInvoice,
	}

	responseHandlers.RespondWithSuccess(w, response, http.StatusOK)
}

// CancelSubscription cancels a subscription
// @Description Cancel a subscription immediately or at period end
// @Tags subscriptions
// @Accept json
// @Produce json
// @Param id path string true "Subscription ID"
// @Param immediate query boolean false "Cancel immediately (default: false - cancel at period end)"
// @Success 200 {object} map[string]interface{} "Cancellation successful"
// @Failure 400 {object} map[string]interface{} "Bad Request: Invalid input"
// @Failure 403 {object} map[string]interface{} "Forbidden: Access denied"
// @Failure 404 {object} map[string]interface{} "Not Found: Subscription not found"
// @Failure 409 {object} map[string]interface{} "Conflict: Subscription already cancelled"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Security Bearer
// @Router /subscriptions/{id}/cancel [post]
func (h *SubscriptionHandlers) CancelSubscription(w http.ResponseWriter, r *http.Request) {
	subscriptionID := strings.TrimSpace(chi.URLParam(r, "id"))
	if subscriptionID == "" {
		responseHandlers.RespondWithError(w, errLib.New("subscription ID is required", http.StatusBadRequest))
		return
	}

	// Parse immediate cancellation parameter
	immediate := r.URL.Query().Get("immediate") == "true"

	cancelledSub, err := h.StripeService.CancelSubscription(r.Context(), subscriptionID, immediate)
	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	response := map[string]interface{}{
		"id":                  cancelledSub.ID,
		"status":              cancelledSub.Status,
		"canceled_at":         cancelledSub.CanceledAt,
		"cancel_at_period_end": cancelledSub.CancelAtPeriodEnd,
		"current_period_end":   cancelledSub.CurrentPeriodEnd,
		"message": func() string {
			if immediate {
				return "Subscription cancelled immediately"
			}
			return "Subscription will cancel at the end of the current billing period"
		}(),
	}

	responseHandlers.RespondWithSuccess(w, response, http.StatusOK)
}

// PauseSubscription pauses a subscription
// @Description Pause a subscription with optional resume date
// @Tags subscriptions
// @Accept json
// @Produce json
// @Param id path string true "Subscription ID"
// @Param resume_at query string false "Resume date (RFC3339 format)"
// @Success 200 {object} map[string]interface{} "Pause successful"
// @Failure 400 {object} map[string]interface{} "Bad Request: Invalid input"
// @Failure 403 {object} map[string]interface{} "Forbidden: Access denied"
// @Failure 404 {object} map[string]interface{} "Not Found: Subscription not found"
// @Failure 409 {object} map[string]interface{} "Conflict: Only active subscriptions can be paused"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Security Bearer
// @Router /subscriptions/{id}/pause [post]
func (h *SubscriptionHandlers) PauseSubscription(w http.ResponseWriter, r *http.Request) {
	subscriptionID := strings.TrimSpace(chi.URLParam(r, "id"))
	if subscriptionID == "" {
		responseHandlers.RespondWithError(w, errLib.New("subscription ID is required", http.StatusBadRequest))
		return
	}

	// Parse resume date if provided
	var resumeAt *time.Time
	if resumeDateStr := r.URL.Query().Get("resume_at"); resumeDateStr != "" {
		parsedTime, parseErr := time.Parse(time.RFC3339, resumeDateStr)
		if parseErr != nil {
			responseHandlers.RespondWithError(w, errLib.New("invalid resume_at date format (use RFC3339)", http.StatusBadRequest))
			return
		}
		resumeAt = &parsedTime
	}

	pausedSub, err := h.StripeService.PauseSubscription(r.Context(), subscriptionID, resumeAt)
	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	response := map[string]interface{}{
		"id":               pausedSub.ID,
		"status":           pausedSub.Status,
		"pause_collection": pausedSub.PauseCollection,
		"message":          "Subscription paused successfully",
	}

	responseHandlers.RespondWithSuccess(w, response, http.StatusOK)
}

// ResumeSubscription resumes a paused subscription
// @Description Resume a paused subscription
// @Tags subscriptions
// @Accept json
// @Produce json
// @Param id path string true "Subscription ID"
// @Success 200 {object} map[string]interface{} "Resume successful"
// @Failure 400 {object} map[string]interface{} "Bad Request: Invalid subscription ID"
// @Failure 403 {object} map[string]interface{} "Forbidden: Access denied"
// @Failure 404 {object} map[string]interface{} "Not Found: Subscription not found"
// @Failure 409 {object} map[string]interface{} "Conflict: Subscription is not paused"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Security Bearer
// @Router /subscriptions/{id}/resume [post]
func (h *SubscriptionHandlers) ResumeSubscription(w http.ResponseWriter, r *http.Request) {
	subscriptionID := strings.TrimSpace(chi.URLParam(r, "id"))
	if subscriptionID == "" {
		responseHandlers.RespondWithError(w, errLib.New("subscription ID is required", http.StatusBadRequest))
		return
	}

	resumedSub, err := h.StripeService.ResumeSubscription(r.Context(), subscriptionID)
	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	response := map[string]interface{}{
		"id":               resumedSub.ID,
		"status":           resumedSub.Status,
		"pause_collection": resumedSub.PauseCollection,
		"message":          "Subscription resumed successfully",
	}

	responseHandlers.RespondWithSuccess(w, response, http.StatusOK)
}

// GetCustomerSubscriptions retrieves all subscriptions for the authenticated customer
// @Description Get all subscriptions for the authenticated customer
// @Tags subscriptions
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{} "Customer subscriptions"
// @Failure 403 {object} map[string]interface{} "Forbidden: Access denied"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Security Bearer
// @Router /subscriptions [get]
func (h *SubscriptionHandlers) GetCustomerSubscriptions(w http.ResponseWriter, r *http.Request) {
	subscriptions, err := h.StripeService.GetCustomerSubscriptions(r.Context())
	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	// Convert subscriptions to response format
	var subscriptionList []map[string]interface{}
	for _, sub := range subscriptions {
		subscriptionList = append(subscriptionList, map[string]interface{}{
			"id":                   sub.ID,
			"status":               sub.Status,
			"current_period_start": sub.CurrentPeriodStart,
			"current_period_end":   sub.CurrentPeriodEnd,
			"cancel_at_period_end": sub.CancelAtPeriodEnd,
			"canceled_at":          sub.CanceledAt,
			"created":              sub.Created,
			"items":                sub.Items,
			"latest_invoice":       sub.LatestInvoice,
		})
	}

	response := map[string]interface{}{
		"data":  subscriptionList,
		"count": len(subscriptionList),
	}

	responseHandlers.RespondWithSuccess(w, response, http.StatusOK)
}

// CreatePortalSession creates a secure customer portal session
// @Description Create a Stripe Customer Portal session for subscription management
// @Tags subscriptions
// @Accept json
// @Produce json
// @Param return_url query string true "Return URL after portal session"
// @Success 200 {object} map[string]interface{} "Portal session URL"
// @Failure 400 {object} map[string]interface{} "Bad Request: Invalid return URL"
// @Failure 403 {object} map[string]interface{} "Forbidden: Access denied"
// @Failure 404 {object} map[string]interface{} "Not Found: Customer not found"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Security Bearer
// @Router /subscriptions/portal [post]
func (h *SubscriptionHandlers) CreatePortalSession(w http.ResponseWriter, r *http.Request) {
	returnURL := strings.TrimSpace(r.URL.Query().Get("return_url"))
	if returnURL == "" {
		responseHandlers.RespondWithError(w, errLib.New("return_url parameter is required", http.StatusBadRequest))
		return
	}

	// Validate URL format
	if !strings.HasPrefix(returnURL, "http://") && !strings.HasPrefix(returnURL, "https://") {
		responseHandlers.RespondWithError(w, errLib.New("return_url must be a valid HTTP/HTTPS URL", http.StatusBadRequest))
		return
	}

	portalURL, err := h.StripeService.CreateCustomerPortalSession(r.Context(), returnURL)
	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	response := map[string]interface{}{
		"portal_url": portalURL,
		"message":    "Portal session created successfully",
	}

	responseHandlers.RespondWithSuccess(w, response, http.StatusOK)
}