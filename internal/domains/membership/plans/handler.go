package membership_plan

import (
	"api/internal/domains/membership/plans/dto"
	response_handlers "api/internal/libs/responses"
	"api/internal/libs/validators"
	"net/http"
)

type Handler struct {
	MembershipPlansService *Service
}

func NewHandler(membershipPlansRepository *Service) *Handler {
	return &Handler{MembershipPlansService: membershipPlansRepository}
}

func (c *Handler) CreateMembershipPlan(w http.ResponseWriter, r *http.Request) {

	var targetBody dto.CreateMembershipPlanRequest

	if err := validators.ParseRequestBodyToJSON(r.Body, &targetBody); err != nil {
		response_handlers.RespondWithError(w, err)
		return
	}

	if err := c.MembershipPlansService.CreateMembershipPlan(r.Context(), targetBody); err != nil {
		response_handlers.RespondWithError(w, err)
		return
	}

	response_handlers.RespondWithSuccess(w, nil, http.StatusCreated)
}

func (c *Handler) GetMembershipPlanDetails(w http.ResponseWriter, r *http.Request) {
	membershipIDStr := r.URL.Query().Get("membership_id")

	membershipID, err := validators.ParseUUID(membershipIDStr)

	if err != nil {
		response_handlers.RespondWithError(w, err)
		return
	}

	plans, err := c.MembershipPlansService.GetPlansMembershipById(r.Context(), membershipID)

	if err != nil {
		response_handlers.RespondWithError(w, err)
		return
	}

	response_handlers.RespondWithSuccess(w, *plans, http.StatusOK)
}

func (c *Handler) UpdateMembershipPlan(w http.ResponseWriter, r *http.Request) {

	var targetBody dto.UpdateMembershipPlanRequest

	if err := validators.ParseRequestBodyToJSON(r.Body, &targetBody); err != nil {
		response_handlers.RespondWithError(w, err)
		return
	}

	if err := c.MembershipPlansService.UpdatePlan(r.Context(), targetBody); err != nil {
		response_handlers.RespondWithError(w, err)
		return
	}

	response_handlers.RespondWithSuccess(w, nil, http.StatusNoContent)
}

func (c *Handler) DeleteMembershipPlan(w http.ResponseWriter, r *http.Request) {
	membershipIDStr := r.URL.Query().Get("membership_id")
	planIDStr := r.URL.Query().Get("id")

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

	if err := c.MembershipPlansService.DeletePlan(r.Context(), membershipID, planId); err != nil {
		response_handlers.RespondWithError(w, err)
		return
	}

	response_handlers.RespondWithSuccess(w, nil, http.StatusNoContent)
}
