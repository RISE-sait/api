package location

import (
	"api/internal/di"
	dto "api/internal/domains/location/dto"
	service "api/internal/domains/location/services"
	responseHandlers "api/internal/libs/responses"
	"api/internal/libs/validators"
	"net/http"

	"github.com/go-chi/chi"
)

type Handler struct {
	Service *service.Service
}

func NewLocationsHandler(container *di.Container) *Handler {
	return &Handler{Service: service.NewService(container)}
}

// CreateLocation creates a new Location.
// @Tags locations
// @Accept json
// @Produce json
// @Param body body dto.RequestDto true "Location details"
// @Success 201 {object} dto.ResponseDto "Location created successfully"
// @Failure 400 {object} map[string]interface{} "Bad Request: Invalid input"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /locations [post]
func (h *Handler) CreateLocation(w http.ResponseWriter, r *http.Request) {
	var requestDto dto.RequestDto

	if err := validators.ParseJSON(r.Body, &requestDto); err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	details, err := requestDto.ToCreateDetails()

	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	location, err := h.Service.CreateLocation(r.Context(), details)

	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	responseBody := dto.ResponseDto{
		ID:      location.ID,
		Name:    location.Name,
		Address: location.Address,
	}

	responseHandlers.RespondWithSuccess(w, responseBody, http.StatusCreated)
}

// GetLocationById retrieves a Location by its UUID.
// @Tags locations
// @Accept json
// @Produce json
// @Param id path string true "Location UUID"
// @Success 200 {object} dto.ResponseDto "Location retrieved successfully"
// @Failure 400 {object} map[string]interface{} "Bad Request: Invalid UUID"
// @Failure 404 {object} map[string]interface{} "Not Found: Location not found"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /locations/{id} [get]
func (h *Handler) GetLocationById(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := validators.ParseUUID(idStr)

	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	Location, err := h.Service.GetLocation(r.Context(), id)
	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	response := dto.NewLocationResponse(Location)

	responseHandlers.RespondWithSuccess(w, response, http.StatusOK)
}

// GetLocations retrieves all locations.
// @Tags locations
// @Accept json
// @Produce json
// @Success 200 {array} dto.ResponseDto "List of locations retrieved successfully"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /locations [get]
func (h *Handler) GetLocations(w http.ResponseWriter, r *http.Request) {

	facilities, err := h.Service.GetLocations(r.Context())
	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	result := make([]dto.ResponseDto, len(facilities))
	for i, Location := range facilities {
		result[i] = dto.NewLocationResponse(Location)
	}

	responseHandlers.RespondWithSuccess(w, result, http.StatusOK)
}

// UpdateLocation updates an existing Location by its UUID.
// @Tags locations
// @Accept json
// @Produce json
// @Param id path string true "Location UUID"
// @Param body body dto.RequestDto true "Updated Location details"
// @Success 204 "No Content: Location updated successfully"
// @Failure 400 {object} map[string]interface{} "Bad Request: Invalid input"
// @Failure 404 {object} map[string]interface{} "Not Found: Location not found"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /locations/{id} [put]
func (h *Handler) UpdateLocation(w http.ResponseWriter, r *http.Request) {

	idStr := chi.URLParam(r, "id")

	var requestDto dto.RequestDto

	if err := validators.ParseJSON(r.Body, &requestDto); err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	LocationUpdate, err := requestDto.ToUpdateDetails(idStr)

	if err != nil {
		responseHandlers.RespondWithError(w, err)
	}

	if err = h.Service.UpdateLocation(r.Context(), LocationUpdate); err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	responseHandlers.RespondWithSuccess(w, nil, http.StatusNoContent)
}

// DeleteLocation deletes a Location by its UUID.
// @Tags locations
// @Accept json
// @Produce json
// @Param id path string true "Location UUID"
// @Success 204 "No Content: Location deleted successfully"
// @Failure 400 {object} map[string]interface{} "Bad Request: Invalid UUID"
// @Failure 404 {object} map[string]interface{} "Not Found: Location not found"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /locations/{id} [delete]
func (h *Handler) DeleteLocation(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := validators.ParseUUID(idStr)

	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	if err = h.Service.DeleteLocation(r.Context(), id); err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	responseHandlers.RespondWithSuccess(w, nil, http.StatusNoContent)
}
