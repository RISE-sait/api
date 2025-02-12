package course

import (
	"api/internal/domains/course/dto"
	entity "api/internal/domains/course/entities"
	response_handlers "api/internal/libs/responses"
	"api/internal/libs/validators"
	"net/http"

	"github.com/go-chi/chi"
)

type CourseController struct {
	CourseService *CourseService
}

func NewCourseController(service *CourseService) *CourseController {
	return &CourseController{CourseService: service}
}

// CreateCourse creates a new course.
// @Summary Create a new course
// @Description Create a new course
// @Tags courses
// @Accept json
// @Produce json
// @Param course body dto.CourseRequestDto true "Course details"
// @Security Bearer
// @Success 201 {object} dto.CourseResponse "Course created successfully"
// @Failure 400 {object} map[string]interface{} "Bad Request: Invalid input"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /courses [post]
func (h *CourseController) CreateCourse(w http.ResponseWriter, r *http.Request) {
	var dto dto.CourseRequestDto

	if err := validators.ParseJSON(r.Body, &dto); err != nil {
		response_handlers.RespondWithError(w, err)
		return
	}

	courseCreate, err := dto.ToCreateValueObjects()

	if err != nil {
		response_handlers.RespondWithError(w, err)
		return
	}

	course, err := h.CourseService.CreateCourse(r.Context(), courseCreate)

	if err != nil {
		response_handlers.RespondWithError(w, err)
		return
	}

	responseBody := mapEntityToResponse(course)

	response_handlers.RespondWithSuccess(w, responseBody, http.StatusCreated)
}

// GetCourseById retrieves a course by ID.
// @Summary Get a course by ID
// @Description Get a course by ID
// @Tags courses
// @Accept json
// @Produce json
// @Param id path string true "Course ID"
// @Success 200 {object} dto.CourseResponse "Course retrieved successfully"
// @Failure 400 {object} map[string]interface{} "Bad Request: Invalid ID"
// @Failure 404 {object} map[string]interface{} "Not Found: Course not found"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /courses/{id} [get]
func (h *CourseController) GetCourseById(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := validators.ParseUUID(idStr)

	if err != nil {
		response_handlers.RespondWithError(w, err)
		return
	}

	course, err := h.CourseService.GetCourseById(r.Context(), id)

	if err != nil {
		response_handlers.RespondWithError(w, err)
		return
	}

	response := mapEntityToResponse(course)

	response_handlers.RespondWithSuccess(w, response, http.StatusOK)
}

// GetCourses retrieves a list of courses.
// @Summary Get a list of courses
// @Description Get a list of courses
// @Tags courses
// @Accept json
// @Produce json
// @Param name query string false "Filter by course name"
// @Param description query string false "Filter by course description"
// @Success 200 {array} dto.CourseResponse "List of courses retrieved successfully"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /courses [get]
func (h *CourseController) GetCourses(w http.ResponseWriter, r *http.Request) {

	nameStr := r.URL.Query().Get("name")
	descriptionStr := r.URL.Query().Get("description")

	var name, description *string

	if nameStr != "" {
		name = &nameStr
	}

	if descriptionStr != "" {
		description = &descriptionStr
	}

	courses, err := h.CourseService.GetCourses(r.Context(), name, description)
	if err != nil {
		response_handlers.RespondWithError(w, err)
		return
	}

	result := make([]dto.CourseResponse, len(courses))

	for i, course := range courses {
		result[i] = mapEntityToResponse(&course)
	}

	response_handlers.RespondWithSuccess(w, result, http.StatusOK)
}

// UpdateCourse updates an existing course.
// @Summary Update a course
// @Description Update a course
// @Tags courses
// @Accept json
// @Produce json
// @Param id path string true "Course ID"
// @Param course body dto.CourseRequestDto true "Course details"
// @Security Bearer
// @Success 204 "No Content: Course updated successfully"
// @Failure 400 {object} map[string]interface{} "Bad Request: Invalid input"
// @Failure 404 {object} map[string]interface{} "Not Found: Course not found"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /courses/{id} [put]
func (h *CourseController) UpdateCourse(w http.ResponseWriter, r *http.Request) {

	idStr := chi.URLParam(r, "id")

	var dto dto.CourseRequestDto

	if err := validators.ParseJSON(r.Body, &dto); err != nil {
		response_handlers.RespondWithError(w, err)
		return
	}

	courseUpdate, err := dto.ToUpdateValueObjects(idStr)

	if err != nil {
		response_handlers.RespondWithError(w, err)
		return
	}

	if err := h.CourseService.UpdateCourse(r.Context(), courseUpdate); err != nil {
		response_handlers.RespondWithError(w, err)
		return
	}

	response_handlers.RespondWithSuccess(w, nil, http.StatusNoContent)
}

// DeleteCourse deletes a course by ID.
// @Summary Delete a course
// @Description Delete a course by ID
// @Tags courses
// @Accept json
// @Produce json
// @Param id path string true "Course ID"
// @Security Bearer
// @Success 204
// @Failure 400 {object} map[string]interface{} "Bad Request: Invalid ID"
// @Failure 404 {object} map[string]interface{} "Not Found: Course not found"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /courses/{id} [delete]
func (h *CourseController) DeleteCourse(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := validators.ParseUUID(idStr)

	if err != nil {
		response_handlers.RespondWithError(w, err)
		return
	}

	if err = h.CourseService.DeleteCourse(r.Context(), id); err != nil {
		response_handlers.RespondWithError(w, err)
		return
	}

	response_handlers.RespondWithSuccess(w, nil, http.StatusNoContent)
}

func mapEntityToResponse(course *entity.Course) dto.CourseResponse {
	return dto.CourseResponse{
		ID:          course.ID,
		Name:        course.Name,
		Description: course.Description,
	}
}
