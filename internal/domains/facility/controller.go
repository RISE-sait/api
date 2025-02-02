package facility

import (
	"api/cmd/server/di"
	"api/internal/domains/facility/dto"
	entity "api/internal/domains/facility/entities"
	response_handlers "api/internal/libs/responses"
	"api/internal/libs/validators"
	"net/http"

	"github.com/go-chi/chi"
)

type Controller struct {
	Service *FacilityService
}

func NewController(container *di.Container) *Controller {
	return &Controller{Service: NewFacilityService(container)}
}

func (h *Controller) CreateFacility(w http.ResponseWriter, r *http.Request) {
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

func (h *Controller) GetFacilityById(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := validators.ParseUUID(idStr)

	if err != nil {
		response_handlers.RespondWithError(w, err)
		return
	}

	course, err := h.Service.GetFacilityById(r.Context(), id)
	if err != nil {
		response_handlers.RespondWithError(w, err)
		return
	}

	response_handlers.RespondWithSuccess(w, *course, http.StatusOK)
}

func (h *Controller) GetAllFacilities(w http.ResponseWriter, r *http.Request) {
	facilities, err := h.Service.GetAllFacilities(r.Context())
	if err != nil {
		response_handlers.RespondWithError(w, err)
		return
	}

	result := make([]dto.FacilityResponse, len(facilities))
	for i, facility := range facilities {
		result[i] = mapEntityToResponse(&facility)
	}

	response_handlers.RespondWithSuccess(w, result, http.StatusOK)
}

func (h *Controller) UpdateFacility(w http.ResponseWriter, r *http.Request) {

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

func (h *Controller) DeleteFacility(w http.ResponseWriter, r *http.Request) {
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

func mapEntityToResponse(facility *entity.Facility) dto.FacilityResponse {
	return dto.FacilityResponse{
		ID:           facility.ID,
		Name:         facility.Name,
		Location:     facility.Location,
		FacilityType: facility.FacilityType,
	}
}
