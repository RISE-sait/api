package user

import (
	"net/http"
	"strconv"

	"api/internal/di"
	"api/internal/domains/user/services"
	errLib "api/internal/libs/errors"
	responseHandlers "api/internal/libs/responses"
	"api/internal/libs/validators"
	contextUtils "api/utils/context"

	"github.com/go-chi/chi"
	"github.com/google/uuid"
)

type CreditHandler struct {
	CreditService *services.CustomerCreditService
}

func NewCreditHandler(container *di.Container) *CreditHandler {
	return &CreditHandler{
		CreditService: services.NewCustomerCreditService(container),
	}
}

// GetCustomerCredits retrieves the current credit balance for the authenticated user
// @Tags credits
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{} "Credit balance retrieved successfully"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Security Bearer
// @Router /secure/credits [get]
func (h *CreditHandler) GetCustomerCredits(w http.ResponseWriter, r *http.Request) {
	customerID, err := contextUtils.GetUserID(r.Context())
	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	if credits, err := h.CreditService.GetCustomerCredits(r.Context(), customerID); err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	} else {
		response := map[string]interface{}{
			"customer_id": customerID,
			"credits":     credits,
		}
		responseHandlers.RespondWithSuccess(w, response, http.StatusOK)
	}
}

// GetCustomerCreditTransactions retrieves credit transaction history for the authenticated user
// @Tags credits
// @Accept json
// @Produce json
// @Param limit query int false "Number of items per page" minimum(1) maximum(100) default(20)
// @Param offset query int false "Number of items to skip" minimum(0) default(0)
// @Success 200 {object} map[string]interface{} "Credit transaction history retrieved successfully"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Security Bearer
// @Router /secure/credits/transactions [get]
func (h *CreditHandler) GetCustomerCreditTransactions(w http.ResponseWriter, r *http.Request) {
	customerID, err := contextUtils.GetUserID(r.Context())
	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	// Parse pagination parameters
	limit := 20
	offset := 0

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if parsed, parseErr := strconv.Atoi(limitStr); parseErr == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}

	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if parsed, parseErr := strconv.Atoi(offsetStr); parseErr == nil && parsed >= 0 {
			offset = parsed
		}
	}

	if transactions, err := h.CreditService.GetCustomerCreditTransactions(r.Context(), customerID, int32(limit), int32(offset)); err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	} else {
		response := map[string]interface{}{
			"customer_id":  customerID,
			"transactions": transactions,
			"limit":        limit,
			"offset":       offset,
		}
		responseHandlers.RespondWithSuccess(w, response, http.StatusOK)
	}
}

// AddCustomerCredits adds credits to a customer's account (admin only)
// @Tags credits
// @Accept json
// @Produce json
// @Param id path string true "Customer ID" format(uuid)
// @Param request body map[string]interface{} true "Credit addition request" example({"amount":100,"description":"Bonus credits"})
// @Success 200 {object} map[string]interface{} "Credits added successfully"
// @Failure 400 {object} map[string]interface{} "Bad Request: Invalid input"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Forbidden: Admin access required"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Security Bearer
// @Router /admin/customers/{id}/credits/add [post]
func (h *CreditHandler) AddCustomerCredits(w http.ResponseWriter, r *http.Request) {
	// Parse customer ID from URL
	var customerID uuid.UUID
	if idStr := chi.URLParam(r, "id"); idStr != "" {
		if id, err := validators.ParseUUID(idStr); err != nil {
			responseHandlers.RespondWithError(w, err)
			return
		} else {
			customerID = id
		}
	} else {
		responseHandlers.RespondWithError(w, errLib.New("customer id must be provided", http.StatusBadRequest))
		return
	}

	// Parse request body
	var requestBody struct {
		Amount      int32  `json:"amount" validate:"required,min=1"`
		Description string `json:"description" validate:"required"`
	}

	if err := validators.ParseJSON(r.Body, &requestBody); err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	if err := validators.ValidateDto(&requestBody); err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	if err := h.CreditService.AddCredits(r.Context(), customerID, requestBody.Amount, requestBody.Description); err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	} else {
		response := map[string]interface{}{
			"message":     "Credits added successfully",
			"customer_id": customerID,
			"amount":      requestBody.Amount,
		}
		responseHandlers.RespondWithSuccess(w, response, http.StatusOK)
	}
}

