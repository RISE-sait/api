package courses

import (
	"api/internal/types"
	"api/internal/utils"
	db "api/sqlc"
	"context"
	"database/sql"
	"net/http"

	"github.com/google/uuid"
)

type Repository struct {
	Queries *db.Queries
}

func (r *Repository) GetCourseById(c context.Context, id uuid.UUID) (*db.Course, *types.HTTPError) {
	course, err := r.Queries.GetCourseById(c, id)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, utils.CreateHTTPError("Course not found", http.StatusNotFound)
		}
		return nil, utils.CreateHTTPError(err.Error(), http.StatusInternalServerError)
	}
	return &course, nil
}

func (r *Repository) UpdateCourse(c context.Context, course *db.UpdateCourseParams) *types.HTTPError {
	row, err := r.Queries.UpdateCourse(c, *course)

	if err != nil {
		return utils.CreateHTTPError(err.Error(), http.StatusInternalServerError)
	}

	if row == 0 {
		return utils.CreateHTTPError("Course not found", http.StatusNotFound)
	}

	return nil
}

func (r *Repository) GetAllCourses(c context.Context, after string) (*[]db.Course, error) {
	courses, err := r.Queries.GetAllCourses(c)

	if err != nil {
		return nil, err
	}
	return &courses, nil
}

func (r *Repository) DeleteCourse(c context.Context, id uuid.UUID) error {
	row, err := r.Queries.DeleteCourse(c, id)

	if err != nil {
		return err
	}

	if row == 0 {
		return err
	}

	return nil
}

func (r *Repository) CreateCourse(c context.Context, course *db.CreateCourseParams) *types.HTTPError {
	row, err := r.Queries.CreateCourse(c, *course)

	if err != nil || row == 0 {
		return utils.CreateHTTPError(err.Error(), http.StatusInternalServerError)
	}
	return nil
}
