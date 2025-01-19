package facilities

import (
	dto "api/internal/dtos/facility"
	"api/internal/repositories"
	"api/internal/utils"
	"api/internal/utils/validators"
	"net/http"

	"github.com/go-chi/chi"
)

type FacilitiesController struct {
	Repo *repositories.FacilityRepository
}

func NewFacilitiesController(FacilitiesRepository *repositories.FacilityRepository) *FacilitiesController {
	return &FacilitiesController{Repo: FacilitiesRepository}
}

func (c *FacilitiesController) CreateFacility(w http.ResponseWriter, r *http.Request) {

	var targetBody dto.CreateFacilityRequest

	if err := validators.DecodeAndValidateRequestBody(r.Body, &targetBody); err != nil {
		utils.RespondWithError(w, err)
		return
	}

	params := targetBody.ToDBParams()

	if err := c.Repo.CreateFacility(r.Context(), params); err != nil {
		utils.RespondWithError(w, err)
		return
	}

	utils.RespondWithSuccess(w, nil, http.StatusCreated)
}

func (c *FacilitiesController) GetFacility(w http.ResponseWriter, r *http.Request) {
	id, err := validators.ParseUUID(chi.URLParam(r, "id"))
	if err != nil {
		utils.RespondWithError(w, err)
	}

	facility, err := c.Repo.GetFacility(r.Context(), id)

	if err != nil {
		utils.RespondWithError(w, err)
		return
	}

	utils.RespondWithSuccess(w, facility, http.StatusOK)
}

func (c *FacilitiesController) GetAllFacilities(w http.ResponseWriter, r *http.Request) {
	facilities, err := c.Repo.GetAllFacilities(r.Context())

	if err != nil {
		utils.RespondWithError(w, err)
		return
	}

	utils.RespondWithSuccess(w, facilities, http.StatusOK)
}

func (c *FacilitiesController) UpdateFacility(w http.ResponseWriter, r *http.Request) {

	var targetBody dto.UpdateFacilityRequest

	if err := validators.DecodeAndValidateRequestBody(r.Body, &targetBody); err != nil {
		utils.RespondWithError(w, err)
		return
	}

	params := targetBody.ToDBParams()

	if err := c.Repo.UpdateFacility(r.Context(), params); err != nil {

		utils.RespondWithError(w, err)
		return
	}

	utils.RespondWithSuccess(w, nil, http.StatusOK)

}

func (c *FacilitiesController) DeleteFacility(w http.ResponseWriter, r *http.Request) {
	id, err := validators.ParseUUID(chi.URLParam(r, "id"))
	if err != nil {
		utils.RespondWithError(w, err)
		return
	}

	if err = c.Repo.DeleteFacility(r.Context(), id); err != nil {
		utils.RespondWithError(w, err)
		return
	}

	utils.RespondWithSuccess(w, nil, http.StatusNoContent)
}
