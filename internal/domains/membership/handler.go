package membership

import (
	"api/internal/domains/membership/dto"
	response_handlers "api/internal/libs/responses"
	"api/internal/libs/validators"
	"net/http"

	"github.com/go-chi/chi"
)

type Handler struct {
	Service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{Service: service}
}

func (h *Handler) CreateMembership(w http.ResponseWriter, r *http.Request) {
	var targetBody dto.CreateMembershipRequest

	if err := validators.ParseRequestBodyToJSON(r.Body, &targetBody); err != nil {
		response_handlers.RespondWithError(w, err)
		return
	}

	if err := h.Service.CreateMembership(r.Context(), targetBody); err != nil {
		response_handlers.RespondWithError(w, err)
		return
	}

	response_handlers.RespondWithSuccess(w, nil, http.StatusCreated)
}

func (h *Handler) GetMembershipById(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")

	id, err := validators.ParseUUID(idStr)
	if err != nil {
		response_handlers.RespondWithError(w, err)
		return
	}

	membership, err := h.Service.GetMembershipById(r.Context(), id)
	if err != nil {
		response_handlers.RespondWithError(w, err)
		return
	}

	response_handlers.RespondWithSuccess(w, *membership, http.StatusOK)
}

func (h *Handler) GetAllMemberships(w http.ResponseWriter, r *http.Request) {
	memberships, err := h.Service.GetAllMemberships(r.Context())
	if err != nil {
		response_handlers.RespondWithError(w, err)
		return
	}

	response_handlers.RespondWithSuccess(w, memberships, http.StatusOK)
}

func (h *Handler) UpdateMembership(w http.ResponseWriter, r *http.Request) {
	var targetBody dto.UpdateMembershipRequest

	if err := validators.ParseRequestBodyToJSON(r.Body, &targetBody); err != nil {
		response_handlers.RespondWithError(w, err)
		return
	}

	if err := h.Service.UpdateMembership(r.Context(), targetBody); err != nil {
		response_handlers.RespondWithError(w, err)
		return
	}

	response_handlers.RespondWithSuccess(w, nil, http.StatusNoContent)
}

func (h *Handler) DeleteMembership(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")

	id, err := validators.ParseUUID(idStr)
	if err != nil {
		response_handlers.RespondWithError(w, err)
		return
	}

	if err := h.Service.DeleteMembership(r.Context(), id); err != nil {
		response_handlers.RespondWithError(w, err)
		return
	}

	response_handlers.RespondWithSuccess(w, nil, http.StatusNoContent)
}
