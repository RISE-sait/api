package membershipPlan

import (
	membershipPlan "api/internal/domains/membership/plans/application"
	"api/internal/domains/membership/plans/infra/http/dto"
	"api/internal/domains/membership/plans/values"
	response_handlers "api/internal/libs/responses"
	"api/internal/libs/validators"
	"net/http"
)

type Handler struct {
	MembershipPlansService *membershipPlan.MembershipPlansService
}

func NewHandler(membershipPlansRepository *membershipPlan.MembershipPlansService) *Handler {
	return &Handler{MembershipPlansService: membershipPlansRepository}
}

func (c *Handler) CreateMembershipPlan(w http.ResponseWriter, r *http.Request) {

	var requestDto dto.CreateMembershipPlanRequest

	if err := validators.ParseJSON(r.Body, &requestDto); err != nil {
		response_handlers.RespondWithError(w, err)
		return
	}

	plan := values.NewMembershipPlanCreate(
		requestDto.MembershipID,
		requestDto.Name,
		requestDto.PaymentFrequency,
		requestDto.Price,
		requestDto.AmtPeriods,
	)

	if err := c.MembershipPlansService.CreateMembershipPlan(r.Context(), plan); err != nil {
		response_handlers.RespondWithError(w, err)
		return
	}

	response_handlers.RespondWithSuccess(w, nil, http.StatusCreated)
}

func (c *Handler) GetMembershipPlansByMembershipId(w http.ResponseWriter, r *http.Request) {
	membershipIDStr := r.URL.Query().Get("membership_id")

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

	response_handlers.RespondWithSuccess(w, plans, http.StatusOK)
}

func (c *Handler) UpdateMembershipPlan(w http.ResponseWriter, r *http.Request) {

	var requestDto dto.UpdateMembershipPlanRequest

	if err := validators.ParseJSON(r.Body, &requestDto); err != nil {
		response_handlers.RespondWithError(w, err)
		return
	}

	plan := values.NewMembershipPlanUpdate(
		requestDto.ID,
		requestDto.MembershipID,
		requestDto.Name,
		requestDto.PaymentFrequency,
		requestDto.Price,
		requestDto.AmtPeriods,
	)

	if err := c.MembershipPlansService.UpdateMembershipPlan(r.Context(), plan); err != nil {
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

	if err := c.MembershipPlansService.DeleteMembershipPlan(r.Context(), membershipID, planId); err != nil {
		response_handlers.RespondWithError(w, err)
		return
	}

	response_handlers.RespondWithSuccess(w, nil, http.StatusNoContent)
}
