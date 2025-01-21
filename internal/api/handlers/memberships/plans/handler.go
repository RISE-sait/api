package membership_plans

import (
	membership_plans "api/internal/domains/membershipPlans"
	"api/internal/utils"
	"api/internal/utils/validators"
	"net/http"

	dto "api/internal/shared/dto/membershipPlans"
)

type Handler struct {
	Service *membership_plans.Service
}

func NewHandler(service *membership_plans.Service) *Handler {
	return &Handler{Service: service}
}

func (h *Handler) CreateMembershipPlan(w http.ResponseWriter, r *http.Request) {
	var targetBody dto.CreateMembershipPlanRequest

	if err := validators.ParseReqBodyToJSON(r.Body, &targetBody); err != nil {
		utils.RespondWithError(w, err)
		return
	}

	httpErr := h.Service.CreateMembershipPlan(r.Context(), targetBody)
	if httpErr != nil {
		utils.RespondWithError(w, httpErr)
		return
	}

	utils.RespondWithSuccess(w, nil, http.StatusCreated)
}

func (h *Handler) GetMembershipPlanDetails(w http.ResponseWriter, r *http.Request) {
	membershipIDStr := r.URL.Query().Get("membership_id")
	planIDStr := r.URL.Query().Get("plan_id")

	membershipID, err := validators.ParseUUID(membershipIDStr)
	if err != nil {
		utils.RespondWithError(w, err)
		return
	}

	planID, err := validators.ParseUUID(planIDStr)
	if err != nil {
		utils.RespondWithError(w, err)
		return
	}

	plans, httpErr := h.Service.GetMembershipPlanDetails(r.Context(), membershipID, planID)
	if httpErr != nil {
		utils.RespondWithError(w, httpErr)
		return
	}

	utils.RespondWithSuccess(w, plans, http.StatusOK)
}

func (h *Handler) UpdateMembershipPlan(w http.ResponseWriter, r *http.Request) {
	var targetBody dto.UpdateMembershipPlanRequest

	if err := validators.ParseReqBodyToJSON(r.Body, &targetBody); err != nil {
		utils.RespondWithError(w, err)
		return
	}

	httpErr := h.Service.UpdateMembershipPlan(r.Context(), targetBody)
	if httpErr != nil {
		utils.RespondWithError(w, httpErr)
		return
	}

	utils.RespondWithSuccess(w, nil, http.StatusNoContent)
}

func (h *Handler) DeleteMembershipPlan(w http.ResponseWriter, r *http.Request) {
	membershipIDStr := r.URL.Query().Get("membership_id")
	planIDStr := r.URL.Query().Get("id")

	membershipID, err := validators.ParseUUID(membershipIDStr)
	if err != nil {
		utils.RespondWithError(w, err)
		return
	}

	planID, err := validators.ParseUUID(planIDStr)
	if err != nil {
		utils.RespondWithError(w, err)
		return
	}

	httpErr := h.Service.DeleteMembershipPlan(r.Context(), membershipID, planID)
	if httpErr != nil {
		utils.RespondWithError(w, httpErr)
		return
	}

	utils.RespondWithSuccess(w, nil, http.StatusNoContent)
}
