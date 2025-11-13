package handler

import (
	"net/http"
	"strconv"

	"api/internal/di"
	"api/internal/domains/subsidy/dto"
	"api/internal/domains/subsidy/service"
	errLib "api/internal/libs/errors"
	responses "api/internal/libs/responses"
	"api/internal/libs/validators"
	"api/internal/middlewares"
	contextUtils "api/utils/context"

	"github.com/go-chi/chi"
	"github.com/google/uuid"
)

type SubsidyHandler struct {
	Service *service.SubsidyService
}

func NewSubsidyHandler(container *di.Container) *SubsidyHandler {
	return &SubsidyHandler{
		Service: service.NewSubsidyService(container),
	}
}

// ===== ADMIN HANDLERS - PROVIDER MANAGEMENT =====

// CreateProvider creates a new subsidy provider
// @Summary Create subsidy provider
// @Description Create a new subsidy provider organization (admin only)
// @Tags Subsidies - Admin
// @Accept json
// @Produce json
// @Param provider body dto.CreateProviderRequest true "Provider details"
// @Success 201 {object} responses.SuccessResponse
// @Failure 400 {object} responses.ErrorResponse
// @Failure 401 {object} responses.ErrorResponse
// @Failure 403 {object} responses.ErrorResponse
// @Security Bearer
// @Router /subsidies/providers [post]
func (h *SubsidyHandler) CreateProvider(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateProviderRequest

	if err := validators.ParseJSON(r.Body, &req); err != nil {
		responses.RespondWithError(w, err)
		return
	}

	if err := validators.ValidateDto(&req); err != nil {
		responses.RespondWithError(w, err)
		return
	}

	provider, err := h.Service.CreateProvider(r.Context(), &req)
	if err != nil {
		responses.RespondWithError(w, err)
		return
	}

	responses.RespondWithSuccess(w, provider, http.StatusCreated)
}

// GetProvider retrieves a provider by ID
// @Summary Get subsidy provider
// @Description Get details of a specific subsidy provider (admin only)
// @Tags Subsidies - Admin
// @Produce json
// @Param id path string true "Provider ID"
// @Success 200 {object} responses.SuccessResponse
// @Failure 400 {object} responses.ErrorResponse
// @Failure 401 {object} responses.ErrorResponse
// @Failure 403 {object} responses.ErrorResponse
// @Failure 404 {object} responses.ErrorResponse
// @Security Bearer
// @Router /subsidies/providers/{id} [get]
func (h *SubsidyHandler) GetProvider(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		responses.RespondWithError(w, errLib.New("Invalid provider ID", http.StatusBadRequest))
		return
	}

	provider, serviceErr := h.Service.GetProvider(r.Context(), id)
	if serviceErr != nil {
		responses.RespondWithError(w, serviceErr)
		return
	}

	responses.RespondWithSuccess(w, provider, http.StatusOK)
}

// ListProviders lists all providers
// @Summary List subsidy providers
// @Description Get list of all subsidy providers with optional filters (admin only)
// @Tags Subsidies - Admin
// @Produce json
// @Param is_active query boolean false "Filter by active status"
// @Success 200 {object} responses.SuccessResponse
// @Failure 401 {object} responses.ErrorResponse
// @Failure 403 {object} responses.ErrorResponse
// @Security Bearer
// @Router /subsidies/providers [get]
func (h *SubsidyHandler) ListProviders(w http.ResponseWriter, r *http.Request) {
	var isActive *bool
	if isActiveStr := r.URL.Query().Get("is_active"); isActiveStr != "" {
		val, err := strconv.ParseBool(isActiveStr)
		if err == nil {
			isActive = &val
		}
	}

	providers, err := h.Service.ListProviders(r.Context(), isActive)
	if err != nil {
		responses.RespondWithError(w, err)
		return
	}

	responses.RespondWithSuccess(w, providers, http.StatusOK)
}

