package program

import (
	databaseErrors "api/internal/constants"
	"api/internal/di"
	db "api/internal/domains/program/persistence/sqlc/generated"
	"api/internal/domains/program/values"
	errLib "api/internal/libs/errors"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

type Repository struct {
	Queries *db.Queries
	Tx      *sql.Tx
}

func NewProgramRepository(container *di.Container) *Repository {
	return &Repository{
		Queries: container.Queries.ProgramDb,
	}
}

func (r *Repository) GetTx() *sql.Tx {
	return r.Tx
}

func (r *Repository) WithTx(tx *sql.Tx) *Repository {
	return &Repository{
		Queries: r.Queries.WithTx(tx),
		Tx:      tx,
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

	if !db.ProgramProgramType(program.Type).Valid() {
		validTypes := db.AllProgramProgramTypeValues()
		return errLib.New(fmt.Sprintf("Invalid program type. Valid types are: %v", validTypes), http.StatusBadRequest)
	}

	params := db.UpdateProgramParams{
		ID:          program.ID,
		Name:        program.ProgramDetails.Name,
		Description: program.ProgramDetails.Description,
		Type:        db.ProgramProgramType(program.Type),
		Level:       db.ProgramProgramLevel(program.ProgramDetails.Level),
	}

	if program.Capacity != nil {
		params.Capacity = sql.NullInt32{
			Int32: *program.Capacity,
			Valid: true,
		}
	}

	_, err := r.Queries.UpdateProgram(ctx, params)

	if err != nil {
		// Check if the error is a unique violation (duplicate name)
		var pqErr *pq.Error
		if errors.As(err, &pqErr) {
			if pqErr.Code == databaseErrors.UniqueViolation {
				return errLib.New("Program name already exists", http.StatusConflict)
			}
		}
		log.Println(fmt.Sprintf("Database error when updating program: %v", err.Error()))
		return errLib.New("Internal server error when updating program", http.StatusInternalServerError)
	}

	return nil
}

func (r *Repository) GetProgramByID(ctx context.Context, id uuid.UUID) (values.GetProgramValues, *errLib.CommonError) {

	dbProgram, err := r.Queries.GetProgramById(ctx, id)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return values.GetProgramValues{}, errLib.New("Program not found", http.StatusNotFound)
		}
		log.Printf("Error getting program: %v", err)
		return values.GetProgramValues{}, errLib.New("Internal server error", http.StatusInternalServerError)
	}

	response := values.GetProgramValues{
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

	if dbProgram.Capacity.Valid {
		response.Capacity = &dbProgram.Capacity.Int32
	}

	return response, nil
}

func (r *Repository) List(ctx context.Context, programTypeStr string) ([]values.GetProgramValues, *errLib.CommonError) {

	var params db.NullProgramProgramType

	programType := db.ProgramProgramType(programTypeStr)
	isValidProgramType := programType.Valid()

	if programTypeStr != "" {
		if !isValidProgramType {
			validTypes := db.AllProgramProgramTypeValues()
			return nil, errLib.New(fmt.Sprintf("Invalid program type. Valid types are: %v", validTypes), http.StatusBadRequest)
		}

		params.ProgramProgramType = programType
		params.Valid = true
	} else {
		params.Valid = false
	}

	dbPrograms, err := r.Queries.GetPrograms(ctx, params)

	if err != nil {

		log.Println("Error getting programs: ", err)
		dbErr := errLib.New("Internal server error", http.StatusInternalServerError)

		return nil, dbErr
	}

	programs := make([]values.GetProgramValues, len(dbPrograms))

	for i, dbProgram := range dbPrograms {
		val := values.GetProgramValues{
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

		if dbProgram.Capacity.Valid {
			val.Capacity = &dbProgram.Capacity.Int32
		}

		programs[i] = val
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

	if details.Capacity != nil {
		dbPracticeParams.Capacity = sql.NullInt32{
			Int32: *details.Capacity,
			Valid: true,
		}
	}

	_, err := r.Queries.CreateProgram(c, dbPracticeParams)

	if err != nil {
		// Check if the error is a unique violation (error code 23505)
		var pqErr *pq.Error
		if errors.As(err, &pqErr) {
			if pqErr.Code == databaseErrors.UniqueViolation {
				// Return a custom error for unique violation
				return errLib.New("Program name already exists", http.StatusConflict)
			}
		}

		log.Printf("Error creating program: %v", err)
		return errLib.New("Internal server error", http.StatusInternalServerError)
	}

	return nil
}
