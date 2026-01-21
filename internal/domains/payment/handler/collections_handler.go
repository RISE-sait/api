package payment

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"api/internal/di"
	service "api/internal/domains/payment/services"
	errLib "api/internal/libs/errors"
	responses "api/internal/libs/responses"
	contextUtils "api/utils/context"

	"github.com/go-chi/chi"
	"github.com/google/uuid"
)

type CollectionsHandler struct {
	service *service.CollectionsService
}

func NewCollectionsHandler(container *di.Container) *CollectionsHandler {
	return &CollectionsHandler{
		service: service.NewCollectionsService(container),
	}
}

// GetCustomerBalance returns the customer's past due balance
// @Summary Get customer balance
// @Description Get customer's past due amount and open invoices from Stripe
// @Tags Collections
// @Accept json
// @Produce json
// @Param customer_id path string true "Customer ID"
// @Success 200 {object} service.CustomerBalance "Customer balance"
// @Failure 404 {object} map[string]string "Customer not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Security Bearer
// @Router /admin/collections/customers/{customer_id}/balance [get]
func (h *CollectionsHandler) GetCustomerBalance(w http.ResponseWriter, r *http.Request) {
	customerIDStr := chi.URLParam(r, "customer_id")
	customerID, err := uuid.Parse(customerIDStr)
	if err != nil {
		responses.RespondWithError(w, errLib.New("Invalid customer ID", http.StatusBadRequest))
		return
	}

	balance, svcErr := h.service.GetCustomerBalance(r.Context(), customerID)
	if svcErr != nil {
		responses.RespondWithError(w, svcErr)
		return
	}

	responses.RespondWithSuccess(w, balance, http.StatusOK)
}

// GetCustomerPaymentMethods returns saved payment methods for a customer
// @Summary Get customer payment methods
// @Description Get saved payment methods (cards) for a customer from Stripe
// @Tags Collections
// @Accept json
// @Produce json
// @Param customer_id path string true "Customer ID"
// @Success 200 {object} map[string]interface{} "Payment methods retrieved"
// @Failure 404 {object} map[string]string "Customer not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Security Bearer
// @Router /admin/collections/customers/{customer_id}/payment-methods [get]
func (h *CollectionsHandler) GetCustomerPaymentMethods(w http.ResponseWriter, r *http.Request) {
	customerIDStr := chi.URLParam(r, "customer_id")
	customerID, err := uuid.Parse(customerIDStr)
	if err != nil {
		responses.RespondWithError(w, errLib.New("Invalid customer ID", http.StatusBadRequest))
		return
	}

	methods, svcErr := h.service.GetCustomerPaymentMethods(r.Context(), customerID)
	if svcErr != nil {
		responses.RespondWithError(w, svcErr)
		return
	}

	responses.RespondWithSuccess(w, map[string]interface{}{
		"payment_methods": methods,
	}, http.StatusOK)
}

// ChargeCard charges a saved card on file
// @Summary Charge saved card
// @Description Charge a saved payment method for a customer
// @Tags Collections
// @Accept json
// @Produce json
// @Param body body service.ChargeCardRequest true "Charge card request"
// @Success 200 {object} service.CollectionResult "Charge result"
// @Failure 400 {object} map[string]string "Bad request"
// @Failure 404 {object} map[string]string "Customer not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Security Bearer
// @Router /admin/collections/charge-card [post]
func (h *CollectionsHandler) ChargeCard(w http.ResponseWriter, r *http.Request) {
	adminID, err := contextUtils.GetUserID(r.Context())
	if err != nil {
		responses.RespondWithError(w, errLib.New("Unauthorized", http.StatusUnauthorized))
		return
	}

	var req service.ChargeCardRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		responses.RespondWithError(w, errLib.New("Invalid request body", http.StatusBadRequest))
		return
	}

	if req.CustomerID == uuid.Nil {
		responses.RespondWithError(w, errLib.New("customer_id is required", http.StatusBadRequest))
		return
	}
	if req.PaymentMethodID == "" {
		responses.RespondWithError(w, errLib.New("payment_method_id is required", http.StatusBadRequest))
		return
	}
	if req.Amount <= 0 {
		responses.RespondWithError(w, errLib.New("amount must be greater than 0", http.StatusBadRequest))
		return
	}

	result, svcErr := h.service.ChargeCard(r.Context(), adminID, req)
	if svcErr != nil {
		responses.RespondWithError(w, svcErr)
		return
	}

	log.Printf("[COLLECTIONS-HANDLER] Admin %s charged card for customer %s: success=%v, amount=$%.2f",
		adminID, req.CustomerID, result.Success, req.Amount)

	responses.RespondWithSuccess(w, result, http.StatusOK)
}

// SendPaymentLink creates and sends a payment link
// @Summary Send payment link
// @Description Create a payment link and optionally send it via email
// @Tags Collections
// @Accept json
// @Produce json
// @Param body body service.SendPaymentLinkRequest true "Payment link request"
// @Success 200 {object} service.CollectionResult "Payment link created"
// @Failure 400 {object} map[string]string "Bad request"
// @Failure 404 {object} map[string]string "Customer not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Security Bearer
// @Router /admin/collections/send-payment-link [post]
func (h *CollectionsHandler) SendPaymentLink(w http.ResponseWriter, r *http.Request) {
	adminID, err := contextUtils.GetUserID(r.Context())
	if err != nil {
		responses.RespondWithError(w, errLib.New("Unauthorized", http.StatusUnauthorized))
		return
	}

	var req service.SendPaymentLinkRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		responses.RespondWithError(w, errLib.New("Invalid request body", http.StatusBadRequest))
		return
	}

	if req.CustomerID == uuid.Nil {
		responses.RespondWithError(w, errLib.New("customer_id is required", http.StatusBadRequest))
		return
	}
	if req.Amount <= 0 {
		responses.RespondWithError(w, errLib.New("amount must be greater than 0", http.StatusBadRequest))
		return
	}

	result, svcErr := h.service.SendPaymentLink(r.Context(), adminID, req)
	if svcErr != nil {
		responses.RespondWithError(w, svcErr)
		return
	}

	log.Printf("[COLLECTIONS-HANDLER] Admin %s created payment link for customer %s: $%.2f, send_email=%v",
		adminID, req.CustomerID, req.Amount, req.SendEmail)

	responses.RespondWithSuccess(w, result, http.StatusOK)
}

