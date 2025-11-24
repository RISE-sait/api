package user

import (
	"net/http"

	"api/internal/di"
	suspensionDto "api/internal/domains/user/dto/suspension"
	"api/internal/domains/user/services"
	errLib "api/internal/libs/errors"
	responseHandlers "api/internal/libs/responses"
	"api/internal/libs/validators"
	contextUtils "api/utils/context"

	"github.com/go-chi/chi"
)

type SuspensionHandler struct {
	suspensionService *services.SuspensionService
}

func NewSuspensionHandler(container *di.Container) *SuspensionHandler {
	return &SuspensionHandler{
		suspensionService: services.NewSuspensionService(container),
	}
}

// SuspendUser suspends a user account and their memberships
// @Summary Suspend user account
// @Description Suspends a user account, pauses their memberships and Stripe subscriptions. Requires admin role.
// @Tags customers
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path string true "User ID to suspend"
// @Param suspension_body body suspensionDto.SuspendUserRequestDto true "Suspension details"
// @Success 200 {object} map[string]interface{} "User suspended successfully"
// @Failure 400 {object} map[string]interface{} "Bad Request: Invalid input"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Forbidden: Insufficient permissions"
// @Failure 404 {object} map[string]interface{} "Not Found: User not found"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /customers/{id}/suspend [post]
func (h *SuspensionHandler) SuspendUser(w http.ResponseWriter, r *http.Request) {
	// Get user ID from URL
	userIDStr := chi.URLParam(r, "id")
	userID, parseErr := validators.ParseUUID(userIDStr)
	if parseErr != nil {
		responseHandlers.RespondWithError(w, parseErr)
		return
	}

	// Check authorization - only admins can suspend users
	userRole, roleErr := contextUtils.GetUserRole(r.Context())
	if roleErr != nil {
		responseHandlers.RespondWithError(w, errLib.New("Authentication required", http.StatusUnauthorized))
		return
	}

	if userRole != contextUtils.RoleAdmin && userRole != contextUtils.RoleSuperAdmin {
		responseHandlers.RespondWithError(w, errLib.New("Insufficient permissions to suspend users", http.StatusForbidden))
		return
	}

	// Get the staff member ID who is performing the suspension
	suspendedBy, userErr := contextUtils.GetUserID(r.Context())
	if userErr != nil {
		responseHandlers.RespondWithError(w, userErr)
		return
	}

	// Parse request body
	var requestDto suspensionDto.SuspendUserRequestDto
	if err := validators.ParseJSON(r.Body, &requestDto); err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	// Parse and validate the suspension duration
	duration, durationErr := requestDto.ParseDuration()
	if durationErr != nil {
		responseHandlers.RespondWithError(w, durationErr)
		return
	}

	// Call service to suspend user
	params := services.SuspendUserParams{
		UserID:             userID,
		SuspendedBy:        suspendedBy,
		SuspensionReason:   requestDto.SuspensionReason,
		SuspensionDuration: duration,
	}

	if err := h.suspensionService.SuspendUser(r.Context(), params); err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	responseHandlers.RespondWithSuccess(w, map[string]interface{}{
		"message": "User suspended successfully",
		"user_id": userID,
	}, http.StatusOK)
}

// UnsuspendUser removes suspension from a user account
// @Summary Unsuspend user account
// @Description Unsuspends a user account, resumes their memberships and Stripe subscriptions. Optionally extends membership by suspension duration. Requires admin role.
// @Tags customers
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path string true "User ID to unsuspend"
// @Param unsuspension_body body suspensionDto.UnsuspendUserRequestDto true "Unsuspension options"
// @Success 200 {object} map[string]interface{} "User unsuspended successfully"
// @Failure 400 {object} map[string]interface{} "Bad Request: Invalid input or user not suspended"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Forbidden: Insufficient permissions"
// @Failure 404 {object} map[string]interface{} "Not Found: User not found"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /customers/{id}/unsuspend [post]
func (h *SuspensionHandler) UnsuspendUser(w http.ResponseWriter, r *http.Request) {
	// Get user ID from URL
	userIDStr := chi.URLParam(r, "id")
	userID, parseErr := validators.ParseUUID(userIDStr)
	if parseErr != nil {
		responseHandlers.RespondWithError(w, parseErr)
		return
	}

	// Check authorization - only admins can unsuspend users
	userRole, roleErr := contextUtils.GetUserRole(r.Context())
	if roleErr != nil {
		responseHandlers.RespondWithError(w, errLib.New("Authentication required", http.StatusUnauthorized))
		return
	}

	if userRole != contextUtils.RoleAdmin && userRole != contextUtils.RoleSuperAdmin {
		responseHandlers.RespondWithError(w, errLib.New("Insufficient permissions to unsuspend users", http.StatusForbidden))
		return
	}

	// Get the staff member ID who is performing the unsuspension
	unsuspendedBy, userErr := contextUtils.GetUserID(r.Context())
	if userErr != nil {
		responseHandlers.RespondWithError(w, userErr)
		return
	}

	// Parse request body
	var requestDto suspensionDto.UnsuspendUserRequestDto
	if err := validators.ParseJSON(r.Body, &requestDto); err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	// Call service to unsuspend user
	params := services.UnsuspendUserParams{
		UserID:           userID,
		UnsuspendedBy:    unsuspendedBy,
		ExtendMembership: requestDto.ExtendMembership,
		CollectArrears:   requestDto.CollectArrears,
	}

	if err := h.suspensionService.UnsuspendUser(r.Context(), params); err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	responseHandlers.RespondWithSuccess(w, map[string]interface{}{
		"message": "User unsuspended successfully",
		"user_id": userID,
	}, http.StatusOK)
}

