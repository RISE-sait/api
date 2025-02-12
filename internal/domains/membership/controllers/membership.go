package membership

import (
	"api/internal/di"
	dto "api/internal/domains/membership/dto"
	service "api/internal/domains/membership/services"
	values "api/internal/domains/membership/values/memberships"
	response_handlers "api/internal/libs/responses"
	"api/internal/libs/validators"
	"net/http"

	"github.com/go-chi/chi"
)

type MembershipController struct {
	Service *service.MembershipService
}

func NewMembershipController(container *di.Container) *MembershipController {
	return &MembershipController{Service: service.NewMembershipService(container)}
}

// CreateMembership creates a new membership.
// @Summary Create a new membership
// @Description Create a new membership
// @Tags memberships
// @Accept json
// @Produce json
// @Param membership body dto.MembershipRequestDto true "Membership details"
// @Security Bearer
// @Success 201 "Membership created successfully"
// @Failure 400 {object} map[string]interface{} "Bad Request: Invalid input"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /api/memberships [post]
func (h *MembershipController) CreateMembership(w http.ResponseWriter, r *http.Request) {
	var dto dto.MembershipRequestDto

	if err := validators.ParseJSON(r.Body, &dto); err != nil {
		response_handlers.RespondWithError(w, err)
		return
	}

	membership, err := dto.ToMembershipCreateValueObject()

	if err != nil {
		response_handlers.RespondWithError(w, err)
		return
	}

	if err := h.Service.Create(r.Context(), membership); err != nil {
		response_handlers.RespondWithError(w, err)
		return
	}

	response_handlers.RespondWithSuccess(w, nil, http.StatusCreated)
}

// GetMembershipById retrieves a membership by ID.
// @Summary Get a membership by ID
// @Description Get a membership by ID
// @Tags memberships
// @Accept json
// @Produce json
// @Param id path string true "Membership ID" example("f47ac10b-58cc-4372-a567-0e02b2c3d479")
// @Success 200 {object} dto.MembershipResponse "Membership retrieved successfully"
// @Failure 400 {object} map[string]interface{} "Bad Request: Invalid ID"
// @Failure 404 {object} map[string]interface{} "Not Found: Membership not found"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /api/memberships/{id} [get]
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

// GetAllMemberships retrieves a list of memberships.
// @Summary Get a list of memberships
// @Description Get a list of memberships
// @Tags memberships
// @Accept json
// @Produce json
// @Success 200 {array} dto.MembershipResponse "List of memberships retrieved successfully"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /api/memberships [get]
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

// UpdateMembership updates an existing membership.
// @Summary Update a membership
// @Description Update a membership
// @Tags memberships
// @Accept json
// @Produce json
// @Param id path string true "Membership ID" example("f47ac10b-58cc-4372-a567-0e02b2c3d479")
// @Param membership body dto.MembershipRequestDto true "Membership details"
// @Security Bearer
// @Success 204 "No Content: Membership updated successfully"
// @Failure 400 {object} map[string]interface{} "Bad Request: Invalid input"
// @Failure 404 {object} map[string]interface{} "Not Found: Membership not found"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /api/memberships/{id} [put]
func (h *MembershipController) UpdateMembership(w http.ResponseWriter, r *http.Request) {

	idStr := chi.URLParam(r, "id")

	var dto dto.MembershipRequestDto

	if err := validators.ParseJSON(r.Body, &dto); err != nil {
		response_handlers.RespondWithError(w, err)
		return
	}

	membershipUpdate, err := dto.ToMembershipUpdateValueObject(idStr)

	if err != nil {
		response_handlers.RespondWithError(w, err)
		return
	}

	if err := h.Service.Update(r.Context(), membershipUpdate); err != nil {
		response_handlers.RespondWithError(w, err)
		return
	}

	response_handlers.RespondWithSuccess(w, nil, http.StatusNoContent)
}

// DeleteMembership deletes a membership by ID.
// @Summary Delete a membership
// @Description Delete a membership by ID
// @Tags memberships
// @Accept json
// @Produce json
// @Param id path string true "Membership ID" example("f47ac10b-58cc-4372-a567-0e02b2c3d479")
// @Security Bearer
// @Success 204 "No Content: Membership deleted successfully"
// @Failure 400 {object} map[string]interface{} "Bad Request: Invalid ID"
// @Failure 404 {object} map[string]interface{} "Not Found: Membership not found"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /api/memberships/{id} [delete]
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

func mapEntityToResponse(membership *values.MembershipAllFields) dto.MembershipResponse {
	return dto.MembershipResponse{
		ID:          membership.ID,
		Name:        membership.Name,
		Description: membership.Description,
	}
}
