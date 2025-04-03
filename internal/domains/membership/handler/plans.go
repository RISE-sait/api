package membership

import (
	"api/internal/di"
	"api/internal/domains/membership/dto/membership_plan"
	repo "api/internal/domains/membership/persistence/repositories"
	errLib "api/internal/libs/errors"
	responseHandlers "api/internal/libs/responses"
	"api/internal/libs/validators"
	"github.com/google/uuid"
	"net/http"

	"github.com/go-chi/chi"
)

type PlansHandlers struct {
	Repo *repo.PlansRepository
}

func NewPlansHandlers(container *di.Container) *PlansHandlers {
	return &PlansHandlers{Repo: repo.NewMembershipPlansRepository(container)}
}

// CreateMembershipPlan creates a new membership plan.
// @Summary Create a new membership plan
// @Description Create a new membership plan
// @Tags membership-plans
// @Accept json
// @Produce json
// @Param plan body membership_plan.PlanRequestDto true "Membership plan details"
// @Security Bearer
// @Success 201 "Membership plan created successfully"
// @Failure 400 {object} map[string]interface{} "Bad Request: Invalid input"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /memberships/plans [post]
func (h *PlansHandlers) CreateMembershipPlan(w http.ResponseWriter, r *http.Request) {

	var requestDto membership_plan.PlanRequestDto

	if err := validators.ParseJSON(r.Body, &requestDto); err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	plan, err := requestDto.ToCreateValueObjects()

	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	if err = h.Repo.CreateMembershipPlan(r.Context(), plan); err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	responseHandlers.RespondWithSuccess(w, nil, http.StatusCreated)
}

// GetMembershipPlans retrieves membership plans.
// @Summary Get membership plans by membership ID
// @Description Get membership plans by membership ID
// @Tags membership-plans
// @Accept json
// @Produce json
// @Param id path string true "Membership ID" example("f47ac10b-58cc-4372-a567-0e02b2c3d479")
// @Success 200 {array} membership_plan.PlanResponse "GetMemberships of membership plans retrieved successfully"
// @Failure 400 {object} map[string]interface{} "Bad Request: Invalid membership ID"
// @Failure 404 {object} map[string]interface{} "Not Found: Membership plans not found"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /memberships/{id}/plans [get]
func (h *PlansHandlers) GetMembershipPlans(w http.ResponseWriter, r *http.Request) {

	var membershipId uuid.UUID

	membershipIdStr := chi.URLParam(r, "id")

	membershipId, err := validators.ParseUUID(membershipIdStr)
	if err != nil {
		responseHandlers.RespondWithError(w, errLib.New("Invalid membershipId", http.StatusBadRequest))
		return
	}

	plans, err := h.Repo.GetMembershipPlans(r.Context(), membershipId)

	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	responseBody := make([]membership_plan.PlanResponse, len(plans))

	for i, plan := range plans {
		responseBody[i] = membership_plan.NewPlanResponse(plan)
	}

	responseHandlers.RespondWithSuccess(w, responseBody, http.StatusOK)
}

// UpdateMembershipPlan updates an existing membership plan.
// @Summary Update a membership plan
// @Description Update a membership plan
// @Tags membership-plans
// @Accept json
// @Produce json
// @Param id path string true "Plan ID"
// @Param plan body membership_plan.PlanRequestDto true "Membership plan details"
// @Security Bearer
// @Success 204 "No Content: Membership plan updated successfully"
// @Failure 400 {object} map[string]interface{} "Bad Request: Invalid input"
// @Failure 404 {object} map[string]interface{} "Not Found: Membership plan not found"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /memberships/plans/{id} [put]
func (h *PlansHandlers) UpdateMembershipPlan(w http.ResponseWriter, r *http.Request) {

	idStr := chi.URLParam(r, "id")

	var requestDto membership_plan.PlanRequestDto

	if err := validators.ParseJSON(r.Body, &requestDto); err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	plan, err := requestDto.ToUpdateValueObjects(idStr)

	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	if err = h.Repo.UpdateMembershipPlan(r.Context(), plan); err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	responseHandlers.RespondWithSuccess(w, nil, http.StatusNoContent)
}

// DeleteMembershipPlan deletes a membership plan by ID.
// @Summary Delete a membership plan
// @Description Delete a membership plan by ID
// @Tags membership-plans
// @Accept json
// @Produce json
// @Param id path string true "Plan ID"
// @Security Bearer
// @Success 204 "No Content: Membership plan deleted successfully"
// @Failure 400 {object} map[string]interface{} "Bad Request: Invalid ID"
// @Failure 404 {object} map[string]interface{} "Not Found: Membership plan not found"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /memberships/plans/{id} [delete]
func (h *PlansHandlers) DeleteMembershipPlan(w http.ResponseWriter, r *http.Request) {

	idStr := chi.URLParam(r, "id")

	id, err := validators.ParseUUID(idStr)

	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	if err = h.Repo.DeleteMembershipPlan(r.Context(), id); err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	responseHandlers.RespondWithSuccess(w, nil, http.StatusNoContent)
}
