package course

import (
	"api/internal/di"
	entity "api/internal/domains/course/entities"
	db "api/internal/domains/course/persistence/sqlc/generated"
	"api/internal/domains/course/values"
	errLib "api/internal/libs/errors"
	"context"
	"database/sql"
	"errors"
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

type CourseRepositoryInterface interface {
	CreateCourse(ctx context.Context, input *values.CourseDetails) (*entity.Course, *errLib.CommonError)
	GetCourseById(ctx context.Context, id uuid.UUID) (*entity.Course, *errLib.CommonError)
	GetCourses(ctx context.Context, name, description *string) ([]entity.Course, *errLib.CommonError)
	UpdateCourse(ctx context.Context, input *entity.Course) *errLib.CommonError
	DeleteCourse(ctx context.Context, id uuid.UUID) *errLib.CommonError
}

var _ CourseRepositoryInterface = (*CourseRepository)(nil)

type CourseRepository struct {
	Queries *db.Queries
}

func NewCourseRepository(container *di.Container) *CourseRepository {
	return &CourseRepository{
		Queries: container.Queries.CoursesDb,
	}
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
	}, nil
}

func (r *CourseRepository) UpdateCourse(c context.Context, course *entity.Course) *errLib.CommonError {

	dbCourseParams := db.UpdateCourseParams{
		ID:   course.ID,
		Name: course.Name,
		Description: sql.NullString{
			String: course.Description,
			Valid:  course.Description != "",
		},
	}

	row, err := r.Queries.UpdateCourse(c, dbCourseParams)

	if err != nil {
		// Check if the error is a unique violation (duplicate name)
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			return errLib.New("Course name already exists", http.StatusConflict)
		}
		return errLib.New("Internal server error", http.StatusInternalServerError)
	}

	if row == 0 {
		return errLib.New("Course not found", http.StatusNotFound)
	}

	return nil
}

func (r *CourseRepository) GetCourses(c context.Context, name, description *string) ([]entity.Course, *errLib.CommonError) {

	dbParams := db.GetCoursesParams{}

	if name != nil {
		dbParams.Name = sql.NullString{
			String: *name,
			Valid:  *name != "",
		}
	}

	if description != nil {
		dbParams.Description = sql.NullString{
			String: *description,
			Valid:  *description != "",
		}
	}

	dbCourses, err := r.Queries.GetCourses(c, dbParams)

	if err != nil {

		log.Println("Error getting courses: ", err)
		dbErr := errLib.New("Internal server error", http.StatusInternalServerError)

		return []entity.Course{}, dbErr
	}

	courses := make([]entity.Course, len(dbCourses))

	for i, dbCourse := range dbCourses {
		courses[i] = entity.Course{
			ID:          dbCourse.ID,
			Name:        dbCourse.Name,
			Description: dbCourse.Description.String,
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

func (r *CourseRepository) CreateCourse(c context.Context, courseDetails *values.CourseDetails) (*entity.Course, *errLib.CommonError) {

	dbCourseParams := db.CreateCourseParams{
		Name: courseDetails.Name, Description: sql.NullString{
			String: courseDetails.Description,
			Valid:  courseDetails.Description != "",
		},
	}

	course, err := r.Queries.CreateCourse(c, dbCourseParams)

	if err != nil {
		// Check if the error is a unique violation (error code 23505)
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			// Return a custom error for unique violation
			return nil, errLib.New("Course name already exists", http.StatusConflict)
		}

		// Return a generic internal server error for other cases
		return nil, errLib.New("Internal server error", http.StatusInternalServerError)
	}

	return &entity.Course{
		ID:          course.ID,
		Name:        course.Name,
		Description: course.Description.String,
	}, nil
}
