package course

import (
	course "api/internal/domains/course/application"
	"api/internal/domains/course/infra/http/dto"
	"api/internal/domains/course/infra/http/mapper"
	"api/internal/domains/course/values"
	response_handlers "api/internal/libs/responses"
	"api/internal/libs/validators"
	"net/http"

	"github.com/go-chi/chi"
)

type Handler struct {
	CourseService *course.CourseService
}

func NewHandler(courseService *course.CourseService) *Handler {
	return &Handler{CourseService: courseService}
}

func (h *Handler) CreateCourse(w http.ResponseWriter, r *http.Request) {
	var dto dto.CreateCourseRequestBody

	if err := validators.ParseJSON(r.Body, &dto); err != nil {
		response_handlers.RespondWithError(w, err)
		return
	}

	courseCreate := values.NewCourseCreate(dto.Name, dto.Description, dto.StartDate, dto.EndDate)

	if err := h.CourseService.CreateCourse(r.Context(), courseCreate); err != nil {
		response_handlers.RespondWithError(w, err)
		return
	}

	response_handlers.RespondWithSuccess(w, nil, http.StatusCreated)
}

func (h *Handler) GetCourseById(w http.ResponseWriter, r *http.Request) {
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

	response := mapper.MapEntityToResponse(course)

	response_handlers.RespondWithSuccess(w, response, http.StatusOK)
}

func (h *Handler) GetAllCourses(w http.ResponseWriter, r *http.Request) {
	courses, err := h.CourseService.GetAllCourses(r.Context())
	if err != nil {
		response_handlers.RespondWithError(w, err)
		return
	}

	result := make([]dto.CourseResponse, len(courses))

	for i, course := range courses {
		result[i] = mapper.MapEntityToResponse(&course)
	}

	response_handlers.RespondWithSuccess(w, result, http.StatusOK)
}

func (h *Handler) UpdateCourse(w http.ResponseWriter, r *http.Request) {

	idStr := chi.URLParam(r, "id")
	id, err := validators.ParseUUID(idStr)

	if err != nil {
		response_handlers.RespondWithError(w, err)
		return
	}

	var dto dto.UpdateCourseRequest

	if err := validators.ParseJSON(r.Body, &dto); err != nil {
		response_handlers.RespondWithError(w, err)
		return
	}

	courseUpdate := values.NewCourseUpdate(id, dto.Name, dto.Description, dto.StartDate, dto.EndDate)

	if err := h.CourseService.UpdateCourse(r.Context(), courseUpdate); err != nil {
		response_handlers.RespondWithError(w, err)
		return
	}

	response_handlers.RespondWithSuccess(w, nil, http.StatusNoContent)
}

func (h *Handler) DeleteCourse(w http.ResponseWriter, r *http.Request) {
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
