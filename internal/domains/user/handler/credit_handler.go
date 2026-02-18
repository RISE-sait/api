package user

import (
	"net/http"
	"strconv"

	"api/internal/di"
	familyService "api/internal/domains/family/service"
	"api/internal/domains/user/services"
	errLib "api/internal/libs/errors"
	responseHandlers "api/internal/libs/responses"
	"api/internal/libs/validators"
	contextUtils "api/utils/context"

	"github.com/go-chi/chi"
	"github.com/google/uuid"
)

type CreditHandler struct {
	CreditService       *services.CustomerCreditService
	WeeklyCreditService *services.CreditService
	FamilyService       *familyService.Service
}

func NewCreditHandler(container *di.Container) *CreditHandler {
	return &CreditHandler{
		CreditService:       services.NewCustomerCreditService(container),
		WeeklyCreditService: services.NewCreditService(container),
		FamilyService:       familyService.NewService(container),
	}
}

// GetCustomerCredits retrieves the current credit balance for the authenticated user
// Parents can view their child's credits by passing the child_id query parameter.
// @Tags credits
// @Accept json
// @Produce json
// @Param child_id query string false "Child user ID (for parent viewing child's credits)" format(uuid)
// @Success 200 {object} map[string]interface{} "Credit balance retrieved successfully"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Forbidden: Not authorized to view child's credits"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Security Bearer
// @Router /secure/credits [get]
func (h *CreditHandler) GetCustomerCredits(w http.ResponseWriter, r *http.Request) {
	customerID, err := contextUtils.GetUserID(r.Context())
	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	// Check if parent is requesting child's credits
	targetCustomerID := customerID
	if childIDStr := r.URL.Query().Get("child_id"); childIDStr != "" {
		childID, parseErr := validators.ParseUUID(childIDStr)
		if parseErr != nil {
			responseHandlers.RespondWithError(w, parseErr)
			return
		}

		// Verify parent has access to this child
		if verifyErr := h.FamilyService.VerifyParentChildAccess(r.Context(), customerID, childID); verifyErr != nil {
			responseHandlers.RespondWithError(w, verifyErr)
			return
		}

		targetCustomerID = childID
	}

	if credits, err := h.CreditService.GetCustomerCredits(r.Context(), targetCustomerID); err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	} else {
		response := map[string]interface{}{
			"customer_id": targetCustomerID,
			"credits":     credits,
		}
		responseHandlers.RespondWithSuccess(w, response, http.StatusOK)
	}
}

// GetCustomerCreditTransactions retrieves credit transaction history for the authenticated user
// Parents can view their child's transactions by passing the child_id query parameter.
// @Tags credits
// @Accept json
// @Produce json
// @Param limit query int false "Number of items per page" minimum(1) maximum(100) default(20)
// @Param offset query int false "Number of items to skip" minimum(0) default(0)
// @Param child_id query string false "Child user ID (for parent viewing child's transactions)" format(uuid)
// @Success 200 {object} map[string]interface{} "Credit transaction history retrieved successfully"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Forbidden: Not authorized to view child's transactions"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Security Bearer
// @Router /secure/credits/transactions [get]
func (h *CreditHandler) GetCustomerCreditTransactions(w http.ResponseWriter, r *http.Request) {
	customerID, err := contextUtils.GetUserID(r.Context())
	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	// Check if parent is requesting child's transactions
	targetCustomerID := customerID
	if childIDStr := r.URL.Query().Get("child_id"); childIDStr != "" {
		childID, parseErr := validators.ParseUUID(childIDStr)
		if parseErr != nil {
			responseHandlers.RespondWithError(w, parseErr)
			return
		}

		// Verify parent has access to this child
		if verifyErr := h.FamilyService.VerifyParentChildAccess(r.Context(), customerID, childID); verifyErr != nil {
			responseHandlers.RespondWithError(w, verifyErr)
			return
		}

		targetCustomerID = childID
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

	if transactions, err := h.CreditService.GetCustomerCreditTransactions(r.Context(), targetCustomerID, int32(limit), int32(offset)); err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	} else {
		response := map[string]interface{}{
			"customer_id":  targetCustomerID,
			"transactions": transactions,
			"limit":        limit,
			"offset":       offset,
		}
		responseHandlers.RespondWithSuccess(w, response, http.StatusOK)
	}
}

