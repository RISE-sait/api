package facilities

import (
	dto "api/internal/dtos/facilityType"
	"api/internal/repositories"
	"api/internal/utils"
	"api/internal/utils/validators"
	"net/http"

	"github.com/go-chi/chi"
)

type FacilityTypesController struct {
	FacilityTypesRepository *repositories.FacilityTypesRepository
}

func NewFacilityTypesController(facilityTypesRepository *repositories.FacilityTypesRepository) *FacilityTypesController {
	return &FacilityTypesController{FacilityTypesRepository: facilityTypesRepository}
}

func (ctrl *FacilityTypesController) GetFacilityTypeByID(w http.ResponseWriter, r *http.Request) {

	idStr := chi.URLParam(r, "id")

	parsedID, err := validators.ParseUUID(idStr)

	if err != nil {
		http.Error(w, err.Message, err.StatusCode)
		return
	}

	facilityType, err := ctrl.FacilityTypesRepository.GetFacilityType(r.Context(), parsedID)

	if err != nil {
		utils.RespondWithError(w, err)
		return
	}
	utils.RespondWithSuccess(w, facilityType, http.StatusOK)
}

func (ctrl *FacilityTypesController) GetAllFacilityTypes(w http.ResponseWriter, r *http.Request) {
	facilityTypes, err := ctrl.FacilityTypesRepository.GetAllFacilityTypes(r.Context())

	if err != nil {
		utils.RespondWithError(w, err)
		return
	}
	utils.RespondWithSuccess(w, facilityTypes, http.StatusOK)
}

func (ctrl *FacilityTypesController) CreateFacilityType(w http.ResponseWriter, r *http.Request) {
	var targetBody struct {
		Name string `json:"name" validate:"required,notwhitespace"`
	}

	if err := validators.DecodeAndValidateRequestBody(r.Body, &targetBody); err != nil {
		utils.RespondWithError(w, err)
		return
	}

	if err := ctrl.FacilityTypesRepository.CreateFacilityType(r.Context(), targetBody.Name); err != nil {

		utils.RespondWithError(w, err)
		return
	}

	utils.RespondWithSuccess(w, nil, http.StatusCreated)
}

func (ctrl *FacilityTypesController) UpdateFacilityType(w http.ResponseWriter, r *http.Request) {

	var targetBody dto.UpdateFacilityTypeRequest

	if err := validators.DecodeAndValidateRequestBody(r.Body, &targetBody); err != nil {
		utils.RespondWithError(w, err)
		return
	}

	params := targetBody.ToDBParams()

	if err := ctrl.FacilityTypesRepository.UpdateFacilityType(r.Context(), params); err != nil {
		utils.RespondWithError(w, err)
		return
	}

	utils.RespondWithSuccess(w, nil, http.StatusOK)
}

func (ctrl *FacilityTypesController) DeleteFacilityType(w http.ResponseWriter, r *http.Request) {

	idStr := chi.URLParam(r, "id")

	id, err := validators.ParseUUID(idStr)

	if err != nil {
		utils.RespondWithError(w, err)
		return
	}

	if err = ctrl.FacilityTypesRepository.DeleteFacilityType(r.Context(), id); err != nil {
		utils.RespondWithError(w, err)
		return
	}

	utils.RespondWithSuccess(w, nil, http.StatusNoContent)
}
