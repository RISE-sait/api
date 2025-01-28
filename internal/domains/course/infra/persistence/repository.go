package persistence

import (
	entity "api/internal/domains/course/entities"
	db "api/internal/domains/course/infra/persistence/sqlc/generated"
	"api/internal/domains/course/values"
	errLib "api/internal/libs/errors"
	"context"
	"database/sql"
	"errors"
	"net/http"

	"github.com/google/uuid"
)

type CourseRepository struct {
	Queries *db.Queries
}

func (r *CourseRepository) GetCourseById(c context.Context, id uuid.UUID) (*entity.Course, *errLib.CommonError) {
	course, err := r.Queries.GetCourseById(c, id)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errLib.New("Course not found", http.StatusNotFound)
		}
		return nil, errLib.New("Internal server error", http.StatusInternalServerError)
	}

	return &entity.Course{
		ID:          course.ID,
		Name:        course.Name,
		Description: course.Description.String,
		StartDate:   course.StartDate,
		EndDate:     course.EndDate,
	}, nil
}

func (r *CourseRepository) UpdateCourse(c context.Context, course *values.CourseUpdate) *errLib.CommonError {

	dbCourseParams := db.UpdateCourseParams{
		ID:   course.ID,
		Name: course.Name,
		Description: sql.NullString{
			String: course.Description,
			Valid:  course.Description != "",
		},
		StartDate: course.StartDate,
		EndDate:   course.EndDate,
	}

	row, err := r.Queries.UpdateCourse(c, dbCourseParams)

	if err != nil {
		return errLib.New("Internal server error", http.StatusInternalServerError)
	}

	if row == 0 {
		return errLib.New("Course not found", http.StatusNotFound)
	}

	return nil
}

func (r *CourseRepository) GetAllCourses(c context.Context, after string) ([]entity.Course, *errLib.CommonError) {
	dbCourses, err := r.Queries.GetAllCourses(c)

	if err != nil {
		dbErr := errLib.New("Internal server error", http.StatusInternalServerError)

		return []entity.Course{}, dbErr
	}

	courses := make([]entity.Course, len(dbCourses))

	for i, dbCourse := range dbCourses {
		courses[i] = entity.Course{
			ID:          dbCourse.ID,
			Name:        dbCourse.Name,
			Description: dbCourse.Description.String,
			StartDate:   dbCourse.StartDate,
			EndDate:     dbCourse.EndDate,
		}
	}

	return courses, nil
}

func (r *CourseRepository) DeleteCourse(c context.Context, id uuid.UUID) *errLib.CommonError {
	row, err := r.Queries.DeleteCourse(c, id)

	if err != nil {
		return errLib.New("Internal server error", http.StatusInternalServerError)
	}

	if row == 0 {
		return errLib.New("Course not found", http.StatusNotFound)
	}

	return nil
}

func (r *CourseRepository) CreateCourse(c context.Context, course *values.CourseCreate) *errLib.CommonError {

	dbCourseParams := db.CreateCourseParams{
		Name: course.Name, Description: sql.NullString{
			String: course.Description,
			Valid:  course.Description != "",
		},
		StartDate: course.StartDate,
		EndDate:   course.EndDate,
	}

	row, err := r.Queries.CreateCourse(c, dbCourseParams)

	if err != nil {
		return errLib.New("Internal server error", http.StatusInternalServerError)
	}

	if row == 0 {
		return errLib.New("Course not created", http.StatusInternalServerError)
	}

	return nil
}
