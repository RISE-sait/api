package membership

import (
	membership "api/internal/domains/membership/application"
	"api/internal/domains/membership/dto"
	"api/internal/domains/membership/mapper"
	response_handlers "api/internal/libs/responses"
	"api/internal/libs/validators"
	"net/http"

	"github.com/go-chi/chi"
)

type Handler struct {
	Service *membership.MembershipService
}

func NewHandler(service *membership.MembershipService) *Handler {
	return &Handler{Service: service}
}

func (h *Handler) CreateMembership(w http.ResponseWriter, r *http.Request) {
	var requestDto dto.CreateMembershipRequest

	if err := validators.ParseAndValidateJSON(r.Body, &requestDto); err != nil {
		response_handlers.RespondWithError(w, err)
		return
	}

	membership := mapper.MapCreateRequestToEntity(requestDto)

	if err := h.Service.Create(r.Context(), &membership); err != nil {
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

	membership, err := h.Service.GetById(r.Context(), id)
	if err != nil {
		response_handlers.RespondWithError(w, err)
		return
	}

	response_handlers.RespondWithSuccess(w, *membership, http.StatusOK)
}

func (h *Handler) GetAllMemberships(w http.ResponseWriter, r *http.Request) {
	memberships, err := h.Service.GetAll(r.Context())
	if err != nil {
		response_handlers.RespondWithError(w, err)
		return
	}

	result := []dto.MembershipResponse{}
	for i, membership := range memberships {
		result[i] = mapper.MapEntityToResponse(&membership)
	}

	response_handlers.RespondWithSuccess(w, result, http.StatusOK)
}

func (h *Handler) UpdateMembership(w http.ResponseWriter, r *http.Request) {
	var requestDto dto.UpdateMembershipRequest

	if err := validators.ParseAndValidateJSON(r.Body, &requestDto); err != nil {
		response_handlers.RespondWithError(w, err)
		return
	}

	membership := mapper.MapUpdateRequestToEntity(requestDto)

	if err := h.Service.Update(r.Context(), &membership); err != nil {
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

	if err := h.Service.Delete(r.Context(), id); err != nil {
		response_handlers.RespondWithError(w, err)
		return
	}

	response_handlers.RespondWithSuccess(w, nil, http.StatusNoContent)
}