// GetWeeklyUsage retrieves the current weekly credit usage and limits for the authenticated user
// Parents can view their child's weekly usage by passing the child_id query parameter.
// @Tags credits
// @Accept json
// @Produce json
// @Param child_id query string false "Child user ID (for parent viewing child's weekly usage)" format(uuid)
// @Success 200 {object} map[string]interface{} "Weekly usage retrieved successfully" example({"customer_id":"uuid","current_week_usage":1,"weekly_limit":2,"remaining_credits":1})
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Forbidden: Not authorized to view child's weekly usage"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Security Bearer
// @Router /secure/credits/weekly-usage [get]
func (h *CreditHandler) GetWeeklyUsage(w http.ResponseWriter, r *http.Request) {
	customerID, err := contextUtils.GetUserID(r.Context())
	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	// Check if parent is requesting child's weekly usage
	targetCustomerID := customerID
	if childIDStr := r.URL.Query().Get("child_id"); childIDStr != "" {
		childID, parseErr := validators.ParseUUID(childIDStr)
		if parseErr != nil {
			responseHandlers.RespondWithError(w, parseErr)
			return
		}

		// Verify parent has access to this child
		if verifyErr := h.FamilyService.VerifyParentChildAccess(r.Context(), customerID, childID); verifyErr != nil {
			responseHandlers.RespondWithError(w, verifyErr)
			return
		}

		targetCustomerID = childID
	}

	currentUsage, weeklyLimit, err := h.WeeklyCreditService.GetWeeklyUsage(r.Context(), targetCustomerID)
	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	response := map[string]interface{}{
		"customer_id":         targetCustomerID,
		"current_week_usage":  currentUsage,
		"weekly_limit":        weeklyLimit,
		"remaining_credits":   nil,
	}

	if weeklyLimit != nil {
		remaining := *weeklyLimit - currentUsage
		if remaining < 0 {
			remaining = 0
		}
		response["remaining_credits"] = remaining
	}

	responseHandlers.RespondWithSuccess(w, response, http.StatusOK)
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

// GetAnyCustomerCredits retrieves credit balance for any customer by ID (admin only)
// @Tags credits
// @Accept json
// @Produce json
// @Param id path string true "Customer ID" format(uuid)
// @Success 200 {object} map[string]interface{} "Credit balance retrieved successfully"
// @Failure 400 {object} map[string]interface{} "Bad Request: Invalid customer ID"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Forbidden: Admin access required"
// @Failure 404 {object} map[string]interface{} "Not Found: Customer not found"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Security Bearer
// @Router /admin/customers/{id}/credits [get]
func (h *CreditHandler) GetAnyCustomerCredits(w http.ResponseWriter, r *http.Request) {
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

// GetAnyCustomerCreditTransactions retrieves credit transaction history for any customer by ID (admin only)
// @Tags credits
// @Accept json
// @Produce json
// @Param id path string true "Customer ID" format(uuid)
// @Param limit query int false "Number of items per page" minimum(1) maximum(100) default(20)
// @Param offset query int false "Number of items to skip" minimum(0) default(0)
// @Success 200 {object} map[string]interface{} "Credit transaction history retrieved successfully"
// @Failure 400 {object} map[string]interface{} "Bad Request: Invalid customer ID"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Forbidden: Admin access required"
// @Failure 404 {object} map[string]interface{} "Not Found: Customer not found"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Security Bearer
// @Router /admin/customers/{id}/credits/transactions [get]
func (h *CreditHandler) GetAnyCustomerCreditTransactions(w http.ResponseWriter, r *http.Request) {
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

// GetAnyCustomerWeeklyUsage retrieves weekly credit usage for any customer by ID (admin only)
// @Tags credits
// @Accept json
// @Produce json
// @Param id path string true "Customer ID" format(uuid)
// @Success 200 {object} map[string]interface{} "Weekly usage retrieved successfully"
// @Failure 400 {object} map[string]interface{} "Bad Request: Invalid customer ID"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Forbidden: Admin access required"
// @Failure 404 {object} map[string]interface{} "Not Found: Customer not found"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Security Bearer
// @Router /admin/customers/{id}/credits/weekly-usage [get]
func (h *CreditHandler) GetAnyCustomerWeeklyUsage(w http.ResponseWriter, r *http.Request) {
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

	currentUsage, weeklyLimit, err := h.WeeklyCreditService.GetWeeklyUsage(r.Context(), customerID)
	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	response := map[string]interface{}{
		"customer_id":         customerID,
		"current_week_usage":  currentUsage,
		"weekly_limit":        weeklyLimit,
		"remaining_credits":   nil,
	}

	if weeklyLimit != nil {
		remaining := *weeklyLimit - currentUsage
		if remaining < 0 {
			remaining = 0
		}
		response["remaining_credits"] = remaining
	}

	responseHandlers.RespondWithSuccess(w, response, http.StatusOK)
}

// GetCreditRefundLogs retrieves credit refund audit logs (admin only)
// @Tags credits
// @Accept json
// @Produce json
// @Param customer_id query string false "Filter by customer ID" format(uuid)
// @Param event_id query string false "Filter by event ID" format(uuid)
// @Param limit query int false "Number of items per page" minimum(1) maximum(100) default(20)
// @Param offset query int false "Number of items to skip" minimum(0) default(0)
// @Success 200 {object} map[string]interface{} "Credit refund logs retrieved successfully"
// @Failure 400 {object} map[string]interface{} "Bad Request: Invalid input"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Forbidden: Admin access required"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Security Bearer
// @Router /admin/credit-refund-logs [get]
func (h *CreditHandler) GetCreditRefundLogs(w http.ResponseWriter, r *http.Request) {
	// Parse optional customer_id filter
	var customerID *uuid.UUID
	if customerIDStr := r.URL.Query().Get("customer_id"); customerIDStr != "" {
		if id, err := validators.ParseUUID(customerIDStr); err != nil {
			responseHandlers.RespondWithError(w, err)
			return
		} else {
			customerID = &id
		}
	}

	// Parse optional event_id filter
	var eventID *uuid.UUID
	if eventIDStr := r.URL.Query().Get("event_id"); eventIDStr != "" {
		if id, err := validators.ParseUUID(eventIDStr); err != nil {
			responseHandlers.RespondWithError(w, err)
			return
		} else {
			eventID = &id
		}
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

	logs, err := h.CreditService.GetCreditRefundLogs(r.Context(), customerID, eventID, int32(limit), int32(offset))
	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	// Transform logs to include formatted names
	formattedLogs := make([]map[string]interface{}, len(logs))
	for i, log := range logs {
		formattedLogs[i] = map[string]interface{}{
			"id":                log.ID,
			"customer_id":       log.CustomerID,
			"customer_name":     log.CustomerFirstName + " " + log.CustomerLastName,
			"event_id":          log.EventID,
			"event_name":        log.EventName.String,
			"event_start_at":    log.EventStartAt.Time,
			"program_name":      log.ProgramName.String,
			"location_name":     log.LocationName.String,
			"credits_refunded":  log.CreditsRefunded,
			"performed_by":      log.PerformedBy,
			"performed_by_name": log.StaffFirstName + " " + log.StaffLastName,
			"staff_role":        log.StaffRole.String,
			"reason":            log.Reason.String,
			"ip_address":        log.IpAddress.String,
			"created_at":        log.CreatedAt,
		}
	}

	response := map[string]interface{}{
		"logs":   formattedLogs,
		"limit":  limit,
		"offset": offset,
	}
	responseHandlers.RespondWithSuccess(w, response, http.StatusOK)
}
