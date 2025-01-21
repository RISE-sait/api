package facilities

import (
	"api/internal/domains/facilities"
	dto "api/internal/shared/dto/facility"
	"api/internal/utils"
	"api/internal/utils/validators"
	"net/http"

	"github.com/go-chi/chi"
)

type FacilitiesHandler struct {
	Service *facilities.Service
}

func NewFacilitiesHandler(service *facilities.Service) *FacilitiesHandler {
	return &FacilitiesHandler{Service: service}
}

func (h *FacilitiesHandler) CreateFacility(w http.ResponseWriter, r *http.Request) {
	var targetBody dto.CreateFacilityRequest

	if err := validators.ParseReqBodyToJSON(r.Body, &targetBody); err != nil {
		utils.RespondWithError(w, err)
		return
	}

	if err := h.Service.CreateFacility(r.Context(), &targetBody); err != nil {
		utils.RespondWithError(w, err)
		return
	}

	utils.RespondWithSuccess(w, nil, http.StatusCreated)
}

func (h *FacilitiesHandler) GetFacility(w http.ResponseWriter, r *http.Request) {
	id, err := validators.ParseUUID(chi.URLParam(r, "id"))
	if err != nil {
		utils.RespondWithError(w, err)
		return
	}

	facility, err := h.Service.GetFacility(r.Context(), id)
	if err != nil {
		utils.RespondWithError(w, err)
		return
	}

	utils.RespondWithSuccess(w, facility, http.StatusOK)
}

func (h *FacilitiesHandler) GetAllFacilities(w http.ResponseWriter, r *http.Request) {
	facilities, err := h.Service.GetAllFacilities(r.Context())
	if err != nil {
		utils.RespondWithError(w, err)
		return
	}

	utils.RespondWithSuccess(w, facilities, http.StatusOK)
}

func (h *FacilitiesHandler) UpdateFacility(w http.ResponseWriter, r *http.Request) {
	var targetBody dto.UpdateFacilityRequest

	if err := validators.ParseReqBodyToJSON(r.Body, &targetBody); err != nil {
		utils.RespondWithError(w, err)
		return
	}

	if err := h.Service.UpdateFacility(r.Context(), &targetBody); err != nil {
		utils.RespondWithError(w, err)
		return
	}

	utils.RespondWithSuccess(w, nil, http.StatusOK)
}

func (h *FacilitiesHandler) DeleteFacility(w http.ResponseWriter, r *http.Request) {
	id, err := validators.ParseUUID(chi.URLParam(r, "id"))
	if err != nil {
		utils.RespondWithError(w, err)
		return
	}

	if err = h.Service.DeleteFacility(r.Context(), id); err != nil {
		utils.RespondWithError(w, err)
		return
	}

	utils.RespondWithSuccess(w, nil, http.StatusNoContent)
}
