package course

import (
	databaseErrors "api/internal/constants"
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

type Repository struct {
	Queries *db.Queries
}

func NewCourseRepository(dbQueries *db.Queries) *Repository {
	return &Repository{
		Queries: dbQueries,
	}
}

func (r *Repository) GetCourseById(c context.Context, id uuid.UUID) (values.ReadDetails, *errLib.CommonError) {

	var course values.ReadDetails

	dbCourse, err := r.Queries.GetCourseById(c, id)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return course, errLib.New("Course not found", http.StatusNotFound)
		}
		return course, errLib.New("Internal server error", http.StatusInternalServerError)
	}

	return values.ReadDetails{
		ID:        dbCourse.ID,
		CreatedAt: dbCourse.CreatedAt,
		UpdatedAt: dbCourse.UpdatedAt,
		Details: values.Details{
			Name:        dbCourse.Name,
			Description: dbCourse.Description.String,
		},
	}, nil
}

func (r *Repository) UpdateCourse(c context.Context, course values.UpdateCourseDetails) *errLib.CommonError {

	dbCourseParams := db.UpdateCourseParams{
		ID:   course.ID,
		Name: course.Name,
		Description: sql.NullString{
			String: course.Description,
			Valid:  course.Description != "",
		},
	}

	impactedRows, err := r.Queries.UpdateCourse(c, dbCourseParams)

	if err != nil {
		// Check if the error is a unique violation (duplicate name)
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == databaseErrors.UniqueViolation {
			return errLib.New("Course name already exists", http.StatusConflict)
		}
		return errLib.New("Internal server error", http.StatusInternalServerError)
	}

	if impactedRows == 0 {
		return errLib.New("Course with associated ID not found", http.StatusNotFound)
	}

	return nil
}

func (r *Repository) GetCourses(ctx context.Context) ([]values.ReadDetails, *errLib.CommonError) {

	dbCourses, err := r.Queries.GetCourses(ctx)

	if err != nil {

		log.Println("Error getting courses: ", err)
		dbErr := errLib.New("Internal server error", http.StatusInternalServerError)

		return nil, dbErr
	}

	courses := make([]values.ReadDetails, len(dbCourses))

	for i, dbCourse := range dbCourses {
		courses[i] = values.ReadDetails{
			ID:        dbCourse.ID,
			CreatedAt: dbCourse.CreatedAt,
			UpdatedAt: dbCourse.UpdatedAt,
			Details: values.Details{
				Name:        dbCourse.Name,
				Description: dbCourse.Description.String,
			},
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

func (r *Repository) CreateCourse(c context.Context, courseDetails values.CreateCourseDetails) (values.ReadDetails, *errLib.CommonError) {

	var createdCourse values.ReadDetails

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
		if errors.As(err, &pqErr) && pqErr.Code == databaseErrors.UniqueViolation {
			// Return a custom error for unique violation
			return createdCourse, errLib.New("Course name already exists", http.StatusConflict)
		}

		// Return a generic internal server error for other cases
		log.Println("error creating course: ", err)
		return createdCourse, errLib.New("Internal server error", http.StatusInternalServerError)
	}

	return values.ReadDetails{
		ID:        course.ID,
		Details:   values.Details{Name: courseDetails.Name, Description: courseDetails.Description},
		CreatedAt: course.CreatedAt,
		UpdatedAt: course.UpdatedAt,
	}, nil
}
