package payment

import (
	"fmt"
	"net/http"

	"api/internal/di"
	dto "api/internal/domains/payment/dto"
	service "api/internal/domains/payment/services"
	"api/internal/domains/payment/services/stripe"
	errLib "api/internal/libs/errors"
	"api/internal/libs/logger"
	responseHandlers "api/internal/libs/responses"
	"api/internal/libs/validators"
	contextUtils "api/utils/context"

	"github.com/go-chi/chi"
	"github.com/google/uuid"
)

type CheckoutHandlers struct {
	Service             *service.Service
	VerificationService *service.CheckoutVerificationService
}

func NewCheckoutHandlers(container *di.Container) *CheckoutHandlers {
	return &CheckoutHandlers{
		Service:             service.NewPurchaseService(container),
		VerificationService: service.NewCheckoutVerificationService(container),
	}
}

// CheckoutMembership allows a customer to check out a membership plan.
// @Description Generates a payment link for purchasing a membership plan.
// @Tags payments
// @Accept json
// @Produce json
// @Param id path string true "Membership plan ID"
// @Param discount_code query string false "Discount code to apply"
// @Success 200 {object} dto.CheckoutResponseDto "Payment link generated successfully"
// @Failure 400 {object} map[string]interface{} "Bad Request: Invalid input"
// @Failure 500 {object} map[string]interface{} "Internal Server Error: Failed to process checkout"
// @Security Bearer
// @Router /checkout/membership_plans/{id} [post]
func (h *CheckoutHandlers) CheckoutMembership(w http.ResponseWriter, r *http.Request) {

	var membershipPlanId uuid.UUID

	if idStr := chi.URLParam(r, "id"); idStr != "" {
		if id, err := validators.ParseUUID(idStr); err != nil {
			responseHandlers.RespondWithError(w, err)
			return
		} else {
			membershipPlanId = id
		}
	} else {
		responseHandlers.RespondWithError(w, errLib.New("membership planID must be provided", http.StatusBadRequest))
		return
	}

	var responseDto dto.CheckoutResponseDto

		var discountCode *string
	if code := r.URL.Query().Get("discount_code"); code != "" {
		discountCode = &code
	}

	// Get success and cancel URLs based on request origin
	successURL, cancelURL := stripe.GetCheckoutURLs(r)

	if paymentLink, err := h.Service.CheckoutMembershipPlan(r.Context(), membershipPlanId, discountCode, successURL, cancelURL); err != nil {
		responseDto.PaymentURL = paymentLink
		responseHandlers.RespondWithError(w, err)
	} else {
		responseDto.PaymentURL = paymentLink
		responseHandlers.RespondWithSuccess(w, responseDto, http.StatusOK)
	}
}

// CheckoutProgram allows a customer to check out a program.
// @Tags payments
// @Accept json
// @Produce json
// @Param id path string true "Program ID" format(uuid)
// @Param discount_code query string false "Discount code to apply"
// @Success 200 {object} dto.CheckoutResponseDto "Payment link generated successfully"
// @Failure 400 {object} map[string]interface{} "Bad Request: Invalid input"
// @Failure 500 {object} map[string]interface{} "Internal Server Error: Failed to process checkout"
// @Security Bearer
// @Router /checkout/programs/{id} [post]
func (h *CheckoutHandlers) CheckoutProgram(w http.ResponseWriter, r *http.Request) {

	var programID uuid.UUID

	if idStr := chi.URLParam(r, "id"); idStr != "" {
		if id, err := validators.ParseUUID(idStr); err != nil {
			responseHandlers.RespondWithError(w, err)
			return
		} else {
			programID = id
		}
	} else {
		responseHandlers.RespondWithError(w, errLib.New("program id must be provided", http.StatusBadRequest))
		return
	}

	var responseDto dto.CheckoutResponseDto

	var discountCode *string
	if code := r.URL.Query().Get("discount_code"); code != "" {
		discountCode = &code
	}

	// Get success and cancel URLs based on request origin
	successURL, cancelURL := stripe.GetCheckoutURLs(r)

	if paymentLink, err := h.Service.CheckoutProgram(r.Context(), programID, discountCode, successURL, cancelURL); err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	} else {
		responseDto.PaymentURL = paymentLink
		responseHandlers.RespondWithSuccess(w, responseDto, http.StatusOK)
	}
}

