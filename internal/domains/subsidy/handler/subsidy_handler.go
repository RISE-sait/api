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

	subsidy, err := h.Service.CreateSubsidy(r.Context(), &req, staffID)
	if err != nil {
		responses.RespondWithError(w, err)
		return
	}

	responses.RespondWithSuccess(w, subsidy, http.StatusCreated)
}

// GetSubsidy retrieves subsidy details
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

	if serviceErr := h.Service.DeactivateSubsidy(r.Context(), id, staffID, req.Reason); serviceErr != nil {
		responses.RespondWithError(w, serviceErr)
		return
	}

	responses.RespondWithSuccess(w, map[string]string{"message": "Subsidy deactivated successfully"}, http.StatusOK)
}

// GetSubsidySummary retrieves overall subsidy statistics
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
