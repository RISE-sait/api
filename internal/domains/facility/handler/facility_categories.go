package facility

import (
	"api/internal/di"
	dto "api/internal/domains/facility/dto"
	entity "api/internal/domains/facility/entity"
	"api/internal/domains/facility/service"
	responseHandlers "api/internal/libs/responses"
	"api/internal/libs/validators"
	"net/http"

	"github.com/go-chi/chi"
)

type CategoriesHandler struct {
	Service *service.FacilityCategoriesService
}

func NewFacilityCategoriesHandler(container *di.Container) *CategoriesHandler {
	return &CategoriesHandler{Service: service.NewFacilityCategoriesService(container)}
}

// Create creates a new facility category.
// @Summary Create a new facility category
// @Description Registers a new facility category with the provided name.
// @Tags facility-categories
// @Accept json
// @Produce json
// @Param facility_category body dto.CategoryRequestDto true "Facility Category details"
// @Success 201 {object} map[string]interface{} "Facility category created successfully"
// @Failure 400 {object} map[string]interface{} "Bad Request: Invalid input"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /facilities/categories [post]
func (h *CategoriesHandler) Create(w http.ResponseWriter, r *http.Request) {
	var requestDto dto.CategoryRequestDto

	if err := validators.ParseJSON(r.Body, &requestDto); err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	facilityType, err := requestDto.ToCreateFacilityCategoryValueObject()

	if err != nil {
		responseHandlers.RespondWithError(w, err)
	}

	if err := h.Service.Create(r.Context(), *facilityType); err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	responseHandlers.RespondWithSuccess(w, nil, http.StatusCreated)
}

// GetById retrieves a facility category by ID.
// @Summary Get a facility category by ID
// @Description Retrieves a facility category by its ID.
// @Tags facility-categories
// @Accept json
// @Produce json
// @Param id path string true "Facility Category ID"
// @Success 200 {object} facility.CategoryResponseDto "Facility category retrieved successfully"
// @Failure 400 {object} map[string]interface{} "Bad Request: Invalid ID"
// @Failure 404 {object} map[string]interface{} "Not Found: Facility category not found"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /facilities/categories/{id} [get]
func (h *CategoriesHandler) GetById(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := validators.ParseUUID(idStr)

	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	name, err := h.Service.GetById(r.Context(), id)
	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	category := entity.Category{
		ID:   id,
		Name: *name,
	}

	response := dto.NewFacilityCategoryResponse(category)

	responseHandlers.RespondWithSuccess(w, response, http.StatusOK)
}

// List retrieves all facility categories.
// @Summary Get all facility categories
// @Description Retrieves a list of all facility categories.
// @Tags facility-categories
// @Accept json
// @Produce json
// @Success 200 {array} dto.CategoryResponseDto "GetMemberships of facility categories retrieved successfully"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /facilities/categories [get]
func (h *CategoriesHandler) List(w http.ResponseWriter, r *http.Request) {
	facilityCategories, err := h.Service.List(r.Context())
	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	result := make([]dto.CategoryResponseDto, len(facilityCategories))
	for i, facilityCategory := range facilityCategories {
		result[i] = dto.NewFacilityCategoryResponse(facilityCategory)
	}

	responseHandlers.RespondWithSuccess(w, result, http.StatusOK)
}

// Update updates an existing facility category by ID.
// @Summary Update a facility category
// @Description Updates the details of an existing facility category by ID.
// @Tags facility-categories
// @Accept json
// @Produce json
// @Param id path string true "Facility Category ID"
// @Param facility_category body dto.CategoryRequestDto true "Updated facility category details"
// @Success 204 {object} map[string]interface{} "No Content: Facility category updated successfully"
// @Failure 400 {object} map[string]interface{} "Bad Request: Invalid input"
// @Failure 404 {object} map[string]interface{} "Not Found: Facility category not found"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /facilities/categories/{id} [put]
func (h *CategoriesHandler) Update(w http.ResponseWriter, r *http.Request) {

	idStr := chi.URLParam(r, "id")

	var requestDto dto.CategoryRequestDto

	if err := validators.ParseJSON(r.Body, &requestDto); err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	facilityTypeVo, err := requestDto.ToUpdateFacilityCategoryValueObject(idStr)

	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	if err := h.Service.Repo.UpdateFacilityType(r.Context(), facilityTypeVo); err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	responseHandlers.RespondWithSuccess(w, nil, http.StatusNoContent)
}

// Delete deletes a facility category by ID.
// @Summary Delete a facility category
// @Description Deletes a facility category by its ID.
// @Tags facility-categories
// @Accept json
// @Produce json
// @Param id path string true "Facility Category ID"
// @Success 204 {object} map[string]interface{} "No Content: Facility category deleted successfully"
// @Failure 400 {object} map[string]interface{} "Bad Request: Invalid ID"
// @Failure 404 {object} map[string]interface{} "Not Found: Facility category not found"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /facilities/categories/{id} [delete]
func (h *CategoriesHandler) Delete(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := validators.ParseUUID(idStr)

	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	if err = h.Service.Delete(r.Context(), id); err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	responseHandlers.RespondWithSuccess(w, nil, http.StatusNoContent)
}
