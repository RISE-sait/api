package facility_types

import (
	facility_types "api/internal/domains/facilityTypes"
	dto "api/internal/shared/dto/facilityType"
	"api/internal/utils"
	"api/internal/utils/validators"
	"net/http"

	"github.com/go-chi/chi"
)

type FacilityTypesHandler struct {
	Service *facility_types.Service
}

func NewFacilityTypesHandler(service *facility_types.Service) *FacilityTypesHandler {
	return &FacilityTypesHandler{Service: service}
}

func (h *FacilityTypesHandler) GetFacilityTypeByID(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	parsedID, err := validators.ParseUUID(idStr)

	if err != nil {
		http.Error(w, err.Message, err.StatusCode)
		return
	}

	facilityType, httpErr := h.Service.GetFacilityTypeByID(r.Context(), parsedID)
	if httpErr != nil {
		utils.RespondWithError(w, httpErr)
		return
	}

	utils.RespondWithSuccess(w, facilityType, http.StatusOK)
}

func (h *FacilityTypesHandler) GetAllFacilityTypes(w http.ResponseWriter, r *http.Request) {
	facilityTypes, httpErr := h.Service.GetAllFacilityTypes(r.Context())
	if httpErr != nil {
		utils.RespondWithError(w, httpErr)
		return
	}

	utils.RespondWithSuccess(w, facilityTypes, http.StatusOK)
}

func (h *FacilityTypesHandler) CreateFacilityType(w http.ResponseWriter, r *http.Request) {
	var targetBody struct {
		Name string `json:"name" validate:"required,notwhitespace"`
	}

	if err := validators.ParseReqBodyToJSON(r.Body, &targetBody); err != nil {
		utils.RespondWithError(w, err)
		return
	}

	httpErr := h.Service.CreateFacilityType(r.Context(), targetBody.Name)
	if httpErr != nil {
		utils.RespondWithError(w, httpErr)
		return
	}

	utils.RespondWithSuccess(w, nil, http.StatusCreated)
}

func (h *FacilityTypesHandler) UpdateFacilityType(w http.ResponseWriter, r *http.Request) {
	var targetBody dto.UpdateFacilityTypeRequest

	if err := validators.ParseReqBodyToJSON(r.Body, &targetBody); err != nil {
		utils.RespondWithError(w, err)
		return
	}

	httpErr := h.Service.UpdateFacilityType(r.Context(), targetBody)
	if httpErr != nil {
		utils.RespondWithError(w, httpErr)
		return
	}

	utils.RespondWithSuccess(w, nil, http.StatusOK)
}

func (h *FacilityTypesHandler) DeleteFacilityType(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	parsedID, err := validators.ParseUUID(idStr)

	if err != nil {
		utils.RespondWithError(w, err)
		return
	}

	httpErr := h.Service.DeleteFacilityType(r.Context(), parsedID)
	if httpErr != nil {
		utils.RespondWithError(w, httpErr)
		return
	}

	utils.RespondWithSuccess(w, nil, http.StatusNoContent)
}