// GetProviderStats retrieves provider statistics
// @Summary Get provider statistics
// @Description Get usage statistics for a specific provider (admin only)
// @Tags Subsidies - Admin
// @Produce json
// @Param id path string true "Provider ID"
// @Success 200 {object} responses.SuccessResponse
// @Failure 400 {object} responses.ErrorResponse
// @Failure 401 {object} responses.ErrorResponse
// @Failure 403 {object} responses.ErrorResponse
// @Failure 404 {object} responses.ErrorResponse
// @Security Bearer
// @Router /subsidies/providers/{id}/stats [get]
func (h *SubsidyHandler) GetProviderStats(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		responses.RespondWithError(w, errLib.New("Invalid provider ID", http.StatusBadRequest))
		return
	}

	stats, serviceErr := h.Service.GetProviderStats(r.Context(), id)
	if serviceErr != nil {
		responses.RespondWithError(w, serviceErr)
		return
	}

	responses.RespondWithSuccess(w, stats, http.StatusOK)
}

// ===== ADMIN HANDLERS - SUBSIDY MANAGEMENT =====

// CreateSubsidy creates a new subsidy for a customer
// @Summary Create subsidy
// @Description Create a new subsidy for a customer (admin only)
// @Tags Subsidies - Admin
// @Accept json
// @Produce json
// @Param subsidy body dto.CreateSubsidyRequest true "Subsidy details"
// @Success 201 {object} responses.SuccessResponse
// @Failure 400 {object} responses.ErrorResponse
// @Failure 401 {object} responses.ErrorResponse
// @Failure 403 {object} responses.ErrorResponse
// @Security Bearer
// @Router /subsidies [post]
func (h *SubsidyHandler) CreateSubsidy(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateSubsidyRequest

	if err := validators.ParseJSON(r.Body, &req); err != nil {
		responses.RespondWithError(w, err)
		return
	}

	if err := validators.ValidateDto(&req); err != nil {
		responses.RespondWithError(w, err)
		return
	}

	// SECURITY: Get staff ID from context with authorization check
	staffID, ok := getStaffIDFromContext(w, r)
	if !ok {
		return // Error already sent by helper
	}

	// Get IP address for audit logging
	ipAddress := middlewares.GetRealIP(r)

	subsidy, err := h.Service.CreateSubsidy(r.Context(), &req, staffID, ipAddress)
	if err != nil {
		responses.RespondWithError(w, err)
		return
	}

	responses.RespondWithSuccess(w, subsidy, http.StatusCreated)
}

// GetSubsidy retrieves subsidy details
// @Summary Get subsidy
// @Description Get details of a specific subsidy (admin only)
// @Tags Subsidies - Admin
// @Produce json
// @Param id path string true "Subsidy ID"
// @Success 200 {object} responses.SuccessResponse
// @Failure 400 {object} responses.ErrorResponse
// @Failure 401 {object} responses.ErrorResponse
// @Failure 403 {object} responses.ErrorResponse
// @Failure 404 {object} responses.ErrorResponse
// @Security Bearer
// @Router /subsidies/{id} [get]
func (h *SubsidyHandler) GetSubsidy(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		responses.RespondWithError(w, errLib.New("Invalid subsidy ID", http.StatusBadRequest))
		return
	}

	subsidy, serviceErr := h.Service.GetSubsidy(r.Context(), id)
	if serviceErr != nil {
		responses.RespondWithError(w, serviceErr)
		return
	}

	responses.RespondWithSuccess(w, subsidy, http.StatusOK)
}

