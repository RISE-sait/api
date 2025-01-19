package memberships

import (
	dto "api/internal/dtos/membershipPlans"
	"api/internal/repositories"
	"api/internal/utils"
	"api/internal/utils/validators"
	db "api/sqlc"
	"net/http"

	"github.com/google/uuid"
)

type MembershipPlansController struct {
	MembershipPlansRepository *repositories.MembershipPlansRepository
}

func NewMembershipPlansController(membershipPlansRepository *repositories.MembershipPlansRepository) *MembershipPlansController {
	return &MembershipPlansController{MembershipPlansRepository: membershipPlansRepository}
}

func (c *MembershipPlansController) CreateMembershipPlan(w http.ResponseWriter, r *http.Request) {

	var targetBody dto.CreateMembershipPlanRequest

	if err := validators.DecodeAndValidateRequestBody(r.Body, &targetBody); err != nil {
		utils.RespondWithError(w, err)
		return
	}

	params := targetBody.ToDBParams()

	if err := c.MembershipPlansRepository.CreateMembershipPlan(r.Context(), params); err != nil {
		utils.RespondWithError(w, err)
		return
	}

	utils.RespondWithSuccess(w, nil, http.StatusCreated)
}

func (c *MembershipPlansController) GetMembershipPlanDetails(w http.ResponseWriter, r *http.Request) {
	membershipIDStr := r.URL.Query().Get("membership_id")
	planIDStr := r.URL.Query().Get("plan_id")

	var membershipID, planID uuid.UUID = uuid.Nil, uuid.Nil
	if membershipIDStr != "" {
		id, err := uuid.Parse(membershipIDStr)
		if err != nil {
			utils.RespondWithError(w, utils.CreateHTTPError("Invalid membership ID format", http.StatusBadRequest))
			return
		}
		membershipID = id
	}

	if planIDStr != "" {
		id, err := uuid.Parse(planIDStr)
		if err != nil {
			utils.RespondWithError(w, utils.CreateHTTPError("Invalid plan ID format", http.StatusBadRequest))
			return
		}
		planID = id
	}

	plans, err := c.MembershipPlansRepository.GetMembershipPlanDetails(r.Context(), membershipID, planID)

	if err != nil {
		utils.RespondWithError(w, err)
		return
	}

	utils.RespondWithSuccess(w, plans, http.StatusOK)
}

func (c *MembershipPlansController) UpdateMembershipPlan(w http.ResponseWriter, r *http.Request) {

	var targetBody dto.UpdateMembershipPlanRequest

	if err := validators.DecodeAndValidateRequestBody(r.Body, &targetBody); err != nil {
		utils.RespondWithError(w, err)
		return
	}

	params := targetBody.ToDBParams()

	if err := c.MembershipPlansRepository.UpdateMembershipPlan(r.Context(), params); err != nil {
		utils.RespondWithError(w, err)
		return
	}

	utils.RespondWithSuccess(w, nil, http.StatusNoContent)
}

func (c *MembershipPlansController) DeleteMembershipPlan(w http.ResponseWriter, r *http.Request) {
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

	params := db.DeleteMembershipPlanParams{
		MembershipID: membershipID,
		ID:           planID,
	}

	if err := c.MembershipPlansRepository.DeleteMembershipPlan(r.Context(), &params); err != nil {
		utils.RespondWithError(w, err)
		return
	}

	utils.RespondWithSuccess(w, nil, http.StatusNoContent)
}
