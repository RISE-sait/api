package facility

import (
	"api/internal/domains/facility"
	"api/internal/domains/facility/dto"
	response_handlers "api/internal/libs/responses"
	"api/internal/libs/validators"
	"net/http"

	"github.com/go-chi/chi"
)

type Handler struct {
	FacilityService *facility.Service
}

func NewHandler(courseService *facility.Service) *Handler {
	return &Handler{FacilityService: courseService}
}

func (h *Handler) CreateFacility(w http.ResponseWriter, r *http.Request) {
	var targetBody dto.CreateFacilityRequest

	if err := validators.ParseRequestBodyToJSON(r.Body, &targetBody); err != nil {
		response_handlers.RespondWithError(w, err)
		return
	}

	if err := h.FacilityService.CreateFacility(r.Context(), targetBody); err != nil {
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
	courses, err := h.FacilityService.GetAllFacilities(r.Context())
	if err != nil {
		response_handlers.RespondWithError(w, err)
		return
	}

	response_handlers.RespondWithSuccess(w, *courses, http.StatusOK)
}

func (h *Handler) UpdateFacility(w http.ResponseWriter, r *http.Request) {
	var targetBody dto.UpdateFacilityRequest

	if err := validators.ParseRequestBodyToJSON(r.Body, &targetBody); err != nil {
		response_handlers.RespondWithError(w, err)
		return
	}

	if err := h.FacilityService.UpdateFacility(r.Context(), targetBody); err != nil {
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
