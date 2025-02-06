package controller

import (
	"api/internal/di"
	"api/internal/domains/facility/dto"
	entity "api/internal/domains/facility/entities"
	service "api/internal/domains/facility/services"
	response_handlers "api/internal/libs/responses"
	"api/internal/libs/validators"
	"net/http"

	"github.com/go-chi/chi"
)

type FacilitiesController struct {
	Service *service.FacilityService
}

func NewFacilitiesController(container *di.Container) *FacilitiesController {
	return &FacilitiesController{Service: service.NewFacilityService(container)}
}

func (h *FacilitiesController) CreateFacility(w http.ResponseWriter, r *http.Request) {
	var dto dto.FacilityRequestDto

	if err := validators.ParseJSON(r.Body, &dto); err != nil {
		response_handlers.RespondWithError(w, err)
		return
	}

	facilityCreate, err := dto.ToFacilityCreateValueObject()

	if err != nil {
		response_handlers.RespondWithError(w, err)
		return
	}

	if err := h.Service.CreateFacility(r.Context(), facilityCreate); err != nil {
		response_handlers.RespondWithError(w, err)
		return
	}

	response_handlers.RespondWithSuccess(w, nil, http.StatusCreated)
}

func (h *FacilitiesController) GetFacilityById(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := validators.ParseUUID(idStr)

	if err != nil {
		response_handlers.RespondWithError(w, err)
		return
	}

	facility, err := h.Service.GetFacilityById(r.Context(), id)
	if err != nil {
		response_handlers.RespondWithError(w, err)
		return
	}

	response := mapFacilityEntityToResponse(facility)

	response_handlers.RespondWithSuccess(w, response, http.StatusOK)
}

func (h *FacilitiesController) GetFacilities(w http.ResponseWriter, r *http.Request) {

	name := r.URL.Query().Get("name")

	facilities, err := h.Service.GetFacilities(r.Context(), name)
	if err != nil {
		response_handlers.RespondWithError(w, err)
		return
	}

	result := make([]dto.FacilityResponse, len(facilities))
	for i, facility := range facilities {
		result[i] = mapFacilityEntityToResponse(&facility)
	}

	response_handlers.RespondWithSuccess(w, result, http.StatusOK)
}

func (h *FacilitiesController) UpdateFacility(w http.ResponseWriter, r *http.Request) {

	idStr := chi.URLParam(r, "id")

	var dto dto.FacilityRequestDto

	if err := validators.ParseJSON(r.Body, &dto); err != nil {
		response_handlers.RespondWithError(w, err)
		return
	}

	facilityUpdate, err := dto.ToFacilityUpdateValueObject(idStr)

	if err != nil {
		response_handlers.RespondWithError(w, err)
	}

	if err := h.Service.UpdateFacility(r.Context(), facilityUpdate); err != nil {
		response_handlers.RespondWithError(w, err)
		return
	}

	response_handlers.RespondWithSuccess(w, nil, http.StatusNoContent)
}

func (h *FacilitiesController) DeleteFacility(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := validators.ParseUUID(idStr)

	if err != nil {
		response_handlers.RespondWithError(w, err)
		return
	}

	if err = h.Service.DeleteFacility(r.Context(), id); err != nil {
		response_handlers.RespondWithError(w, err)
		return
	}

	response_handlers.RespondWithSuccess(w, nil, http.StatusNoContent)
}

func mapFacilityEntityToResponse(facility *entity.Facility) dto.FacilityResponse {
	return dto.FacilityResponse{
		ID:           facility.ID,
		Name:         facility.Name,
		Location:     facility.Location,
		FacilityType: facility.FacilityType,
	}
}
