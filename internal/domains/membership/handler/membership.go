package membership

import (
	"api/internal/di"
	dto "api/internal/domains/membership/dto/membership"
	service "api/internal/domains/membership/services"
	values "api/internal/domains/membership/values"
	responseHandlers "api/internal/libs/responses"
	"api/internal/libs/validators"
	"net/http"

	"github.com/go-chi/chi"
)

type Handlers struct {
	Service *service.Service
}

func NewHandlers(container *di.Container) *Handlers {
	return &Handlers{Service: service.NewMembershipService(container)}
}

// CreateMembership creates a new membership.
// @Summary Create a new membership
// @Description Create a new membership
// @Tags memberships
// @Accept json
// @Produce json
// @Param membership body dto.RequestDto true "Membership details"
// @Security Bearer
// @Success 201 "Membership created successfully"
// @Failure 400 {object} map[string]interface{} "Bad Request: Invalid input"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /memberships [post]
func (h *Handlers) CreateMembership(w http.ResponseWriter, r *http.Request) {
	var requestDto dto.RequestDto

	if err := validators.ParseJSON(r.Body, &requestDto); err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	membership, err := requestDto.ToMembershipCreateValueObject()

	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	if err := h.Service.Create(r.Context(), membership); err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	responseHandlers.RespondWithSuccess(w, nil, http.StatusCreated)
}

// GetMembershipById retrieves a membership by HubSpotId.
// @Summary Get a membership by HubSpotId
// @Description Get a membership by HubSpotId
// @Tags memberships
// @Accept json
// @Produce json
// @Param id path string true "Membership ID" example("f47ac10b-58cc-4372-a567-0e02b2c3d479")
// @Success 200 {object} membership.Response "Membership retrieved successfully"
// @Failure 400 {object} map[string]interface{} "Bad Request: Invalid HubSpotId"
// @Failure 404 {object} map[string]interface{} "Not Found: Membership not found"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /memberships/{id} [get]
func (h *Handlers) GetMembershipById(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")

	id, err := validators.ParseUUID(idStr)
	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	membership, err := h.Service.GetById(r.Context(), id)
	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	response := mapEntityToResponse(*membership)

	responseHandlers.RespondWithSuccess(w, response, http.StatusOK)
}

// GetMemberships retrieves a list of memberships.
// @Summary Get a list of memberships
// @Description Get a list of memberships
// @Tags memberships
// @Accept json
// @Produce json
// @Success 200 {array} membership.Response "GetMemberships of memberships retrieved successfully"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /memberships [get]
func (h *Handlers) GetMemberships(w http.ResponseWriter, r *http.Request) {
	memberships, err := h.Service.GetMemberships(r.Context())
	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	result := make([]dto.Response, len(memberships))
	for i, membership := range memberships {
		result[i] = mapEntityToResponse(membership)
	}

	responseHandlers.RespondWithSuccess(w, result, http.StatusOK)
}

// UpdateMembership updates an existing membership.
// @Summary Update a membership
// @Description Update a membership
// @Tags memberships
// @Accept json
// @Produce json
// @Param id path string true "Membership ID" example("f47ac10b-58cc-4372-a567-0e02b2c3d479")
// @Param membership body dto.RequestDto true "Membership details"
// @Security Bearer
// @Success 204 "No Content: Membership updated successfully"
// @Failure 400 {object} map[string]interface{} "Bad Request: Invalid input"
// @Failure 404 {object} map[string]interface{} "Not Found: Membership not found"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /memberships/{id} [put]
func (h *Handlers) UpdateMembership(w http.ResponseWriter, r *http.Request) {

	idStr := chi.URLParam(r, "id")

	var requestDto dto.RequestDto

	if err := validators.ParseJSON(r.Body, &requestDto); err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	membershipUpdate, err := requestDto.ToMembershipUpdateValueObject(idStr)

	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	if err := h.Service.Update(r.Context(), membershipUpdate); err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	responseHandlers.RespondWithSuccess(w, nil, http.StatusNoContent)
}

// DeleteMembership deletes a membership by HubSpotId.
// @Summary Delete a membership
// @Description Delete a membership by HubSpotId
// @Tags memberships
// @Accept json
// @Produce json
// @Param id path string true "Membership ID" example("f47ac10b-58cc-4372-a567-0e02b2c3d479")
// @Security Bearer
// @Success 204 "No Content: Membership deleted successfully"
// @Failure 400 {object} map[string]interface{} "Bad Request: Invalid HubSpotId"
// @Failure 404 {object} map[string]interface{} "Not Found: Membership not found"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /memberships/{id} [delete]
func (h *Handlers) DeleteMembership(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")

	id, err := validators.ParseUUID(idStr)
	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	if err := h.Service.Delete(r.Context(), id); err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	responseHandlers.RespondWithSuccess(w, nil, http.StatusNoContent)
}

func mapEntityToResponse(entity values.ReadValues) dto.Response {
	return dto.Response{
		ID:          entity.ID,
		Name:        entity.Name,
		Description: entity.Description,
		CreatedAt:   entity.CreatedAt,
		UpdatedAt:   entity.UpdatedAt,
	}
}
