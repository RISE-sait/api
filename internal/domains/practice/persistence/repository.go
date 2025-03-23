package practice

import (
	databaseErrors "api/internal/constants"
	"api/internal/custom_types"
	db "api/internal/domains/practice/persistence/sqlc/generated"
	"api/internal/domains/practice/values"
	errLib "api/internal/libs/errors"
	"context"
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

func (r *Repository) GetPracticeLevels() []string {
	dbLevels := db.AllPracticeLevelValues()

	var levels []string

	for _, dbLevel := range dbLevels {
		levels = append(levels, string(dbLevel))
	}

	return levels
}

func (r *Repository) Update(ctx context.Context, practice values.UpdatePracticeValues) *errLib.CommonError {

	params := db.UpdatePracticeParams{
		ID:          practice.ID,
		Name:        practice.Name,
		Description: practice.Description,
		Level:       db.PracticeLevel(practice.Level),
	}

	if practice.PayGPrice != nil {
		params.PaygPrice = custom_types.NullDecimal{
			Decimal: *practice.PayGPrice,
			Valid:   true,
		}
	}

	if err := r.Queries.UpdatePractice(ctx, params); err != nil {
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
		practice := values.GetPracticeValues{
			ID:        dbPractice.ID,
			CreatedAt: dbPractice.CreatedAt,
			UpdatedAt: dbPractice.UpdatedAt,
			PracticeDetails: values.PracticeDetails{
				Name:        dbPractice.Name,
				Description: dbPractice.Description,
				Level:       string(dbPractice.Level),
			},
		}

		if dbPractice.PaygPrice.Valid {
			practice.PracticeDetails.PayGPrice = &dbPractice.PaygPrice.Decimal
		}
		practices[i] = practice
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

func (r *Repository) Create(c context.Context, practiceDetails values.CreatePracticeValues) *errLib.CommonError {

	dbPracticeParams := db.CreatePracticeParams{
		Name:        practiceDetails.Name,
		Description: practiceDetails.Description,
		Level:       db.PracticeLevel(practiceDetails.Level),
	}

	if practiceDetails.PayGPrice != nil {
		dbPracticeParams.PaygPrice = custom_types.NullDecimal{
			Decimal: dbPracticeParams.PaygPrice.Decimal,
			Valid:   true,
		}
	}

	if err := r.Queries.CreatePractice(c, dbPracticeParams); err != nil {
		// Check if the error is a unique violation (error code 23505)
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == databaseErrors.UniqueViolation {
			// Return a custom error for unique violation
			return errLib.New("Practice name already exists", http.StatusConflict)
		}

		// Return a generic internal server error for other cases
		return errLib.New("Internal server error", http.StatusInternalServerError)
	}

	return nil
}
