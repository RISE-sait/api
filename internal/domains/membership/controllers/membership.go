package membership

import (
	"api/cmd/server/di"
	membership "api/internal/domains/membership/application"
	entity "api/internal/domains/membership/entities"
	"api/internal/domains/membership/infra/http/dto"
	persistence "api/internal/domains/membership/infra/persistence"
	"api/internal/domains/membership/values"
	response_handlers "api/internal/libs/responses"
	"api/internal/libs/validators"
	"net/http"

	"github.com/go-chi/chi"
)

type MembershipController struct {
	Service *membership.MembershipService
}

func NewMembershipController(container *di.Container) *MembershipController {
	return &MembershipController{Service: membership.NewMembershipService(
		&persistence.MembershipsRepository{
			Queries: container.Queries.MembershipDb,
		},
	)}
}

func (h *MembershipController) CreateMembership(w http.ResponseWriter, r *http.Request) {
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

func (h *MembershipController) GetMembershipById(w http.ResponseWriter, r *http.Request) {
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

	response := mapEntityToResponse(membership)

	response_handlers.RespondWithSuccess(w, response, http.StatusOK)
}

func (h *MembershipController) GetAllMemberships(w http.ResponseWriter, r *http.Request) {
	memberships, err := h.Service.GetAll(r.Context())
	if err != nil {
		response_handlers.RespondWithError(w, err)
		return
	}

	result := make([]dto.MembershipResponse, len(memberships))
	for i, membership := range memberships {
		result[i] = mapEntityToResponse(&membership)
	}

	response_handlers.RespondWithSuccess(w, result, http.StatusOK)
}

func (h *MembershipController) UpdateMembership(w http.ResponseWriter, r *http.Request) {

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

func (h *MembershipController) DeleteMembership(w http.ResponseWriter, r *http.Request) {
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

func mapEntityToResponse(membership *entity.Membership) dto.MembershipResponse {
	return dto.MembershipResponse{
		ID:          membership.ID,
		Name:        membership.Name,
		Description: membership.Description,
		StartDate:   membership.StartDate,
		EndDate:     membership.EndDate,
	}
}
