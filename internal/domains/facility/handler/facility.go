package facility

import (
	"api/internal/di"
	dto "api/internal/domains/facility/dto"
	"api/internal/domains/facility/service"
	responseHandlers "api/internal/libs/responses"
	"api/internal/libs/validators"
	"net/http"

	"github.com/go-chi/chi"
)

type FacilitiesHandler struct {
	Service *service.FacilityService
}

func NewFacilitiesHandler(container *di.Container) *FacilitiesHandler {
	return &FacilitiesHandler{Service: service.NewFacilityService(container)}
}

// CreateFacility creates a new facility.
// @Summary Create a new facility
// @Description Registers a new facility with the provided details.
// @Tags facilities
// @Accept json
// @Produce json
// @Param facility body dto.RequestDto true "Facility details"
// @Success 201 {object} dto.ResponseDto "Facility created successfully"
// @Failure 400 {object} map[string]interface{} "Bad Request: Invalid input"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /facilities [post]
func (h *FacilitiesHandler) CreateFacility(w http.ResponseWriter, r *http.Request) {
	var requestDto dto.RequestDto

	if err := validators.ParseJSON(r.Body, &requestDto); err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	details, err := requestDto.ToDetails()

	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	facility, err := h.Service.CreateFacility(r.Context(), details)

	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	responseBody := dto.ResponseDto{
		ID:               facility.ID,
		Name:             facility.Name,
		Address:          facility.Address,
		FacilityCategory: facility.FacilityCategoryName,
	}

	responseHandlers.RespondWithSuccess(w, responseBody, http.StatusCreated)
}

// GetFacilityById retrieves a facility by HubSpotId.
// @Summary Get a facility by HubSpotId
// @Description Retrieves a facility by its HubSpotId.
// @Tags facilities
// @Accept json
// @Produce json
// @Param id path string true "Facility HubSpotId"
// @Success 200 {object} dto.ResponseDto "Facility retrieved successfully"
// @Failure 400 {object} map[string]interface{} "Bad Request: Invalid HubSpotId"
// @Failure 404 {object} map[string]interface{} "Not Found: Facility not found"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /facilities/{id} [get]
func (h *FacilitiesHandler) GetFacilityById(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := validators.ParseUUID(idStr)

	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	facility, err := h.Service.GetFacilityById(r.Context(), id)
	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	response := dto.NewFacilityResponse(*facility)

	responseHandlers.RespondWithSuccess(w, response, http.StatusOK)
}

// GetFacilities retrieves all facilities with optional filtering by name.
// @Summary Get all facilities
// @Description Retrieves a list of all facilities, optionally filtered by name.
// @Tags facilities
// @Accept json
// @Produce json
// @Param name query string false "Facility name filter"
// @Success 200 {array} dto.ResponseDto "GetMemberships of facilities retrieved successfully"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /facilities [get]
func (h *FacilitiesHandler) GetFacilities(w http.ResponseWriter, r *http.Request) {

	name := r.URL.Query().Get("name")

	facilities, err := h.Service.GetFacilities(r.Context(), name)
	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	result := make([]dto.ResponseDto, len(facilities))
	for i, facility := range facilities {
		result[i] = dto.NewFacilityResponse(facility)
	}

	responseHandlers.RespondWithSuccess(w, result, http.StatusOK)
}

// UpdateFacility updates an existing facility by HubSpotId.
// @Summary Update a facility
// @Description Updates the details of an existing facility by its HubSpotId.
// @Tags facilities
// @Accept json
// @Produce json
// @Param id path string true "Facility HubSpotId"
// @Param facility body dto.RequestDto true "Updated facility details"
// @Success 204 {object} map[string]interface{} "No Content: Facility updated successfully"
// @Failure 400 {object} map[string]interface{} "Bad Request: Invalid input"
// @Failure 404 {object} map[string]interface{} "Not Found: Facility not found"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /facilities/{id} [put]
func (h *FacilitiesHandler) UpdateFacility(w http.ResponseWriter, r *http.Request) {

	idStr := chi.URLParam(r, "id")

	var requestDto dto.RequestDto

	if err := validators.ParseJSON(r.Body, &requestDto); err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	facilityUpdate, err := requestDto.ToEntity(idStr)

	if err != nil {
		responseHandlers.RespondWithError(w, err)
	}

	if err := h.Service.UpdateFacility(r.Context(), facilityUpdate); err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	responseHandlers.RespondWithSuccess(w, nil, http.StatusNoContent)
}

// DeleteFacility deletes a facility by HubSpotId.
// @Summary Delete a facility
// @Description Deletes a facility by its HubSpotId.
// @Tags facilities
// @Accept json
// @Produce json
// @Param id path string true "Facility HubSpotId"
// @Success 204 {object} map[string]interface{} "No Content: Facility deleted successfully"
// @Failure 400 {object} map[string]interface{} "Bad Request: Invalid HubSpotId"
// @Failure 404 {object} map[string]interface{} "Not Found: Facility not found"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /facilities/{id} [delete]
func (h *FacilitiesHandler) DeleteFacility(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := validators.ParseUUID(idStr)

	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	if err = h.Service.DeleteFacility(r.Context(), id); err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	responseHandlers.RespondWithSuccess(w, nil, http.StatusNoContent)
}