// CheckoutEvent allows a customer to check out an event.
// @Tags payments
// @Accept json
// @Produce json
// @Param id path string true "Event ID" format(uuid)
// @Param discount_code query string false "Discount code to apply"
// @Success 200 {object} dto.CheckoutResponseDto "Payment link generated successfully"
// @Failure 400 {object} map[string]interface{} "Bad Request: Invalid input or missing event ID"
// @Failure 404 {object} map[string]interface{} "Not Found: Event not found"
// @Failure 409 {object} map[string]interface{} "Conflict: Event is full or already booked"
// @Failure 500 {object} map[string]interface{} "Internal Server Error: Failed to process checkout"
// @Security Bearer
// @Router /checkout/events/{id} [post]
func (h *CheckoutHandlers) CheckoutEvent(w http.ResponseWriter, r *http.Request) {

	var eventID uuid.UUID

	if idStr := chi.URLParam(r, "id"); idStr != "" {
		if id, err := validators.ParseUUID(idStr); err != nil {
			responseHandlers.RespondWithError(w, err)
			return
		} else {
			eventID = id
		}
	} else {
		responseHandlers.RespondWithError(w, errLib.New("event id must be provided", http.StatusBadRequest))
		return
	}

	var responseDto dto.CheckoutResponseDto

	var discountCode *string
	if code := r.URL.Query().Get("discount_code"); code != "" {
		discountCode = &code
	}

	// Get success and cancel URLs based on request origin
	successURL, cancelURL := stripe.GetCheckoutURLs(r)

	if paymentLink, err := h.Service.CheckoutEvent(r.Context(), eventID, discountCode, successURL, cancelURL); err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	} else {
		responseDto.PaymentURL = paymentLink
		responseHandlers.RespondWithSuccess(w, responseDto, http.StatusOK)
	}
}

// GetEventEnrollmentOptions returns available enrollment options for an event
// @Tags payments
// @Accept json
// @Produce json
// @Param id path string true "Event ID" format(uuid)
// @Success 200 {object} payment.EventEnrollmentOptions "Event enrollment options retrieved successfully"
// @Failure 400 {object} map[string]interface{} "Bad Request: Invalid event ID"
// @Failure 404 {object} map[string]interface{} "Not Found: Event not found"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Security Bearer
// @Router /checkout/events/{id}/options [get]
func (h *CheckoutHandlers) GetEventEnrollmentOptions(w http.ResponseWriter, r *http.Request) {
	var eventID uuid.UUID

	if idStr := chi.URLParam(r, "id"); idStr != "" {
		if id, err := validators.ParseUUID(idStr); err != nil {
			responseHandlers.RespondWithError(w, err)
			return
		} else {
			eventID = id
		}
	} else {
		responseHandlers.RespondWithError(w, errLib.New("event id must be provided", http.StatusBadRequest))
		return
	}

	if options, err := h.Service.CheckEventEnrollmentOptions(r.Context(), eventID); err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	} else {
		responseHandlers.RespondWithSuccess(w, options, http.StatusOK)
	}
}