// RecordManualPayment records a manual payment entry
// @Summary Record manual payment
// @Description Record a manual payment (cash, check, etc.) for a customer
// @Tags Collections
// @Accept json
// @Produce json
// @Param body body service.RecordManualPaymentRequest true "Manual payment request"
// @Success 200 {object} service.CollectionResult "Payment recorded"
// @Failure 400 {object} map[string]string "Bad request"
// @Failure 404 {object} map[string]string "Customer not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Security Bearer
// @Router /admin/collections/record-manual [post]
func (h *CollectionsHandler) RecordManualPayment(w http.ResponseWriter, r *http.Request) {
	adminID, err := contextUtils.GetUserID(r.Context())
	if err != nil {
		responses.RespondWithError(w, errLib.New("Unauthorized", http.StatusUnauthorized))
		return
	}

	var req service.RecordManualPaymentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		responses.RespondWithError(w, errLib.New("Invalid request body", http.StatusBadRequest))
		return
	}

	if req.CustomerID == uuid.Nil {
		responses.RespondWithError(w, errLib.New("customer_id is required", http.StatusBadRequest))
		return
	}
	if req.Amount <= 0 {
		responses.RespondWithError(w, errLib.New("amount must be greater than 0", http.StatusBadRequest))
		return
	}
	if req.PaymentMethod == "" {
		responses.RespondWithError(w, errLib.New("payment_method is required (e.g., 'cash', 'check')", http.StatusBadRequest))
		return
	}

	result, svcErr := h.service.RecordManualPayment(r.Context(), adminID, req)
	if svcErr != nil {
		responses.RespondWithError(w, svcErr)
		return
	}

	log.Printf("[COLLECTIONS-HANDLER] Admin %s recorded manual payment for customer %s: $%.2f (%s)",
		adminID, req.CustomerID, req.Amount, req.PaymentMethod)

	responses.RespondWithSuccess(w, result, http.StatusOK)
}

// GetCollectionAttempts returns collection attempt history
// @Summary List collection attempts
// @Description Get collection attempts with optional filters
// @Tags Collections
// @Accept json
// @Produce json
// @Param customer_id query string false "Filter by customer ID"
// @Param status query string false "Filter by status (pending, success, failed, disputed)"
// @Param method query string false "Filter by method (card_charge, payment_link, manual_entry)"
// @Param start_date query string false "Start date (RFC3339)"
// @Param end_date query string false "End date (RFC3339)"
// @Param limit query int false "Page size" default(50)
// @Param offset query int false "Page offset" default(0)
// @Success 200 {object} map[string]interface{} "Collection attempts"
// @Failure 400 {object} map[string]string "Bad request"
// @Failure 500 {object} map[string]string "Internal server error"
// @Security Bearer
// @Router /admin/collections/attempts [get]
func (h *CollectionsHandler) GetCollectionAttempts(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	customerIDStr := r.URL.Query().Get("customer_id")
	status := r.URL.Query().Get("status")
	method := r.URL.Query().Get("method")
	startDateStr := r.URL.Query().Get("start_date")
	endDateStr := r.URL.Query().Get("end_date")
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	// Parse optional customer ID
	var customerID *uuid.UUID
	if customerIDStr != "" {
		if parsed, err := uuid.Parse(customerIDStr); err == nil {
			customerID = &parsed
		} else {
			responses.RespondWithError(w, errLib.New("Invalid customer_id format", http.StatusBadRequest))
			return
		}
	}

	// Parse dates
	var startDate, endDate *time.Time
	if startDateStr != "" {
		if parsed, err := time.Parse(time.RFC3339, startDateStr); err == nil {
			startDate = &parsed
		} else {
			responses.RespondWithError(w, errLib.New("Invalid start_date format (use RFC3339)", http.StatusBadRequest))
			return
		}
	}
	if endDateStr != "" {
		if parsed, err := time.Parse(time.RFC3339, endDateStr); err == nil {
			endDate = &parsed
		} else {
			responses.RespondWithError(w, errLib.New("Invalid end_date format (use RFC3339)", http.StatusBadRequest))
			return
		}
	}

	// Parse pagination
	limit := int32(50)
	offset := int32(0)
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 1000 {
			limit = int32(l)
		}
	}
	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = int32(o)
		}
	}

	attempts, total, err := h.service.GetCollectionAttempts(r.Context(), customerID, nil, status, method, startDate, endDate, limit, offset)
	if err != nil {
		log.Printf("[COLLECTIONS-HANDLER] Error fetching collection attempts: %v", err)
		responses.RespondWithError(w, errLib.New("Failed to fetch collection attempts", http.StatusInternalServerError))
		return
	}

	responses.RespondWithSuccess(w, map[string]interface{}{
		"attempts": attempts,
		"total":    total,
		"limit":    limit,
		"offset":   offset,
	}, http.StatusOK)
}
