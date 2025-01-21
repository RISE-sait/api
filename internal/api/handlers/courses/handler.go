package courses

import (
	"api/internal/domains/courses"
	"api/internal/utils"
	"api/internal/utils/validators"
	"net/http"

	"github.com/go-chi/chi"

	dto "api/internal/shared/dto/course"
)

type Handler struct {
	Service *courses.Service
}

func NewCourseHandler(service *courses.Service) *Handler {
	return &Handler{Service: service}
}

func (h *Handler) CreateCourse(w http.ResponseWriter, r *http.Request) {
	var targetBody dto.CreateCourseRequestBody

	if err := validators.ParseReqBodyToJSON(r.Body, &targetBody); err != nil {
		utils.RespondWithError(w, err)
		return
	}

	if err := h.Service.CreateCourse(r.Context(), targetBody); err != nil {
		utils.RespondWithError(w, err)
		return
	}

	utils.RespondWithSuccess(w, "Course created", http.StatusCreated)
}

func (h *Handler) GetCourseById(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")

	id, err := validators.ParseUUID(idStr)
	if err != nil {
		utils.RespondWithError(w, err)
		return
	}

	course, err := h.Service.GetCourseById(r.Context(), id)
	if err != nil {
		utils.RespondWithError(w, err)
		return
	}

	utils.RespondWithSuccess(w, course, http.StatusOK)
}

func (h *Handler) GetAllCourses(w http.ResponseWriter, r *http.Request) {
	courses, err := h.Service.GetAllCourses(r.Context())
	if err != nil {
		utils.RespondWithError(w, err)
		return
	}

	utils.RespondWithSuccess(w, courses, http.StatusOK)
}

func (h *Handler) UpdateCourse(w http.ResponseWriter, r *http.Request) {
	var targetBody dto.UpdateCourseRequest

	if err := validators.ParseReqBodyToJSON(r.Body, &targetBody); err != nil {
		utils.RespondWithError(w, err)
		return
	}

	if err := h.Service.UpdateCourse(r.Context(), targetBody); err != nil {
		utils.RespondWithError(w, err)
		return
	}

	utils.RespondWithSuccess(w, nil, http.StatusNoContent)
}

func (h *Handler) DeleteCourse(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")

	id, err := validators.ParseUUID(idStr)
	if err != nil {
		utils.RespondWithError(w, err)
		return
	}

	if err := h.Service.DeleteCourse(r.Context(), id); err != nil {
		utils.RespondWithError(w, err)
		return
	}

	utils.RespondWithSuccess(w, nil, http.StatusNoContent)
}