// CheckoutEventEnhanced uses the enhanced checkout with membership validation and supports multiple payment methods
// @Tags payments
// @Accept json
// @Produce json
// @Param id path string true "Event ID" format(uuid)
// @Param discount_code query string false "Discount code to apply"
// @Param request body map[string]interface{} false "Payment method request" example({"payment_method":"stripe"}) enum("stripe","credits")
// @Success 200 {object} dto.CheckoutResponseDto "Payment link generated or free enrollment completed"
// @Failure 400 {object} map[string]interface{} "Bad Request: Invalid input or missing event ID"
// @Failure 404 {object} map[string]interface{} "Not Found: Event not found"
// @Failure 409 {object} map[string]interface{} "Conflict: Event is full or already booked"
// @Failure 500 {object} map[string]interface{} "Internal Server Error: Failed to process checkout"
// @Security Bearer
// @Router /checkout/events/{id}/enhanced [post]
func (h *CheckoutHandlers) CheckoutEventEnhanced(w http.ResponseWriter, r *http.Request) {
	var eventID uuid.UUID

	if idStr := chi.URLParam(r, "id"); idStr != "" {
		if id, err := validators.ParseUUID(idStr); err != nil {
			responseHandlers.RespondWithError(w, err)
			return
		} else {
			eventID = id
		}
	} else {
		responseHandlers.RespondWithError(w, errLib.New("event id must be provided", http.StatusBadRequest))
		return
	}

	// Parse request body for payment method (optional)
	var requestBody struct {
		PaymentMethod string `json:"payment_method"`
	}

	// Only parse body if content is present
	if r.ContentLength > 0 {
		if err := validators.ParseJSON(r.Body, &requestBody); err != nil {
			responseHandlers.RespondWithError(w, err)
			return
		}
	}

	// Default to stripe if no payment method specified
	if requestBody.PaymentMethod == "" {
		requestBody.PaymentMethod = "stripe"
	}

	// Validate payment method
	if requestBody.PaymentMethod != "stripe" && requestBody.PaymentMethod != "credits" {
		responseHandlers.RespondWithError(w, errLib.New("payment_method must be 'stripe' or 'credits'", http.StatusBadRequest))
		return
	}

	// Handle credit payment
	if requestBody.PaymentMethod == "credits" {
		if err := h.Service.CheckoutEventWithCredits(r.Context(), eventID); err != nil {
			responseHandlers.RespondWithError(w, err)
			return
		} else {
			response := map[string]interface{}{
				"message": "Event enrollment completed successfully using credits",
			}
			responseHandlers.RespondWithSuccess(w, response, http.StatusOK)
			return
		}
	}

	// Handle stripe payment (default)
	var discountCode *string
	if code := r.URL.Query().Get("discount_code"); code != "" {
		discountCode = &code
	}

	// Get success and cancel URLs based on request origin
	successURL, cancelURL := stripe.GetCheckoutURLs(r)

	if paymentLink, err := h.Service.CheckoutEventEnhanced(r.Context(), eventID, discountCode, successURL, cancelURL); err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	} else {
		var responseDto dto.CheckoutResponseDto
		if paymentLink == "" {
			// Free enrollment completed
			response := map[string]interface{}{
				"message": "Event enrollment completed successfully (free for your membership)",
			}
			responseHandlers.RespondWithSuccess(w, response, http.StatusOK)
		} else {
			// Stripe payment required
			responseDto.PaymentURL = paymentLink
			responseHandlers.RespondWithSuccess(w, responseDto, http.StatusOK)
		}
	}
}

// TestSlackAlert manually triggers a Slack alert for testing
func (h *CheckoutHandlers) TestSlackAlert(w http.ResponseWriter, r *http.Request) {
	// Use structured logger to trigger Slack alert with critical component
	testLogger := logger.WithComponent("checkout-service")
	testLogger.Error("TEST: Payment processing failed", fmt.Errorf("this is a test error to verify Slack integration timing"))

	response := map[string]interface{}{
		"message": "Test Slack alert sent - check your Slack channel with optimized timing!",
		"status": "sent",
	}
	responseHandlers.RespondWithSuccess(w, response, http.StatusOK)
}

// VerifyCheckoutSession verifies a checkout session and ensures enrollment is complete
// This is called by the frontend after redirect from Stripe checkout as a safety net
// @Tags payments
// @Accept json
// @Produce json
// @Param session_id path string true "Stripe Checkout Session ID"
// @Success 200 {object} service.VerificationResult "Verification result"
// @Failure 400 {object} map[string]interface{} "Bad Request: Invalid session ID"
// @Failure 403 {object} map[string]interface{} "Forbidden: Session doesn't belong to user"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Security Bearer
// @Router /checkout/verify/{session_id} [post]
func (h *CheckoutHandlers) VerifyCheckoutSession(w http.ResponseWriter, r *http.Request) {
	sessionID := chi.URLParam(r, "session_id")
	if sessionID == "" {
		responseHandlers.RespondWithError(w, errLib.New("session_id is required", http.StatusBadRequest))
		return
	}

	// Get user ID from context (set by JWT middleware)
	userID, err := contextUtils.GetUserID(r.Context())
	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	result, verifyErr := h.VerificationService.VerifyCheckoutSession(r.Context(), sessionID, userID)
	if verifyErr != nil {
		responseHandlers.RespondWithError(w, verifyErr)
		return
	}

	responseHandlers.RespondWithSuccess(w, result, http.StatusOK)
}