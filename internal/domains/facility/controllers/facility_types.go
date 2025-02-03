package controller

import (
	"api/cmd/server/di"
	"api/internal/domains/facility/dto"
	entity "api/internal/domains/facility/entities"
	service "api/internal/domains/facility/services"
	response_handlers "api/internal/libs/responses"
	"api/internal/libs/validators"
	"net/http"

	"github.com/go-chi/chi"
)

type FacilityTypesController struct {
	Service *service.FacilityTypesService
}

func NewFacilityTypesController(container *di.Container) *FacilityTypesController {
	return &FacilityTypesController{Service: service.NewFacilityTypesService(container)}
}

func (c *FacilityTypesController) CreateFacility(w http.ResponseWriter, r *http.Request) {
	var dto dto.FacilityTypeRequestDto

	if err := validators.ParseJSON(r.Body, &dto); err != nil {
		response_handlers.RespondWithError(w, err)
		return
	}

	if err := validators.ValidateDto(&dto); err != nil {
		response_handlers.RespondWithError(w, err)
	}

	if err := c.Service.CreateFacility(r.Context(), dto.Name); err != nil {
		response_handlers.RespondWithError(w, err)
		return
	}

	response_handlers.RespondWithSuccess(w, nil, http.StatusCreated)
}

func (c *FacilityTypesController) GetFacilityById(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := validators.ParseUUID(idStr)

	if err != nil {
		response_handlers.RespondWithError(w, err)
		return
	}

	name, err := c.Service.GetFacilityTypeById(r.Context(), id)
	if err != nil {
		response_handlers.RespondWithError(w, err)
		return
	}

	facilityType := entity.FacilityType{
		ID:   id,
		Name: *name,
	}

	response := mapFacilityTypeEntityToResponse(&facilityType)

	response_handlers.RespondWithSuccess(w, response, http.StatusOK)
}

func (c *FacilityTypesController) GetAllFacilityTypes(w http.ResponseWriter, r *http.Request) {
	facilityTypes, err := c.Service.GetAllFacilityTypes(r.Context())
	if err != nil {
		response_handlers.RespondWithError(w, err)
		return
	}

	result := make([]dto.FacilityTypeResponse, len(facilityTypes))
	for i, facilityType := range facilityTypes {
		result[i] = mapFacilityTypeEntityToResponse(&facilityType)
	}

	response_handlers.RespondWithSuccess(w, result, http.StatusOK)
}

func (c *FacilityTypesController) UpdateFacility(w http.ResponseWriter, r *http.Request) {

	idStr := chi.URLParam(r, "id")

	var dto dto.FacilityTypeRequestDto

	if err := validators.ParseJSON(r.Body, &dto); err != nil {
		response_handlers.RespondWithError(w, err)
		return
	}

	facilityTypeVo, err := dto.ToFacilityTypeUpdateValueObject(idStr)

	if err != nil {
		response_handlers.RespondWithError(w, err)
		return
	}

	if err := c.Service.Repo.UpdateFacilityType(r.Context(), facilityTypeVo); err != nil {
		response_handlers.RespondWithError(w, err)
		return
	}

	response_handlers.RespondWithSuccess(w, nil, http.StatusNoContent)
}

func (c *FacilityTypesController) DeleteFacilityType(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := validators.ParseUUID(idStr)

	if err != nil {
		response_handlers.RespondWithError(w, err)
		return
	}

	if err = c.Service.DeleteFacilityType(r.Context(), id); err != nil {
		response_handlers.RespondWithError(w, err)
		return
	}

	response_handlers.RespondWithSuccess(w, nil, http.StatusNoContent)
}

func mapFacilityTypeEntityToResponse(facility *entity.FacilityType) dto.FacilityTypeResponse {
	return dto.FacilityTypeResponse{
		ID:   facility.ID,
		Name: facility.Name,
	}
}
