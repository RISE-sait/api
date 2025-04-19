package membership

import (
	"api/internal/di"
	dto "api/internal/domains/membership/dto/membership"
	repo "api/internal/domains/membership/persistence/repositories"
	values "api/internal/domains/membership/values"
	responseHandlers "api/internal/libs/responses"
	"api/internal/libs/validators"
	"net/http"

	"github.com/go-chi/chi"
)

type Handlers struct {
	Repo *repo.Repository
}

func NewHandlers(container *di.Container) *Handlers {
	return &Handlers{Repo: repo.NewMembershipsRepository(container)}
}

// CreateMembership creates a new membership.
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

	if err = h.Repo.Create(r.Context(), membership); err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	responseHandlers.RespondWithSuccess(w, nil, http.StatusCreated)
}

// GetMembershipById retrieves a membership by ID.
// @Tags memberships
// @Accept json
// @Produce json
// @Param id path string true "Membership ID" example("f47ac10b-58cc-4372-a567-0e02b2c3d479")
// @Success 200 {object} membership.Response "Membership retrieved successfully"
// @Failure 400 {object} map[string]interface{} "Bad Request: Invalid ID"
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

	membership, err := h.Repo.GetByID(r.Context(), id)
	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	response := mapReadValueToResponse(*membership)

	responseHandlers.RespondWithSuccess(w, response, http.StatusOK)
}

// GetMemberships retrieves a list of memberships.
// @Tags memberships
// @Accept json
// @Produce json
// @Success 200 {array} membership.Response "GetMemberships of memberships retrieved successfully"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /memberships [get]
func (h *Handlers) GetMemberships(w http.ResponseWriter, r *http.Request) {
	memberships, err := h.Repo.List(r.Context())
	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	result := make([]dto.Response, len(memberships))
	for i, membership := range memberships {
		result[i] = mapReadValueToResponse(membership)
	}

	responseHandlers.RespondWithSuccess(w, result, http.StatusOK)
}

// UpdateMembership updates an existing membership.
// @Tags memberships
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path string true "Membership ID" example("f47ac10b-58cc-4372-a567-0e02b2c3d479")
// @Param membership body dto.RequestDto true "Membership details"
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

	if err = h.Repo.Update(r.Context(), membershipUpdate); err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	responseHandlers.RespondWithSuccess(w, nil, http.StatusNoContent)
}

// DeleteMembership deletes a membership by ID.
// @Tags memberships
// @Accept json
// @Produce json
// @Param id path string true "Membership ID" example("f47ac10b-58cc-4372-a567-0e02b2c3d479")
// @Security Bearer
// @Success 204 "No Content: Membership deleted successfully"
// @Failure 400 {object} map[string]interface{} "Bad Request: Invalid ID"
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

	if err = h.Repo.Delete(r.Context(), id); err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	responseHandlers.RespondWithSuccess(w, nil, http.StatusNoContent)
}

func mapReadValueToResponse(membership values.ReadValues) dto.Response {
	return dto.Response{
		ID:          membership.ID,
		Name:        membership.Name,
		Description: membership.Description,
		Benefits:    membership.Benefits,
		CreatedAt:   membership.CreatedAt,
		UpdatedAt:   membership.UpdatedAt,
	}
}
