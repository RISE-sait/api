package course

import (
	databaseErrors "api/internal/constants"
	db "api/internal/domains/course/persistence/sqlc/generated"
	values "api/internal/domains/course/values"
	errLib "api/internal/libs/errors"
	"context"
	"database/sql"
	"errors"
	"fmt"
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

// handleDBError centralizes database error handling for common cases
func handleDBError(err error, entity string) *errLib.CommonError {
	if errors.Is(err, sql.ErrNoRows) {
		return errLib.New(fmt.Sprintf("%s not found", entity), http.StatusNotFound)
	}

	var pqErr *pq.Error
	if errors.As(err, &pqErr) && string(pqErr.Code) == databaseErrors.UniqueViolation {
		return errLib.New(fmt.Sprintf("%s already exists", entity), http.StatusConflict)
	}

	log.Printf("Database error on %s: %v", entity, err)
	return errLib.New("Internal server error", http.StatusInternalServerError)
}

func (r *Repository) GetCourseById(c context.Context, id uuid.UUID) (values.ReadDetails, *errLib.CommonError) {

	dbCourse, err := r.Queries.GetCourseById(c, id)

	if err != nil {
		return values.ReadDetails{}, handleDBError(err, "Course")
	}

	return values.ReadDetails{
		ID:        dbCourse.ID,
		CreatedAt: dbCourse.CreatedAt,
		UpdatedAt: dbCourse.UpdatedAt,
		Details: values.Details{
			Name:        dbCourse.Name,
			Description: dbCourse.Description,
			Capacity:    dbCourse.Capacity,
		},
	}, nil
}

func (r *Repository) UpdateCourse(c context.Context, course values.UpdateCourseDetails) *errLib.CommonError {

	dbCourseParams := db.UpdateCourseParams{
		ID:          course.ID,
		Name:        course.Name,
		Description: course.Description,
		Capacity:    course.Capacity,
	}

	impactedRows, err := r.Queries.UpdateCourse(c, dbCourseParams)

	if err != nil {
		return handleDBError(err, "Course")
	}

	if impactedRows == 0 {
		return errLib.New("Course with associated ID not found", http.StatusNotFound)
	}

	return nil
}

func (r *Repository) GetCourses(ctx context.Context) ([]values.ReadDetails, *errLib.CommonError) {

	dbCourses, err := r.Queries.GetCourses(ctx)

	if err != nil {
		return nil, handleDBError(err, "Courses")
	}

	courses := make([]values.ReadDetails, len(dbCourses))

	for i, dbCourse := range dbCourses {
		courses[i] = values.ReadDetails{
			ID:        dbCourse.ID,
			CreatedAt: dbCourse.CreatedAt,
			UpdatedAt: dbCourse.UpdatedAt,
			Details: values.Details{
				Name:        dbCourse.Name,
				Description: dbCourse.Description,
				Capacity:    dbCourse.Capacity,
			},
		}
	}

	return courses, nil
}

func (r *Repository) DeleteCourse(c context.Context, id uuid.UUID) *errLib.CommonError {
	row, err := r.Queries.DeleteCourse(c, id)

	if err != nil {
		return handleDBError(err, "Course")
	}

	if row == 0 {
		return errLib.New("Course not found", http.StatusNotFound)
	}

	return nil
}

func (r *Repository) CreateCourse(c context.Context, courseDetails values.CreateCourseDetails) (values.ReadDetails, *errLib.CommonError) {

	dbCourseParams := db.CreateCourseParams{
		Name: courseDetails.Name, Description: courseDetails.Description,
		Capacity: courseDetails.Capacity,
	}

	course, err := r.Queries.CreateCourse(c, dbCourseParams)

	if err != nil {
		return values.ReadDetails{}, handleDBError(err, "Course")
	}

	return values.ReadDetails{
		ID: course.ID,
		Details: values.Details{Name: courseDetails.Name,
			Description: courseDetails.Description,
			Capacity:    courseDetails.Capacity,
		},
		CreatedAt: course.CreatedAt,
		UpdatedAt: course.UpdatedAt,
	}, nil
}
