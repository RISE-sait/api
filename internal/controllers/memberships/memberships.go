package memberships

import (
	dto "api/internal/dtos/membership"
	"api/internal/repositories"
	"api/internal/utils"
	"api/internal/utils/validators"
	"net/http"

	"github.com/go-chi/chi"
)

type MembershipsController struct {
	MembershipsRepository *repositories.MembershipsRepository
}

func NewMembershipsController(membershipsRepository *repositories.MembershipsRepository) *MembershipsController {
	return &MembershipsController{MembershipsRepository: membershipsRepository}
}

func (c *MembershipsController) CreateMembership(w http.ResponseWriter, r *http.Request) {

	var targetBody dto.CreateMembershipRequest

	if err := validators.DecodeAndValidateRequestBody(r.Body, &targetBody); err != nil {
		utils.RespondWithError(w, err)
		return
	}

	params := targetBody.ToDBParams()

	if err := c.MembershipsRepository.CreateMembership(r.Context(), params); err != nil {
		utils.RespondWithError(w, err)
		return
	}

	utils.RespondWithSuccess(w, nil, http.StatusCreated)
}

func (c *MembershipsController) GetMembershipById(w http.ResponseWriter, r *http.Request) {

	idStr := chi.URLParam(r, "id")

	id, err := validators.ParseUUID(idStr)

	if err != nil {
		utils.RespondWithError(w, err)
		return
	}

	membership, err := c.MembershipsRepository.GetMembershipById(r.Context(), id)
	if err != nil {
		utils.RespondWithError(w, err)
		return
	}

	utils.RespondWithSuccess(w, membership, http.StatusOK)
}

func (c *MembershipsController) GetAllMemberships(w http.ResponseWriter, r *http.Request) {
	memberships, err := c.MembershipsRepository.GetAllMemberships(r.Context())
	if err != nil {
		utils.RespondWithError(w, err)
		return
	}

	utils.RespondWithSuccess(w, &memberships, http.StatusOK)
}

func (c *MembershipsController) UpdateMembership(w http.ResponseWriter, r *http.Request) {

	var targetBody dto.UpdateMembershipRequest

	if err := validators.DecodeAndValidateRequestBody(r.Body, &targetBody); err != nil {
		utils.RespondWithError(w, err)
		return
	}

	params := targetBody.ToDBParams()

	if err := c.MembershipsRepository.UpdateMembership(r.Context(), params); err != nil {
		utils.RespondWithError(w, err)
		return
	}

	utils.RespondWithSuccess(w, nil, http.StatusNoContent)
}

func (c *MembershipsController) DeleteMembership(w http.ResponseWriter, r *http.Request) {

	idStr := chi.URLParam(r, "id")

	id, err := validators.ParseUUID(idStr)

	if err != nil {
		utils.RespondWithError(w, err)
		return
	}

	if err := c.MembershipsRepository.DeleteMembership(r.Context(), id); err != nil {
		utils.RespondWithError(w, err)
		return
	}

	utils.RespondWithSuccess(w, nil, http.StatusNoContent)
}