// ListSubsidies lists subsidies with filters
// @Summary List subsidies
// @Description Get list of subsidies with optional filters (admin only)
// @Tags Subsidies - Admin
// @Produce json
// @Param customer_id query string false "Filter by customer ID"
// @Param provider_id query string false "Filter by provider ID"
// @Param status query string false "Filter by status (pending, approved, active, depleted, expired, revoked)"
// @Param page query integer false "Page number" default(1)
// @Param limit query integer false "Items per page" default(50)
// @Success 200 {object} responses.SuccessResponse
// @Failure 401 {object} responses.ErrorResponse
// @Failure 403 {object} responses.ErrorResponse
// @Security Bearer
// @Router /subsidies [get]
func (h *SubsidyHandler) ListSubsidies(w http.ResponseWriter, r *http.Request) {
	filters := dto.SubsidyFilters{
		Page:  parseIntQuery(r, "page", 1),
		Limit: parseIntQuery(r, "limit", 50),
	}

	if customerIDStr := r.URL.Query().Get("customer_id"); customerIDStr != "" {
		customerID, err := uuid.Parse(customerIDStr)
		if err == nil {
			filters.CustomerID = &customerID
		}
	}

	if providerIDStr := r.URL.Query().Get("provider_id"); providerIDStr != "" {
		providerID, err := uuid.Parse(providerIDStr)
		if err == nil {
			filters.ProviderID = &providerID
		}
	}

	if status := r.URL.Query().Get("status"); status != "" {
		filters.Status = &status
	}

	result, err := h.Service.ListSubsidies(r.Context(), filters)
	if err != nil {
		responses.RespondWithError(w, err)
		return
	}

	responses.RespondWithSuccess(w, result, http.StatusOK)
}

// DeactivateSubsidy deactivates a subsidy
// @Summary Deactivate subsidy
// @Description Deactivate/revoke a subsidy (admin only)
// @Tags Subsidies - Admin
// @Accept json
// @Produce json
// @Param id path string true "Subsidy ID"
// @Param request body dto.DeactivateSubsidyRequest true "Deactivation reason"
// @Success 200 {object} responses.SuccessResponse
// @Failure 400 {object} responses.ErrorResponse
// @Failure 401 {object} responses.ErrorResponse
// @Failure 403 {object} responses.ErrorResponse
// @Failure 404 {object} responses.ErrorResponse
// @Security Bearer
// @Router /subsidies/{id}/deactivate [post]
func (h *SubsidyHandler) DeactivateSubsidy(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		responses.RespondWithError(w, errLib.New("Invalid subsidy ID", http.StatusBadRequest))
		return
	}

	var req dto.DeactivateSubsidyRequest
	if err := validators.ParseJSON(r.Body, &req); err != nil {
		responses.RespondWithError(w, err)
		return
	}

	if err := validators.ValidateDto(&req); err != nil {
		responses.RespondWithError(w, err)
		return
	}

	// SECURITY: Get staff ID from context with authorization check
	staffID, ok := getStaffIDFromContext(w, r)
	if !ok {
		return // Error already sent by helper
	}

	// Get IP address for audit logging
	ipAddress := middlewares.GetRealIP(r)

	if serviceErr := h.Service.DeactivateSubsidy(r.Context(), id, staffID, req.Reason, ipAddress); serviceErr != nil {
		responses.RespondWithError(w, serviceErr)
		return
	}

	responses.RespondWithSuccess(w, map[string]string{"message": "Subsidy deactivated successfully"}, http.StatusOK)
}

// GetSubsidySummary retrieves overall subsidy statistics
// @Summary Get subsidy summary
// @Description Get overall subsidy statistics and reports (admin only)
// @Tags Subsidies - Admin
// @Produce json
// @Success 200 {object} responses.SuccessResponse
// @Failure 401 {object} responses.ErrorResponse
// @Failure 403 {object} responses.ErrorResponse
// @Security Bearer
// @Router /subsidies/summary [get]
func (h *SubsidyHandler) GetSubsidySummary(w http.ResponseWriter, r *http.Request) {
	summary, err := h.Service.GetSubsidySummary(r.Context())
	if err != nil {
		responses.RespondWithError(w, err)
		return
	}

	responses.RespondWithSuccess(w, summary, http.StatusOK)
}

// ===== CUSTOMER HANDLERS =====

