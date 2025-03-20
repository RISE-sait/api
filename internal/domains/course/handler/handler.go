package course

import (
	dto "api/internal/domains/course/dto"
	persistence "api/internal/domains/course/persistence/repository"
	responseHandlers "api/internal/libs/responses"
	"api/internal/libs/validators"
	"net/http"

	"github.com/go-chi/chi"
)

type Handler struct {
	Repo *persistence.Repository
}

func NewHandler(repo *persistence.Repository) *Handler {
	return &Handler{Repo: repo}
}

// CreateCourse creates a new course.
// @Summary Create a new course
// @Description Create a new course
// @Tags courses
// @Accept json
// @Produce json
// @Param course body dto.RequestDto true "Course details"
// @Security Bearer
// @Success 201 {object} course.ResponseDto "Course created successfully"
// @Failure 400 {object} map[string]interface{} "Bad Request: Invalid input"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /courses [post]
func (h *Handler) CreateCourse(w http.ResponseWriter, r *http.Request) {
	var requestDto dto.RequestDto

	if err := validators.ParseJSON(r.Body, &requestDto); err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	details, err := requestDto.ToCreateCourseDetails()

	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	createdCourse, err := h.Repo.CreateCourse(r.Context(), details)

	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	responseBody := dto.NewCourseResponse(createdCourse)

	responseHandlers.RespondWithSuccess(w, responseBody, http.StatusCreated)
}

// GetCourseById retrieves a course by ID.
// @Summary Get a course by ID
// @Description Get a course by ID
// @Tags courses
// @Accept json
// @Produce json
// @Param id path string true "Course ID"
// @Success 200 {object} course.ResponseDto "Course retrieved successfully"
// @Failure 400 {object} map[string]interface{} "Bad Request: Invalid ID"
// @Failure 404 {object} map[string]interface{} "Not Found: Course not found"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /courses/{id} [get]
func (h *Handler) GetCourseById(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := validators.ParseUUID(idStr)

	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	course, err := h.Repo.GetCourseById(r.Context(), id)

	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	response := dto.NewCourseResponse(course)

	responseHandlers.RespondWithSuccess(w, response, http.StatusOK)
}

// GetCourses retrieves a list of courses.
// @Summary Get a list of courses
// @Description Get a list of courses
// @Tags courses
// @Accept json
// @Produce json
// @Success 200 {array} course.ResponseDto "GetMemberships of courses retrieved successfully"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /courses [get]
func (h *Handler) GetCourses(w http.ResponseWriter, r *http.Request) {

	courses, err := h.Repo.GetCourses(r.Context())
	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	result := make([]dto.ResponseDto, len(courses))

	for i, course := range courses {
		result[i] = dto.NewCourseResponse(course)
	}

	responseHandlers.RespondWithSuccess(w, result, http.StatusOK)
}

// UpdateCourse updates an existing course.
// @Summary Update a course
// @Description Update a course
// @Tags courses
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path string true "Course ID"
// @Param course body dto.RequestDto true "Course details"
// @Success 204 "No Content: Course updated successfully"
// @Failure 400 {object} map[string]interface{} "Bad Request: Invalid input"
// @Failure 404 {object} map[string]interface{} "Not Found: Course not found"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /courses/{id} [put]
func (h *Handler) UpdateCourse(w http.ResponseWriter, r *http.Request) {

	idStr := chi.URLParam(r, "id")

	var requestDto dto.RequestDto
	if err := validators.ParseJSON(r.Body, &requestDto); err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	courseUpdate, err := requestDto.ToUpdateCourseDetails(idStr)

	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	if err = h.Repo.UpdateCourse(r.Context(), courseUpdate); err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	responseHandlers.RespondWithSuccess(w, nil, http.StatusNoContent)
}

// DeleteCourse deletes a course by Id.
// @Summary Delete a course
// @Description Delete a course by Id
// @Tags courses
// @Accept json
// @Produce json
// @Param id path string true "Course Id"
// @Security Bearer
// @Success 204
// @Failure 400 {object} map[string]interface{} "Bad Request: Invalid Id"
// @Failure 404 {object} map[string]interface{} "Not Found: Course not found"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /courses/{id} [delete]
func (h *Handler) DeleteCourse(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := validators.ParseUUID(idStr)

	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	if err = h.Repo.DeleteCourse(r.Context(), id); err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	responseHandlers.RespondWithSuccess(w, nil, http.StatusNoContent)
}
