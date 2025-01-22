package course

import (
	"api/internal/domains/course/dto"
	"api/internal/libs/responses"
	"api/internal/libs/validators"
	"github.com/go-chi/chi"
	"net/http"
)

type Handler struct {
	CourseService *Service
}

func NewHandler(courseService *Service) *Handler {
	return &Handler{CourseService: courseService}
}

func (h *Handler) CreateCourse(w http.ResponseWriter, r *http.Request) {
	var targetBody dto.CreateCourseRequestBody

	if err := validators.ParseRequestBodyToJSON(r.Body, &targetBody); err != nil {
		response_handlers.RespondWithError(w, err)
		return
	}

	if err := h.CourseService.CreateCourse(r.Context(), targetBody); err != nil {
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

	response_handlers.RespondWithSuccess(w, *course, http.StatusOK)
}

func (h *Handler) GetAllCourses(w http.ResponseWriter, r *http.Request) {
	courses, err := h.CourseService.GetAllCourses(r.Context())
	if err != nil {
		response_handlers.RespondWithError(w, err)
		return
	}

	response_handlers.RespondWithSuccess(w, *courses, http.StatusOK)
}

func (h *Handler) UpdateCourse(w http.ResponseWriter, r *http.Request) {
	var targetBody dto.UpdateCourseRequest

	if err := validators.ParseRequestBodyToJSON(r.Body, &targetBody); err != nil {
		response_handlers.RespondWithError(w, err)
		return
	}

	if err := h.CourseService.UpdateCourse(r.Context(), targetBody); err != nil {
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
