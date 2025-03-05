package repository

import (
	databaseErrors "api/internal/constants"
	"api/internal/domains/practice/entity"
	db "api/internal/domains/practice/persistence/sqlc/generated"
	"api/internal/domains/practice/values"
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

var _ IPracticeRepository = (*PracticeRepository)(nil)

type PracticeRepository struct {
	Queries *db.Queries
}

func NewPracticeRepository(dbQueries *db.Queries) *PracticeRepository {
	return &PracticeRepository{
		Queries: dbQueries,
	}
}

func (r *PracticeRepository) GetPracticeByName(c context.Context, name string) (*entity.Practice, *errLib.CommonError) {
	practice, err := r.Queries.GetPracticeByName(c, name)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errLib.New("Practice not found", http.StatusNotFound)
		}
		return nil, errLib.New("Internal server error", http.StatusInternalServerError)
	}

	return &entity.Practice{
		ID:          practice.ID,
		Name:        practice.Name,
		Description: practice.Description.String,
	}, nil
}

func (r *PracticeRepository) Update(ctx context.Context, practice entity.Practice) *errLib.CommonError {

	dbCourseParams := db.UpdatePracticeParams{
		ID:   practice.ID,
		Name: practice.Name,
		Description: sql.NullString{
			String: practice.Description,
			Valid:  practice.Description != "",
		},
	}

	row, err := r.Queries.UpdatePractice(ctx, dbCourseParams)

	if err != nil {
		// Check if the error is a unique violation (duplicate name)
		var pqErr *pq.Error
		if errors.As(err, &pqErr) {
			if pqErr.Code == databaseErrors.UniqueViolation {
				return errLib.New("Practice name already exists", http.StatusConflict)
			}
			log.Println(fmt.Sprintf("Database error %v", err.Error()))
			return errLib.New("Database error", http.StatusInternalServerError)
		}
		return errLib.New("Internal server error", http.StatusInternalServerError)
	}

	if row == 0 {
		return errLib.New("Practice not found", http.StatusNotFound)
	}

	return nil
}

func (r *PracticeRepository) List(ctx context.Context) ([]entity.Practice, *errLib.CommonError) {

	dbPractices, err := r.Queries.GetPractices(ctx)

	if err != nil {

		log.Println("Error getting practices: ", err)
		dbErr := errLib.New("Internal server error", http.StatusInternalServerError)

		return []entity.Practice{}, dbErr
	}

	courses := make([]entity.Practice, len(dbPractices))

	for i, dbCourse := range dbPractices {
		courses[i] = entity.Practice{
			ID:          dbCourse.ID,
			Name:        dbCourse.Name,
			Description: dbCourse.Description.String,
		}
	}

	return courses, nil
}

func (r *PracticeRepository) Delete(c context.Context, id uuid.UUID) *errLib.CommonError {
	row, err := r.Queries.DeletePractice(c, id)

	if err != nil {
		return errLib.New("Internal server error", http.StatusInternalServerError)
	}

	if row == 0 {
		return errLib.New("Practice not found", http.StatusNotFound)
	}

	return nil
}

func (r *PracticeRepository) Create(c context.Context, courseDetails *values.PracticeDetails) (*entity.Practice, *errLib.CommonError) {

	dbPracticeParams := db.CreatePracticeParams{
		Name: courseDetails.Name, Description: sql.NullString{
			String: courseDetails.Description,
			Valid:  courseDetails.Description != "",
		},
	}

	course, err := r.Queries.CreatePractice(c, dbPracticeParams)

	if err != nil {
		// Check if the error is a unique violation (error code 23505)
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			// Return a custom error for unique violation
			return nil, errLib.New("Practice name already exists", http.StatusConflict)
		}

		// Return a generic internal server error for other cases
		return nil, errLib.New("Internal server error", http.StatusInternalServerError)
	}

	return &entity.Practice{
		ID:          course.ID,
		Name:        course.Name,
		Description: course.Description.String,
	}, nil
}
