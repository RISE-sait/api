package controller

import (
	membershipPlan "api/internal/domains/membership/plans/application"
	"api/internal/domains/membership/plans/infra/http/dto"
	"api/internal/domains/membership/plans/values"
	response_handlers "api/internal/libs/responses"
	"api/internal/libs/validators"
	"net/http"

	"github.com/go-chi/chi"
)

type MembershipPlansController struct {
	MembershipPlansService *membershipPlan.MembershipPlansService
}

func NewMembershipPlansController(membershipPlansRepository *membershipPlan.MembershipPlansService) *MembershipPlansController {
	return &MembershipPlansController{MembershipPlansService: membershipPlansRepository}
}

func (c *MembershipPlansController) CreateMembershipPlan(w http.ResponseWriter, r *http.Request) {

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

	response_handlers.RespondWithSuccess(w, plans, http.StatusOK)
}

func (c *MembershipPlansController) UpdateMembershipPlan(w http.ResponseWriter, r *http.Request) {

	membershipIdStr := chi.URLParam(r, "membershipId")
	planIdStr := chi.URLParam(r, "planId")

	membershipId, err := validators.ParseUUID(membershipIdStr)
	if err != nil {
		response_handlers.RespondWithError(w, err)
		return
	}

	planId, err := validators.ParseUUID(planIdStr)
	if err != nil {
		response_handlers.RespondWithError(w, err)
		return
	}

	var requestDto dto.UpdateMembershipPlanRequest

	if err := validators.ParseJSON(r.Body, &requestDto); err != nil {
		response_handlers.RespondWithError(w, err)
		return
	}

	plan := values.NewMembershipPlanUpdate(
		planId,
		membershipId,
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
