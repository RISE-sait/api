package membership

import (
	"api/internal/di"
	dto "api/internal/domains/membership/dto"
	entity "api/internal/domains/membership/entities"
	service "api/internal/domains/membership/services"

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

func MapEntityToResponse(membership *entity.MembershipPlan) *dto.MembershipPlanResponse {
	return &dto.MembershipPlanResponse{
		ID:               membership.ID,
		Name:             membership.Name,
		MembershipID:     membership.MembershipID,
		Price:            membership.Price,
		PaymentFrequency: membership.PaymentFrequency,
		AmtPeriods:       membership.AmtPeriods,
	}
}
