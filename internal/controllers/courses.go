package controllers

import (
	dto "api/internal/dtos/course"
	"api/internal/repositories"
	"api/internal/utils"
	"api/internal/utils/validators"
	"net/http"

	"github.com/go-chi/chi"
)

type CoursesController struct {
	CourseRepository *repositories.CourseRepository
}

func NewCoursesController(courseRepository *repositories.CourseRepository) *CoursesController {
	return &CoursesController{CourseRepository: courseRepository}
}

func (c *CoursesController) CreateCourse(w http.ResponseWriter, r *http.Request) {

	var targetBody dto.CreateCourseRequestBody

	if err := validators.DecodeAndValidateRequestBody(r.Body, &targetBody); err != nil {
		utils.RespondWithError(w, err)
		return
	}

	params := targetBody.ToDBParams()

	if err := c.CourseRepository.CreateCourse(r.Context(), params); err != nil {
		utils.RespondWithError(w, err)
		return
	}

	utils.RespondWithSuccess(w, "Course created", http.StatusCreated)
}

func (c *CoursesController) GetCourseById(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")

	id, err := validators.ParseUUID(idStr)

	if err != nil {
		utils.RespondWithError(w, err)
		return
	}

	course, err := c.CourseRepository.GetCourseById(r.Context(), id)
	if err != nil {
		utils.RespondWithError(w, err)
		return
	}

	utils.RespondWithSuccess(w, course, http.StatusOK)
}

func (c *CoursesController) GetAllCourses(w http.ResponseWriter, r *http.Request) {
	courses, err := c.CourseRepository.GetAllCourses(r.Context(), "")
	if err != nil {
		utils.RespondWithError(w, err)
		return
	}

	utils.RespondWithSuccess(w, courses, http.StatusOK)

}

func (c *CoursesController) UpdateCourse(w http.ResponseWriter, r *http.Request) {

	var targetBody dto.UpdateCourseRequest

	if err := validators.DecodeAndValidateRequestBody(r.Body, &targetBody); err != nil {
		utils.RespondWithError(w, err)
		return
	}

	params := targetBody.ToDBParams()

	if err := c.CourseRepository.UpdateCourse(r.Context(), params); err != nil {
		utils.RespondWithError(w, err)
		return
	}

	utils.RespondWithSuccess(w, nil, http.StatusNoContent)
}

func (c *CoursesController) DeleteCourse(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")

	id, err := validators.ParseUUID(idStr)

	if err != nil {
		utils.RespondWithError(w, err)
		return
	}

	if err = c.CourseRepository.DeleteCourse(r.Context(), id); err != nil {
		utils.RespondWithError(w, err)
		return
	}

	utils.RespondWithSuccess(w, nil, http.StatusNoContent)
}
