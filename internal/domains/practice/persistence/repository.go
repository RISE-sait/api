package practice

import (
	databaseErrors "api/internal/constants"
	db "api/internal/domains/practice/persistence/sqlc/generated"
	"api/internal/domains/practice/values"
	errLib "api/internal/libs/errors"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"log"
	"net/http"
)

type Repository struct {
	Queries *db.Queries
}

func NewPracticeRepository(dbQueries *db.Queries) *Repository {
	return &Repository{
		Queries: dbQueries,
	}
}

func (r *Repository) GetPracticeByName(c context.Context, name string) (values.GetPracticeValues, *errLib.CommonError) {

	var response values.GetPracticeValues

	practice, err := r.Queries.GetPracticeByName(c, name)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return response, errLib.New("Practice not found", http.StatusNotFound)
		}
		return response, errLib.New("Internal server error", http.StatusInternalServerError)
	}

	return values.GetPracticeValues{
		PracticeDetails: values.PracticeDetails{
			Name:        practice.Name,
			Description: practice.Description,
		},
		ID:        practice.ID,
		CreatedAt: practice.CreatedAt,
		UpdatedAt: practice.UpdatedAt,
	}, nil
}

func (r *Repository) GetPracticeLevels() []string {
	dbLevels := db.AllPracticeLevelValues()

	var levels []string

	for _, dbLevel := range dbLevels {
		levels = append(levels, string(dbLevel))
	}

	return levels
}

func (r *Repository) Update(ctx context.Context, practice values.UpdatePracticeValues) *errLib.CommonError {

	dbCourseParams := db.UpdatePracticeParams{
		ID:          practice.ID,
		Name:        practice.PracticeDetails.Name,
		Description: practice.PracticeDetails.Description,
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

func (r *Repository) List(ctx context.Context) ([]values.GetPracticeValues, *errLib.CommonError) {

	dbPractices, err := r.Queries.GetPractices(ctx)

	if err != nil {

		log.Println("Error getting practices: ", err)
		dbErr := errLib.New("Internal server error", http.StatusInternalServerError)

		return nil, dbErr
	}

	practices := make([]values.GetPracticeValues, len(dbPractices))

	for i, dbPractice := range dbPractices {
		practices[i] = values.GetPracticeValues{
			ID:        dbPractice.ID,
			CreatedAt: dbPractice.CreatedAt,
			UpdatedAt: dbPractice.UpdatedAt,
			PracticeDetails: values.PracticeDetails{
				Name:        dbPractice.Name,
				Description: dbPractice.Description,
				Level:       string(dbPractice.Level),
			},
		}
	}

	return practices, nil
}

func (r *Repository) Delete(c context.Context, id uuid.UUID) *errLib.CommonError {
	row, err := r.Queries.DeletePractice(c, id)

	if err != nil {
		return errLib.New("Internal server error", http.StatusInternalServerError)
	}

	if row == 0 {
		return errLib.New("Practice not found", http.StatusNotFound)
	}

	return nil
}

func (r *Repository) Create(c context.Context, courseDetails values.CreatePracticeValues) (values.GetPracticeValues, *errLib.CommonError) {

	var response values.GetPracticeValues

	dbPracticeParams := db.CreatePracticeParams{
		Name: courseDetails.Name, Description: courseDetails.Description,
	}

	course, err := r.Queries.CreatePractice(c, dbPracticeParams)

	if err != nil {
		// Check if the error is a unique violation (error code 23505)
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == databaseErrors.UniqueViolation {
			// Return a custom error for unique violation
			return response, errLib.New("Practice name already exists", http.StatusConflict)
		}

		// Return a generic internal server error for other cases
		return response, errLib.New("Internal server error", http.StatusInternalServerError)
	}

	return values.GetPracticeValues{
		ID:        course.ID,
		CreatedAt: course.CreatedAt,
		UpdatedAt: course.UpdatedAt,
		PracticeDetails: values.PracticeDetails{
			Name:        courseDetails.Name,
			Description: courseDetails.Description,
		},
	}, nil
}
