package course

import (
	"api/internal/domains/course/dto"
	"api/internal/domains/course/values"
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

func mapEntityToResponse(course *values.CourseAllFields) dto.CourseResponse {
	return dto.CourseResponse{
		ID:          course.ID,
		Name:        course.Name,
		StartDate:   course.StartDate,
		EndDate:     course.EndDate,
		Description: course.Description,
		Capacity:    course.Capacity,
	}
}
