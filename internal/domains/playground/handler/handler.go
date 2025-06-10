package playground

import (
	"api/internal/di"
	dto "api/internal/domains/playground/dto"
	service "api/internal/domains/playground/services"
	responseHandlers "api/internal/libs/responses"
	"api/internal/libs/validators"
	"net/http"

	contextUtils "api/utils/context"

	"github.com/go-chi/chi"
)

type Handler struct {
	service *service.Service
}

func NewHandler(container *di.Container) *Handler {
	return &Handler{service: service.NewService(container)}
}

// CreateSession handles the creation of a new session.
// @Summary Create a new session
// @Tags playground
// @Accept json
// @Produce json
// @Param session body dto.RequestDto true "Session request body"
// @Success 201 {object} dto.ResponseDto "Session created successfully"
// @Failure 400 {object} map[string]interface{} "Invalid input"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 500 {object} map[string]interface{} "Server error"
// @Security Bearer
// @Router /playground [post]

func (h *Handler) CreateSession(w http.ResponseWriter, r *http.Request) {
	customerID, errCtx := contextUtils.GetUserID(r.Context())
	if errCtx != nil {
		responseHandlers.RespondWithError(w, errCtx)
		return
	}

	var req dto.RequestDto
	if err := validators.ParseJSON(r.Body, &req); err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	value, err := req.ToCreateValue(customerID)
	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	session, err := h.service.CreateSession(r.Context(), value)
	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	responseHandlers.RespondWithSuccess(w, dto.NewResponse(session), http.StatusCreated)
}

// GetSessions retrieves all sessions.
// @Summary Get all sessions
// @Tags playground
// @Produce json
// @Success 200 {array} dto.ResponseDto "List of sessions"
// @Failure 500 {object} map[string]interface{} "Server error"
// @Router /playground [get]

func (h *Handler) GetSessions(w http.ResponseWriter, r *http.Request) {
	sessions, err := h.service.GetSessions(r.Context())
	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}
	resp := make([]dto.ResponseDto, len(sessions))
	for i, s := range sessions {
		resp[i] = dto.NewResponse(s)
	}
	responseHandlers.RespondWithSuccess(w, resp, http.StatusOK)
}

// GetSession retrieves a specific session by its ID.
// @Summary Get session by ID
// @Tags playground
// @Produce json
// @Param id path string true "Session ID" example("f47ac10b-58cc-4372-a567-0e02b2c3d479")
// @Success 200 {object} dto.ResponseDto "Session found"
// @Failure 400 {object} map[string]interface{} "Invalid ID"
// @Failure 404 {object} map[string]interface{} "Session not found"
// @Failure 500 {object} map[string]interface{} "Server error"
// @Router /playground/{id} [get]

func (h *Handler) GetSession(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := validators.ParseUUID(idStr)
	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	session, err := h.service.GetSession(r.Context(), id)
	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}
	responseHandlers.RespondWithSuccess(w, dto.NewResponse(session), http.StatusOK)
}

// DeleteSession deletes a session by its ID.
// @Summary Delete a session
// @Tags playground
// @Param id path string true "Session ID" example("f47ac10b-58cc-4372-a567-0e02b2c3d479")
// @Success 204 "Deleted successfully"
// @Failure 400 {object} map[string]interface{} "Invalid ID"
// @Failure 404 {object} map[string]interface{} "Session not found"
// @Failure 500 {object} map[string]interface{} "Server error"
// @Security Bearer
// @Router /playground/{id} [delete]

func (h *Handler) DeleteSession(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := validators.ParseUUID(idStr)
	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	if err := h.service.DeleteSession(r.Context(), id); err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}
	responseHandlers.RespondWithSuccess(w, nil, http.StatusNoContent)
}