// GetMySubsidies retrieves current user's subsidies
// @Summary Get my subsidies
// @Description Get list of subsidies for the authenticated customer
// @Tags Subsidies - Customer
// @Produce json
// @Param page query integer false "Page number" default(1)
// @Param limit query integer false "Items per page" default(10)
// @Success 200 {object} responses.SuccessResponse
// @Failure 401 {object} responses.ErrorResponse
// @Security Bearer
// @Router /subsidies/me [get]
func (h *SubsidyHandler) GetMySubsidies(w http.ResponseWriter, r *http.Request) {
	// SECURITY: Extract customer ID from auth context
	customerID, ok := getCustomerIDFromContext(w, r)
	if !ok {
		return // Error already sent by helper
	}

	filters := dto.SubsidyFilters{
		CustomerID: &customerID,
		Page:       parseIntQuery(r, "page", 1),
		Limit:      parseIntQuery(r, "limit", 10),
	}

	result, err := h.Service.ListSubsidies(r.Context(), filters)
	if err != nil {
		responses.RespondWithError(w, err)
		return
	}

	responses.RespondWithSuccess(w, result, http.StatusOK)
}

// GetMyBalance retrieves current user's subsidy balance
// @Summary Get my balance
// @Description Get current subsidy balance for the authenticated customer
// @Tags Subsidies - Customer
// @Produce json
// @Success 200 {object} responses.SuccessResponse
// @Failure 401 {object} responses.ErrorResponse
// @Security Bearer
// @Router /subsidies/me/balance [get]
func (h *SubsidyHandler) GetMyBalance(w http.ResponseWriter, r *http.Request) {
	// SECURITY: Extract customer ID from auth context
	customerID, ok := getCustomerIDFromContext(w, r)
	if !ok {
		return // Error already sent by helper
	}

	balance, err := h.Service.GetCustomerBalance(r.Context(), customerID)
	if err != nil {
		responses.RespondWithError(w, err)
		return
	}

	responses.RespondWithSuccess(w, balance, http.StatusOK)
}

// GetMyUsageHistory retrieves current user's usage history
// @Summary Get my usage history
// @Description Get subsidy usage history for the authenticated customer
// @Tags Subsidies - Customer
// @Produce json
// @Param page query integer false "Page number" default(1)
// @Param limit query integer false "Items per page" default(20)
// @Success 200 {object} responses.SuccessResponse
// @Failure 401 {object} responses.ErrorResponse
// @Security Bearer
// @Router /subsidies/me/usage [get]
func (h *SubsidyHandler) GetMyUsageHistory(w http.ResponseWriter, r *http.Request) {
	// SECURITY: Extract customer ID from auth context
	customerID, ok := getCustomerIDFromContext(w, r)
	if !ok {
		return // Error already sent by helper
	}

	page := parseIntQuery(r, "page", 1)
	limit := parseIntQuery(r, "limit", 20)

	history, err := h.Service.GetCustomerUsageHistory(r.Context(), customerID, page, limit)
	if err != nil {
		responses.RespondWithError(w, err)
		return
	}

	responses.RespondWithSuccess(w, history, http.StatusOK)
}

// ===== HELPER FUNCTIONS =====

// getStaffIDFromContext extracts staff ID from context with proper error handling
func getStaffIDFromContext(w http.ResponseWriter, r *http.Request) (uuid.UUID, bool) {
	staffID, err := contextUtils.GetUserID(r.Context())
	if err != nil {
		responses.RespondWithError(w, err)
		return uuid.Nil, false
	}

	// Verify the user is actually staff
	isStaff, err := contextUtils.IsStaff(r.Context())
	if err != nil {
		responses.RespondWithError(w, err)
		return uuid.Nil, false
	}

	if !isStaff {
		responses.RespondWithError(w, errLib.New("Access denied: staff role required", http.StatusForbidden))
		return uuid.Nil, false
	}

	return staffID, true
}

// getCustomerIDFromContext extracts customer ID from context with proper error handling
func getCustomerIDFromContext(w http.ResponseWriter, r *http.Request) (uuid.UUID, bool) {
	customerID, err := contextUtils.GetUserID(r.Context())
	if err != nil {
		responses.RespondWithError(w, err)
		return uuid.Nil, false
	}
	return customerID, true
}

func parseIntQuery(r *http.Request, key string, defaultVal int) int {
	valStr := r.URL.Query().Get(key)
	if valStr == "" {
		return defaultVal
	}

	val, err := strconv.Atoi(valStr)
	if err != nil {
		return defaultVal
	}

	return val
}
