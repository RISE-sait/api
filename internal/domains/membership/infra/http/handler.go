package membership

import (
	membership "api/internal/domains/membership/application"
	"api/internal/domains/membership/infra/http/dto"
	"api/internal/domains/membership/infra/http/mapper"
	"api/internal/domains/membership/values"
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

	if err := validators.ParseJSON(r.Body, &requestDto); err != nil {
		response_handlers.RespondWithError(w, err)
		return
	}

	membershipCreate := values.NewMembershipCreate(requestDto.Name, requestDto.Description, requestDto.StartDate, requestDto.EndDate)

	if err := h.Service.Create(r.Context(), membershipCreate); err != nil {
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

	result := make([]dto.MembershipResponse, len(memberships))
	for i, membership := range memberships {
		result[i] = mapper.MapEntityToResponse(&membership)
	}

	response_handlers.RespondWithSuccess(w, result, http.StatusOK)
}

func (h *Handler) UpdateMembership(w http.ResponseWriter, r *http.Request) {

	idStr := chi.URLParam(r, "id")

	id, err := validators.ParseUUID(idStr)
	if err != nil {
		response_handlers.RespondWithError(w, err)
		return
	}

	var dto dto.UpdateMembershipRequest

	if err := validators.ParseJSON(r.Body, &dto); err != nil {
		response_handlers.RespondWithError(w, err)
		return
	}

	membershipUpdate := values.NewMembershipUpdate(id, dto.Name, dto.Description, dto.StartDate, dto.EndDate)

	if err := h.Service.Update(r.Context(), membershipUpdate); err != nil {
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
