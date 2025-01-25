package facility

import (
	facility "api/internal/domains/facility/application"
	"api/internal/domains/facility/infra/http/dto"
	"api/internal/domains/facility/infra/http/mapper"
	"api/internal/domains/facility/values"
	response_handlers "api/internal/libs/responses"
	"api/internal/libs/validators"
	"net/http"

	"github.com/go-chi/chi"
)

type Handler struct {
	FacilityService *facility.FacilityService
}

func NewHandler(courseService *facility.FacilityService) *Handler {
	return &Handler{FacilityService: courseService}
}

func (h *Handler) CreateFacility(w http.ResponseWriter, r *http.Request) {
	var requestDto dto.CreateFacilityRequest

	if err := validators.ParseJSON(r.Body, &requestDto); err != nil {
		response_handlers.RespondWithError(w, err)
		return
	}

	facilityCreate := values.NewFacilityCreate(requestDto.Name, requestDto.Location, requestDto.FacilityTypeID)

	if err := h.FacilityService.CreateFacility(r.Context(), facilityCreate); err != nil {
		response_handlers.RespondWithError(w, err)
		return
	}

	response_handlers.RespondWithSuccess(w, nil, http.StatusCreated)
}

func (h *Handler) GetFacilityById(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := validators.ParseUUID(idStr)

	if err != nil {
		response_handlers.RespondWithError(w, err)
		return
	}

	course, err := h.FacilityService.GetFacilityById(r.Context(), id)
	if err != nil {
		response_handlers.RespondWithError(w, err)
		return
	}

	response_handlers.RespondWithSuccess(w, *course, http.StatusOK)
}

func (h *Handler) GetAllFacilities(w http.ResponseWriter, r *http.Request) {
	facilities, err := h.FacilityService.GetAllFacilities(r.Context())
	if err != nil {
		response_handlers.RespondWithError(w, err)
		return
	}

	result := make([]dto.FacilityResponse, len(facilities))
	for i, facility := range facilities {
		result[i] = mapper.MapEntityToResponse(&facility)
	}

	response_handlers.RespondWithSuccess(w, result, http.StatusOK)
}

func (h *Handler) UpdateFacility(w http.ResponseWriter, r *http.Request) {
	var dto dto.UpdateFacilityRequest

	if err := validators.ParseJSON(r.Body, &dto); err != nil {
		response_handlers.RespondWithError(w, err)
		return
	}

	facilityUpdate := values.NewFacilityUpdate(dto.ID, dto.Name, dto.Location, dto.FacilityTypeID)

	if err := h.FacilityService.UpdateFacility(r.Context(), facilityUpdate); err != nil {
		response_handlers.RespondWithError(w, err)
		return
	}

	response_handlers.RespondWithSuccess(w, nil, http.StatusNoContent)
}

func (h *Handler) DeleteFacility(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := validators.ParseUUID(idStr)

	if err != nil {
		response_handlers.RespondWithError(w, err)
		return
	}

	if err = h.FacilityService.DeleteFacility(r.Context(), id); err != nil {
		response_handlers.RespondWithError(w, err)
		return
	}

	response_handlers.RespondWithSuccess(w, nil, http.StatusNoContent)
}
