package course

import (
	entity "api/internal/domains/course/entity"
	db "api/internal/domains/course/persistence/sqlc/generated"
	values "api/internal/domains/course/values"
	errLib "api/internal/libs/errors"
	"context"
	"database/sql"
	"errors"
	"github.com/lib/pq"
	"log"
	"net/http"

	"github.com/google/uuid"
)

var _ RepositoryInterface = (*Repository)(nil)

type Repository struct {
	Queries *db.Queries
}

func NewCourseRepository(dbQueries *db.Queries) *Repository {
	return &Repository{
		Queries: dbQueries,
	}
}

func (r *Repository) GetCourseById(c context.Context, id uuid.UUID) (*entity.Course, *errLib.CommonError) {
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

func (r *Repository) UpdateCourse(c context.Context, course *entity.Course) (*entity.Course, *errLib.CommonError) {

	dbCourseParams := db.UpdateCourseParams{
		ID:   course.ID,
		Name: course.Name,
		Description: sql.NullString{
			String: course.Description,
			Valid:  course.Description != "",
		},
	}

	dbCourse, err := r.Queries.UpdateCourse(c, dbCourseParams)

	if err != nil {
		// Check if the error is a unique violation (duplicate name)
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" {
			return nil, errLib.New("Course name already exists", http.StatusConflict)
		}
		return nil, errLib.New("Internal server error", http.StatusInternalServerError)
	}

	return &entity.Course{
		ID:          dbCourse.ID,
		Name:        dbCourse.Name,
		Description: dbCourse.Description.String,
	}, nil
}

func (r *Repository) GetCourses(c context.Context, name, description *string) ([]entity.Course, *errLib.CommonError) {

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

func (r *Repository) DeleteCourse(c context.Context, id uuid.UUID) *errLib.CommonError {
	row, err := r.Queries.DeleteCourse(c, id)

	if err != nil {
		return errLib.New("Internal server error", http.StatusInternalServerError)
	}

	if row == 0 {
		return errLib.New("Course not found", http.StatusNotFound)
	}

	return nil
}

func (r *Repository) CreateCourse(c context.Context, courseDetails *values.Details) (*entity.Course, *errLib.CommonError) {

	dbCourseParams := db.CreateCourseParams{
		Name: courseDetails.Name, Description: sql.NullString{
			String: courseDetails.Description,
			Valid:  courseDetails.Description != "",
		},
	}

	course, err := r.Queries.CreateCourse(c, dbCourseParams)

	if err != nil {
		// Check if the error is a unique violation (error code 23505)
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" {
			// Return a custom error for unique violation
			return nil, errLib.New("Course name already exists", http.StatusConflict)
		}

		// Return a generic internal server error for other cases
		log.Println("error creating course: ", err)
		return nil, errLib.New("Internal server error", http.StatusInternalServerError)
	}

	return &entity.Course{
		ID:          course.ID,
		Name:        course.Name,
		Description: course.Description.String,
	}, nil
}
