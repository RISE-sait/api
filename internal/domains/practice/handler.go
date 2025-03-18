package practice

import (
	dto "api/internal/domains/practice/dto"
	repository "api/internal/domains/practice/persistence"
	"api/internal/domains/practice/values"
	responseHandlers "api/internal/libs/responses"
	"api/internal/libs/validators"
	"net/http"

	"github.com/go-chi/chi"
)

type Handler struct {
	Repo *repository.Repository
}

func NewHandler(repo *repository.Repository) *Handler {
	return &Handler{Repo: repo}
}

// CreatePractice creates a new practice.
// @Summary Create a new practice
// @Description Create a new practice
// @Tags practices
// @Accept json
// @Produce json
// @Param practice body dto.RequestDto true "Practice details"
// @Security Bearer
// @Success 201 {object} dto.Response "Practice created successfully"
// @Failure 400 {object} map[string]interface{} "Bad Request: Invalid input"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /practices [post]
func (h *Handler) CreatePractice(w http.ResponseWriter, r *http.Request) {
	var requestDto dto.RequestDto

	if err := validators.ParseJSON(r.Body, &requestDto); err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	courseCreate, err := requestDto.ToCreateValueObjects()

	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	course, err := h.Repo.Create(r.Context(), courseCreate)

	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	responseBody := mapReadValuesToResponse(course)

	responseHandlers.RespondWithSuccess(w, responseBody, http.StatusCreated)
}

// GetPractices retrieves a list of practices.
// @Summary Get a list of practices
// @Description Get a list of practices
// @Tags practices
// @Accept json
// @Produce json
// @Success 200 {array} dto.Response "GetMemberships of practices retrieved successfully"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /practices [get]
func (h *Handler) GetPractices(w http.ResponseWriter, r *http.Request) {

	practices, err := h.Repo.List(r.Context())
	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	result := make([]dto.Response, len(practices))

	for i, course := range practices {
		result[i] = mapReadValuesToResponse(course)
	}

	responseHandlers.RespondWithSuccess(w, result, http.StatusOK)
}

// GetPracticeLevels retrieves available practice levels.
// @Summary Get practice levels
// @Description Retrieves a list of available practice levels.
// @Tags practices
// @Accept json
// @Produce json
// @Success 200 {array} dto.LevelsResponse "Get practice levels retrieved successfully"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /practices/levels [get]
func (h *Handler) GetPracticeLevels(w http.ResponseWriter, _ *http.Request) {
	levels := h.Repo.GetPracticeLevels()

	response := dto.LevelsResponse{Name: levels}

	responseHandlers.RespondWithSuccess(w, response, http.StatusOK)
}

// UpdatePractice updates an existing practice.
// @Summary Update a practice
// @Description Update a practice
// @Tags practices
// @Accept json
// @Produce json
// @Param id path string true "Practice ID"
// @Param practice body dto.RequestDto true "Practice details"
// @Success 204 "No Content: Practice updated successfully"
// @Failure 400 {object} map[string]interface{} "Bad Request: Invalid input"
// @Failure 404 {object} map[string]interface{} "Not Found: Practice not found"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /practices/{id} [put]
func (h *Handler) UpdatePractice(w http.ResponseWriter, r *http.Request) {

	idStr := chi.URLParam(r, "id")

	var requestDto dto.RequestDto

	if err := validators.ParseJSON(r.Body, &requestDto); err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	practiceUpdate, err := requestDto.ToUpdateValueObjects(idStr)

	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	if err = h.Repo.Update(r.Context(), practiceUpdate); err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	responseHandlers.RespondWithSuccess(w, nil, http.StatusNoContent)
}

// DeletePractice deletes a practice by ID.
// @Summary Delete a practice
// @Description Delete a practice by ID
// @Tags practices
// @Accept json
// @Produce json
// @Param id path string true "Practice ID"
// @Security Bearer
// @Success 204 "No Content: Practice deleted successfully"
// @Failure 400 {object} map[string]interface{} "Bad Request: Invalid ID"
// @Failure 404 {object} map[string]interface{} "Not Found: Practice not found"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /practices/{id} [delete]
func (h *Handler) DeletePractice(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := validators.ParseUUID(idStr)

	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	if err = h.Repo.Delete(r.Context(), id); err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	responseHandlers.RespondWithSuccess(w, nil, http.StatusNoContent)
}

func mapReadValuesToResponse(practice values.GetPracticeValues) dto.Response {
	return dto.Response{
		ID:          practice.ID,
		Name:        practice.PracticeDetails.Name,
		Description: practice.PracticeDetails.Description,
		Level:       practice.PracticeDetails.Level,
		Capacity:    practice.PracticeDetails.Capacity,
		CreatedAt:   practice.CreatedAt,
		UpdatedAt:   practice.UpdatedAt,
	}
}
