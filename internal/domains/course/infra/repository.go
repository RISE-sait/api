package course

import (
	"api/internal/domains/course/dto"
	db "api/internal/domains/course/infra/sqlc"
	errLib "api/internal/libs/errors"
	"context"
	"database/sql"
	"errors"
	"github.com/google/uuid"
	"net/http"
)

type Repository struct {
	Queries *db.Queries
}

func (r *Repository) GetCourseById(c context.Context, id uuid.UUID) (*db.Course, *errLib.CommonError) {
	course, err := r.Queries.GetCourseById(c, id)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errLib.New("Course not found", http.StatusNotFound)
		}
		return nil, errLib.New(err.Error(), http.StatusInternalServerError)
	}
	return &course, nil
}

func (r *Repository) UpdateCourse(c context.Context, course *dto.UpdateCourseRequest) *errLib.CommonError {

	dbCourseParams := course.ToDBParams()

	row, err := r.Queries.UpdateCourse(c, *dbCourseParams)

	if err != nil {
		return errLib.New(err.Error(), http.StatusInternalServerError)
	}

	if row == 0 {
		return errLib.New("Course not found", http.StatusNotFound)
	}

	return nil
}

func (r *Repository) GetAllCourses(c context.Context, after string) ([]db.Course, *errLib.CommonError) {
	courses, err := r.Queries.GetAllCourses(c)

	if err != nil {
		return nil, errLib.New(err.Error(), http.StatusInternalServerError)
	}
	return courses, nil
}

func (r *Repository) DeleteCourse(c context.Context, id uuid.UUID) *errLib.CommonError {
	row, err := r.Queries.DeleteCourse(c, id)

	if err != nil {
		return errLib.New(err.Error(), http.StatusInternalServerError)
	}

	if row == 0 {
		return errLib.New("Course not found", http.StatusNotFound)
	}

	return nil
}

func (r *Repository) CreateCourse(c context.Context, course *dto.CreateCourseRequestBody) *errLib.CommonError {

	dbCourseParams := course.ToDBParams()

	row, err := r.Queries.CreateCourse(c, *dbCourseParams)

	if err != nil {
		return errLib.New(err.Error(), http.StatusInternalServerError)
	}

	if row == 0 {
		return errLib.New("Course not created", http.StatusInternalServerError)
	}

	return nil
}
