package membership

import (
	"api/internal/di"
	dto "api/internal/domains/membership/dto"
	service "api/internal/domains/membership/services"
	values "api/internal/domains/membership/values/plans"

	response_handlers "api/internal/libs/responses"
	"api/internal/libs/validators"
	"net/http"

	"github.com/go-chi/chi"
)

type MembershipPlansController struct {
	MembershipPlansService *service.MembershipPlansService
}

func NewMembershipPlansController(container *di.Container) *MembershipPlansController {
	return &MembershipPlansController{MembershipPlansService: service.NewMembershipPlansService(container)}
}

// CreateMembershipPlan creates a new membership plan.
// @Summary Create a new membership plan
// @Description Create a new membership plan
// @Tags membership-plans
// @Accept json
// @Produce json
// @Param plan body dto.MembershipPlanRequestDto true "Membership plan details"
// @Security Bearer
// @Success 201 "Membership plan created successfully"
// @Failure 400 {object} map[string]interface{} "Bad Request: Invalid input"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /api/memberships/{membershipId}/plans [post]
func (c *MembershipPlansController) CreateMembershipPlan(w http.ResponseWriter, r *http.Request) {

	var requestDto dto.MembershipPlanRequestDto

	if err := validators.ParseJSON(r.Body, &requestDto); err != nil {
		response_handlers.RespondWithError(w, err)
		return
	}

	plan, err := requestDto.ToCreateValueObjects()

	if err != nil {
		response_handlers.RespondWithError(w, err)
		return
	}

	if err := c.MembershipPlansService.CreateMembershipPlan(r.Context(), plan); err != nil {
		response_handlers.RespondWithError(w, err)
		return
	}

	response_handlers.RespondWithSuccess(w, nil, http.StatusCreated)
}

// GetMembershipPlansByMembershipId retrieves membership plans by membership ID.
// @Summary Get membership plans by membership ID
// @Description Get membership plans by membership ID
// @Tags membership-plans
// @Accept json
// @Produce json
// @Param membershipId path string true "Membership ID"
// @Success 200 {array} dto.MembershipPlanResponse "List of membership plans retrieved successfully"
// @Failure 400 {object} map[string]interface{} "Bad Request: Invalid membership ID"
// @Failure 404 {object} map[string]interface{} "Not Found: Membership plans not found"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /api/memberships/{membershipId}/plans [get]
func (c *MembershipPlansController) GetMembershipPlansByMembershipId(w http.ResponseWriter, r *http.Request) {
	membershipIDStr := chi.URLParam(r, "membershipId")

	membershipID, err := validators.ParseUUID(membershipIDStr)

	if err != nil {
		response_handlers.RespondWithError(w, err)
		return
	}

	plans, err := c.MembershipPlansService.GetMembershipPlansByMembershipId(r.Context(), membershipID)

	if err != nil {
		response_handlers.RespondWithError(w, err)
		return
	}

	responseBody := make([]*dto.MembershipPlanResponse, len(plans))

	for i, plan := range plans {
		responseBody[i] = MapEntityToResponse(&plan)
	}

	response_handlers.RespondWithSuccess(w, responseBody, http.StatusOK)
}

// UpdateMembershipPlan updates an existing membership plan.
// @Summary Update a membership plan
// @Description Update a membership plan
// @Tags membership-plans
// @Accept json
// @Produce json
// @Param membershipId path string true "Membership ID"
// @Param planId path string true "Plan ID"
// @Param plan body dto.MembershipPlanRequestDto true "Membership plan details"
// @Security Bearer
// @Success 204 "No Content: Membership plan updated successfully"
// @Failure 400 {object} map[string]interface{} "Bad Request: Invalid input"
// @Failure 404 {object} map[string]interface{} "Not Found: Membership plan not found"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /api/memberships/{membershipId}/plans/{planId} [put]
func (c *MembershipPlansController) UpdateMembershipPlan(w http.ResponseWriter, r *http.Request) {

	membershipIdStr := chi.URLParam(r, "membershipId")
	planIdStr := chi.URLParam(r, "planId")

	var requestDto dto.MembershipPlanRequestDto

	if err := validators.ParseJSON(r.Body, &requestDto); err != nil {
		response_handlers.RespondWithError(w, err)
		return
	}

	plan, err := requestDto.ToUpdateValueObjects(membershipIdStr, planIdStr)

	if err != nil {
		response_handlers.RespondWithError(w, err)
		return
	}

	if err := c.MembershipPlansService.UpdateMembershipPlan(r.Context(), plan); err != nil {
		response_handlers.RespondWithError(w, err)
		return
	}

	response_handlers.RespondWithSuccess(w, nil, http.StatusNoContent)
}

// DeleteMembershipPlan deletes a membership plan by ID.
// @Summary Delete a membership plan
// @Description Delete a membership plan by ID
// @Tags membership-plans
// @Accept json
// @Produce json
// @Param membershipId path string true "Membership ID"
// @Param planId path string true "Plan ID"
// @Security Bearer
// @Success 204 "No Content: Membership plan deleted successfully"
// @Failure 400 {object} map[string]interface{} "Bad Request: Invalid ID"
// @Failure 404 {object} map[string]interface{} "Not Found: Membership plan not found"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /api/memberships/{membershipId}/plans/{planId} [delete]
func (c *MembershipPlansController) DeleteMembershipPlan(w http.ResponseWriter, r *http.Request) {
	membershipIDStr := chi.URLParam(r, "membershipId")
	planIDStr := chi.URLParam(r, "planId")

	membershipID, membershipErr := validators.ParseUUID(membershipIDStr)
	planId, planErr := validators.ParseUUID(planIDStr)

	if membershipErr != nil {
		response_handlers.RespondWithError(w, membershipErr)
		return
	}

	if planErr != nil {
		response_handlers.RespondWithError(w, planErr)
		return
	}

	if err := c.MembershipPlansService.DeleteMembershipPlan(r.Context(), membershipID, planId); err != nil {
		response_handlers.RespondWithError(w, err)
		return
	}

	response_handlers.RespondWithSuccess(w, nil, http.StatusNoContent)
}

func MapEntityToResponse(membership *values.MembershipPlanAllFields) *dto.MembershipPlanResponse {
	return &dto.MembershipPlanResponse{
		ID:               membership.ID,
		Name:             membership.Name,
		MembershipID:     membership.MembershipID,
		Price:            membership.Price,
		PaymentFrequency: membership.PaymentFrequency,
		AmtPeriods:       membership.AmtPeriods,
	}
}
