package practice

import (
	"api/internal/domains/practice/dto"
	repository "api/internal/domains/practice/persistence"
	"api/internal/domains/practice/values"
	errLib "api/internal/libs/errors"
	responseHandlers "api/internal/libs/responses"
	"api/internal/libs/validators"
	"log"
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
// @Param practice body dto.PracticeRequestDto true "Practice details"
// @Security Bearer
// @Success 201 {object} dto.PracticeResponse "Practice created successfully"
// @Failure 400 {object} map[string]interface{} "Bad Request: Invalid input"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /practices [post]
func (h *Handler) CreatePractice(w http.ResponseWriter, r *http.Request) {
	var requestDto dto.PracticeRequestDto

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

	responseBody := mapEntityToResponse(course)

	responseHandlers.RespondWithSuccess(w, responseBody, http.StatusCreated)
}

// GetPracticeByName retrieves a practice by name.
// @Summary Get a practice by name
// @Description Get a practice by name
// @Tags practices
// @Accept json
// @Produce json
// @Param name path string true "Practice Name"
// @Success 200 {object} dto.PracticeResponse "Practice retrieved successfully"
// @Failure 400 {object} map[string]interface{} "Bad Request: Invalid Name"
// @Failure 404 {object} map[string]interface{} "Not Found: Practice not found"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /practices/{name} [get]
func (h *Handler) GetPracticeByName(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")

	if name == "" {
		responseHandlers.RespondWithError(w, errLib.New("Name cannot be empty", http.StatusBadRequest))
	}

	course, err := h.Repo.GetPracticeByName(r.Context(), name)

	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	response := mapEntityToResponse(course)

	responseHandlers.RespondWithSuccess(w, response, http.StatusOK)
}

// GetPractices retrieves a list of practices.
// @Summary Get a list of practices
// @Description Get a list of practices
// @Tags practices
// @Accept json
// @Produce json
// @Param name query string false "Filter by practice name"
// @Param description query string false "Filter by practice description"
// @Success 200 {array} dto.PracticeResponse "GetMemberships of practices retrieved successfully"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /practices [get]
func (h *Handler) GetPractices(w http.ResponseWriter, r *http.Request) {

	practices, err := h.Repo.List(r.Context())
	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	result := make([]dto.PracticeResponse, len(practices))

	for i, course := range practices {
		result[i] = mapEntityToResponse(course)
	}

	responseHandlers.RespondWithSuccess(w, result, http.StatusOK)
}

// GetPracticeLevels retrieves available practice levels.
// @Summary Get practice levels
// @Description Retrieves a list of available practice levels.
// @Tags practices
// @Accept json
// @Produce json
// @Success 200 {object} map[string][]string "List of practice levels"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /practices/levels [get]
func (h *Handler) GetPracticeLevels(w http.ResponseWriter, _ *http.Request) {
	levels := h.Repo.GetPracticeLevels()

	response := map[string][]string{"practice_levels": levels}

	responseHandlers.RespondWithSuccess(w, response, http.StatusOK)
}

// UpdatePractice updates an existing practice.
// @Summary Update a practice
// @Description Update a practice
// @Tags practices
// @Accept json
// @Produce json
// @Param id path string true "Practice ID"
// @Param practice body dto.PracticeRequestDto true "Practice details"
// @Success 204 "No Content: Practice updated successfully"
// @Failure 400 {object} map[string]interface{} "Bad Request: Invalid input"
// @Failure 404 {object} map[string]interface{} "Not Found: Practice not found"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /practices/{id} [put]
func (h *Handler) UpdatePractice(w http.ResponseWriter, r *http.Request) {

	idStr := chi.URLParam(r, "id")

	var requestDto dto.PracticeRequestDto

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

	log.Println("no")

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

func mapEntityToResponse(course values.GetPracticeValues) dto.PracticeResponse {
	return dto.PracticeResponse{
		ID:          course.ID,
		Name:        course.PracticeDetails.Name,
		Description: course.PracticeDetails.Description,
		CreatedAt:   course.CreatedAt,
		UpdatedAt:   course.UpdatedAt,
	}
}
