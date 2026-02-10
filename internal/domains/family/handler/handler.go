package family

import (
	"net/http"

	"api/internal/di"
	dto "api/internal/domains/family/dto"
	service "api/internal/domains/family/service"
	responseHandlers "api/internal/libs/responses"
	"api/internal/libs/validators"

	"github.com/go-chi/chi"
)

type Handler struct {
	Service *service.Service
}

func NewFamilyHandler(container *di.Container) *Handler {
	return &Handler{Service: service.NewService(container)}
}

// RequestLink initiates a parent-child link request.
// @Summary Initiate a parent-child link request
// @Description Initiates a link request between parent and child. If caller has no parent, they're adding a child. If caller has a parent, they're changing/adding a parent.
// @Tags family
// @Accept json
// @Produce json
// @Param body body dto.RequestLinkRequest true "Target email address"
// @Success 200 {object} dto.RequestLinkResponse "Link request initiated"
// @Failure 400 {object} map[string]interface{} "Bad Request: Invalid input"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 404 {object} map[string]interface{} "User not found"
// @Failure 409 {object} map[string]interface{} "Pending request already exists"
// @Router /family/link/request [post]
func (h *Handler) RequestLink(w http.ResponseWriter, r *http.Request) {
	var requestDto dto.RequestLinkRequest

	if err := validators.ParseJSON(r.Body, &requestDto); err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	if err := requestDto.Validate(); err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	response, err := h.Service.RequestLink(r.Context(), requestDto.TargetEmail)
	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	responseHandlers.RespondWithSuccess(w, response, http.StatusOK)
}

// ConfirmLink confirms a link request with a verification code.
// @Summary Confirm a parent-child link request
// @Description Confirms a link request using the verification code sent via email
// @Tags family
// @Accept json
// @Produce json
// @Param body body dto.ConfirmLinkRequest true "Verification code"
// @Success 200 {object} dto.ConfirmLinkResponse "Link confirmation status"
// @Failure 400 {object} map[string]interface{} "Bad Request: Invalid input"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Forbidden: Code not for this user"
// @Failure 404 {object} map[string]interface{} "Invalid or expired code"
// @Router /family/link/confirm [post]
func (h *Handler) ConfirmLink(w http.ResponseWriter, r *http.Request) {
	var requestDto dto.ConfirmLinkRequest

	if err := validators.ParseJSON(r.Body, &requestDto); err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	if err := requestDto.Validate(); err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	response, err := h.Service.ConfirmLink(r.Context(), requestDto.Code)
	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	responseHandlers.RespondWithSuccess(w, response, http.StatusOK)
}

// CancelRequest cancels a pending link request.
// @Summary Cancel a pending link request
// @Description Cancels the caller's pending link request
// @Tags family
// @Produce json
// @Success 204 "Request cancelled"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 404 {object} map[string]interface{} "No pending request found"
// @Router /family/link/request [delete]
func (h *Handler) CancelRequest(w http.ResponseWriter, r *http.Request) {
	if err := h.Service.CancelRequest(r.Context()); err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	responseHandlers.RespondWithSuccess(w, nil, http.StatusNoContent)
}

// GetPendingRequests gets all pending link requests involving the user.
// @Summary Get pending link requests
// @Description Gets all pending link requests where the user is child, new parent, or old parent
// @Tags family
// @Produce json
// @Success 200 {array} dto.PendingRequestResponse "Pending requests"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Router /family/link/requests [get]
func (h *Handler) GetPendingRequests(w http.ResponseWriter, r *http.Request) {
	requests, err := h.Service.GetPendingRequests(r.Context())
	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	responseHandlers.RespondWithSuccess(w, requests, http.StatusOK)
}

// GetChildren gets all children for the authenticated parent.
// @Summary Get children
// @Description Gets all children linked to the authenticated parent
// @Tags family
// @Produce json
// @Success 200 {array} dto.ChildResponse "Children list"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Router /family/children [get]
func (h *Handler) GetChildren(w http.ResponseWriter, r *http.Request) {
	children, err := h.Service.GetChildren(r.Context())
	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	responseHandlers.RespondWithSuccess(w, children, http.StatusOK)
}

// AdminUnlink removes a parent-child link (admin only).
// @Summary Admin unlink parent-child
// @Description Removes the parent-child link for a user (admin only)
// @Tags family
// @Produce json
// @Param id path string true "Child user ID"
// @Success 204 "Link removed"
// @Failure 400 {object} map[string]interface{} "Bad Request: No parent link exists"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Forbidden: Admin only"
// @Failure 404 {object} map[string]interface{} "User not found"
// @Router /admin/family/link/{id} [delete]
func (h *Handler) AdminUnlink(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := validators.ParseUUID(idStr)
	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	if err := h.Service.AdminUnlink(r.Context(), id); err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	responseHandlers.RespondWithSuccess(w, nil, http.StatusNoContent)
}
