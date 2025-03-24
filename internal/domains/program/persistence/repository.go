package program

import (
	databaseErrors "api/internal/constants"
	db "api/internal/domains/program/persistence/sqlc/generated"
	"api/internal/domains/program/values"
	errLib "api/internal/libs/errors"
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

type Repository struct {
	Queries *db.Queries
}

func NewProgramRepository(dbQueries *db.Queries) *Repository {
	return &Repository{
		Queries: dbQueries,
	}
}

func (r *Repository) GetProgramLevels() []string {
	dbLevels := db.AllProgramProgramLevelValues()

	var levels []string

	for _, dbLevel := range dbLevels {
		levels = append(levels, string(dbLevel))
	}

	return levels
}

func (r *Repository) Update(ctx context.Context, program values.UpdateProgramValues) *errLib.CommonError {

	params := db.UpdateProgramParams{
		ID:          program.ID,
		Name:        program.ProgramDetails.Name,
		Description: program.ProgramDetails.Description,
		Level:       db.ProgramProgramLevel(program.ProgramDetails.Level),
	}

	if err := r.Queries.UpdateProgram(ctx, params); err != nil {
		// Check if the error is a unique violation (duplicate name)
		var pqErr *pq.Error
		if errors.As(err, &pqErr) {
			if pqErr.Code == databaseErrors.UniqueViolation {
				return errLib.New("Program name already exists", http.StatusConflict)
			}
			log.Println(fmt.Sprintf("Database error %v", err.Error()))
			return errLib.New("Database error", http.StatusInternalServerError)
		}
		return errLib.New("Internal server error", http.StatusInternalServerError)
	}

	return nil
}

func (r *Repository) List(ctx context.Context, programType string) ([]values.GetProgramValues, *errLib.CommonError) {

	if !db.ProgramProgramType(programType).Valid() {

		validTypes := db.AllProgramProgramTypeValues()

		return nil, errLib.New(fmt.Sprintf("Invalid program type. Valid types are: %v", validTypes), http.StatusBadRequest)
	}

	dbPrograms, err := r.Queries.GetPrograms(ctx, db.NullProgramProgramType{
		ProgramProgramType: db.ProgramProgramType(programType),
		Valid:              true,
	})

	if err != nil {

		log.Println("Error getting programs: ", err)
		dbErr := errLib.New("Internal server error", http.StatusInternalServerError)

		return nil, dbErr
	}

	programs := make([]values.GetProgramValues, len(dbPrograms))

	for i, dbProgram := range dbPrograms {
		programs[i] = values.GetProgramValues{
			ID:        dbProgram.ID,
			CreatedAt: dbProgram.CreatedAt,
			UpdatedAt: dbProgram.UpdatedAt,
			ProgramDetails: values.ProgramDetails{
				Name:        dbProgram.Name,
				Description: dbProgram.Description,
				Level:       string(dbProgram.Level),
				Type:        string(dbProgram.Type),
			},
		}
	}

	return programs, nil
}

func (r *Repository) Delete(c context.Context, id uuid.UUID) *errLib.CommonError {
	row, err := r.Queries.DeleteProgram(c, id)

	if err != nil {
		return errLib.New("Internal server error", http.StatusInternalServerError)
	}

	if row == 0 {
		return errLib.New("Program not found", http.StatusNotFound)
	}

	return nil
}

func (r *Repository) Create(c context.Context, details values.CreateProgramValues) *errLib.CommonError {

	dbPracticeParams := db.CreateProgramParams{
		Name:        details.Name,
		Description: details.Description,
		Level:       db.ProgramProgramLevel(details.Level),
		Type:        db.ProgramProgramType(details.Type),
	}

	if err := r.Queries.CreateProgram(c, dbPracticeParams); err != nil {
		// Check if the error is a unique violation (error code 23505)
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == databaseErrors.UniqueViolation {
			// Return a custom error for unique violation
			return errLib.New("Program name already exists", http.StatusConflict)
		}

		// Return a generic internal server error for other cases
		return errLib.New("Internal server error", http.StatusInternalServerError)
	}

	return nil
}