// GetSuspensionInfo retrieves suspension information for a user
// @Summary Get user suspension status
// @Description Retrieves suspension information for a user including reason, suspended by, and expiration date
// @Tags customers
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path string true "User ID"
// @Success 200 {object} suspensionDto.SuspensionInfoResponseDto "Suspension information"
// @Failure 400 {object} map[string]interface{} "Bad Request: Invalid user ID"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 404 {object} map[string]interface{} "Not Found: User not found"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /customers/{id}/suspension [get]
func (h *SuspensionHandler) GetSuspensionInfo(w http.ResponseWriter, r *http.Request) {
	// Get user ID from URL
	userIDStr := chi.URLParam(r, "id")
	userID, parseErr := validators.ParseUUID(userIDStr)
	if parseErr != nil {
		responseHandlers.RespondWithError(w, parseErr)
		return
	}

	// Check authorization - admins can view any user's suspension, users can view their own
	userRole, roleErr := contextUtils.GetUserRole(r.Context())
	if roleErr != nil {
		responseHandlers.RespondWithError(w, errLib.New("Authentication required", http.StatusUnauthorized))
		return
	}

	// If not admin, verify user is checking their own suspension status
	if userRole != contextUtils.RoleAdmin && userRole != contextUtils.RoleSuperAdmin {
		currentUserID, userErr := contextUtils.GetUserID(r.Context())
		if userErr != nil {
			responseHandlers.RespondWithError(w, userErr)
			return
		}

		if currentUserID != userID {
			responseHandlers.RespondWithError(w, errLib.New("You can only view your own suspension status", http.StatusForbidden))
			return
		}
	}

	// Query suspension info from database
	suspensionInfo, err := h.suspensionService.GetSuspensionInfo(r.Context(), userID)
	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	// Build response DTO
	response := suspensionDto.SuspensionInfoResponseDto{
		IsSuspended:         suspensionInfo.SuspendedAt.Valid,
		SuspendedAt:         nil,
		SuspensionReason:    nil,
		SuspendedBy:         nil,
		SuspensionExpiresAt: nil,
	}

	if suspensionInfo.SuspendedAt.Valid {
		response.SuspendedAt = &suspensionInfo.SuspendedAt.Time
	}

	if suspensionInfo.SuspensionReason.Valid {
		response.SuspensionReason = &suspensionInfo.SuspensionReason.String
	}

	if suspensionInfo.SuspendedBy != nil {
		if name, ok := suspensionInfo.SuspendedBy.(string); ok && name != "" {
			response.SuspendedBy = &name
		}
	}

	if suspensionInfo.SuspensionExpiresAt.Valid {
		response.SuspensionExpiresAt = &suspensionInfo.SuspensionExpiresAt.Time
	}

	responseHandlers.RespondWithSuccess(w, response, http.StatusOK)
}

// CollectArrears manually collects arrears for a suspended user
// @Summary Manually collect arrears for suspended user
// @Description Calculates and creates Stripe invoice items for missed billing periods during suspension. Does not unsuspend the user. Requires admin role.
// @Tags customers
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path string true "User ID to collect arrears for"
// @Success 200 {object} map[string]interface{} "Arrears collected successfully"
// @Failure 400 {object} map[string]interface{} "Bad Request: Invalid input or user not suspended"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Forbidden: Insufficient permissions"
// @Failure 404 {object} map[string]interface{} "Not Found: User not found"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /customers/{id}/collect-arrears [post]
func (h *SuspensionHandler) CollectArrears(w http.ResponseWriter, r *http.Request) {
	// Get user ID from URL
	userIDStr := chi.URLParam(r, "id")
	userID, parseErr := validators.ParseUUID(userIDStr)
	if parseErr != nil {
		responseHandlers.RespondWithError(w, parseErr)
		return
	}

	// Check authorization - only admins can collect arrears
	userRole, roleErr := contextUtils.GetUserRole(r.Context())
	if roleErr != nil {
		responseHandlers.RespondWithError(w, errLib.New("Authentication required", http.StatusUnauthorized))
		return
	}

	if userRole != contextUtils.RoleAdmin && userRole != contextUtils.RoleSuperAdmin {
		responseHandlers.RespondWithError(w, errLib.New("Insufficient permissions to collect arrears", http.StatusForbidden))
		return
	}

	// Get the staff member ID who is performing the arrears collection
	collectedBy, userErr := contextUtils.GetUserID(r.Context())
	if userErr != nil {
		responseHandlers.RespondWithError(w, userErr)
		return
	}

	// Call service to collect arrears
	arrearsTotal, err := h.suspensionService.CollectArrearsManually(r.Context(), userID, collectedBy)
	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	responseHandlers.RespondWithSuccess(w, map[string]interface{}{
		"message":       "Arrears collected successfully",
		"user_id":       userID,
		"arrears_total": arrearsTotal,
	}, http.StatusOK)
}