// DeductCustomerCredits removes credits from a customer's account (admin only)
// @Tags credits
// @Accept json
// @Produce json
// @Param id path string true "Customer ID" format(uuid)
// @Param request body map[string]interface{} true "Credit deduction request" example({"amount":50,"description":"Penalty deduction"})
// @Success 200 {object} map[string]interface{} "Credits deducted successfully"
// @Failure 400 {object} map[string]interface{} "Bad Request: Invalid input or insufficient credits"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Forbidden: Admin access required"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Security Bearer
// @Router /admin/customers/{id}/credits/deduct [post]
func (h *CreditHandler) DeductCustomerCredits(w http.ResponseWriter, r *http.Request) {
	// Parse customer ID from URL
	var customerID uuid.UUID
	if idStr := chi.URLParam(r, "id"); idStr != "" {
		if id, err := validators.ParseUUID(idStr); err != nil {
			responseHandlers.RespondWithError(w, err)
			return
		} else {
			customerID = id
		}
	} else {
		responseHandlers.RespondWithError(w, errLib.New("customer id must be provided", http.StatusBadRequest))
		return
	}

	// Parse request body
	var requestBody struct {
		Amount      int32  `json:"amount" validate:"required,min=1"`
		Description string `json:"description" validate:"required"`
	}

	if err := validators.ParseJSON(r.Body, &requestBody); err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	if err := validators.ValidateDto(&requestBody); err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	if err := h.CreditService.DeductCredits(r.Context(), customerID, requestBody.Amount, requestBody.Description); err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	} else {
		response := map[string]interface{}{
			"message":     "Credits deducted successfully",
			"customer_id": customerID,
			"amount":      requestBody.Amount,
		}
		responseHandlers.RespondWithSuccess(w, response, http.StatusOK)
	}
}

// GetEventCreditTransactions retrieves all credit transactions for a specific event (admin only)
// @Tags credits
// @Accept json
// @Produce json
// @Param id path string true "Event ID" format(uuid)
// @Success 200 {object} map[string]interface{} "Event credit transactions retrieved successfully"
// @Failure 400 {object} map[string]interface{} "Bad Request: Invalid event ID"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Forbidden: Admin access required"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Security Bearer
// @Router /admin/events/{id}/credit-transactions [get]
func (h *CreditHandler) GetEventCreditTransactions(w http.ResponseWriter, r *http.Request) {
	// Parse event ID from URL
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

	if transactions, err := h.CreditService.GetEventCreditTransactions(r.Context(), eventID); err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	} else {
		response := map[string]interface{}{
			"event_id":     eventID,
			"transactions": transactions,
		}
		responseHandlers.RespondWithSuccess(w, response, http.StatusOK)
	}
}

// UpdateEventCreditCost updates the credit cost for an event (admin only)
// @Tags credits
// @Accept json
// @Produce json
// @Param id path string true "Event ID" format(uuid)
// @Param request body map[string]interface{} true "Credit cost update request" example({"credit_cost":25})
// @Success 200 {object} map[string]interface{} "Event credit cost updated successfully"
// @Failure 400 {object} map[string]interface{} "Bad Request: Invalid input"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Forbidden: Admin access required"
// @Failure 404 {object} map[string]interface{} "Not Found: Event not found"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Security Bearer
// @Router /admin/events/{id}/credit-cost [put]
func (h *CreditHandler) UpdateEventCreditCost(w http.ResponseWriter, r *http.Request) {
	// Parse event ID from URL
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

	// Parse request body
	var requestBody struct {
		CreditCost *int32 `json:"credit_cost" validate:"omitempty,min=0"`
	}

	if err := validators.ParseJSON(r.Body, &requestBody); err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	if err := validators.ValidateDto(&requestBody); err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	if err := h.CreditService.UpdateEventCreditCost(r.Context(), eventID, requestBody.CreditCost); err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	} else {
		response := map[string]interface{}{
			"message":     "Event credit cost updated successfully",
			"event_id":    eventID,
			"credit_cost": requestBody.CreditCost,
		}
		responseHandlers.RespondWithSuccess(w, response, http.StatusOK)
	}
}