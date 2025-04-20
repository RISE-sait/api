package game

import (
	databaseErrors "api/internal/constants"
	"api/internal/di"
	db "api/internal/domains/game/persistence/sqlc/generated"
	values "api/internal/domains/game/values"
	errLib "api/internal/libs/errors"
	"context"
	"database/sql"
	"errors"
	"log"
	"net/http"

	"github.com/lib/pq"

	"github.com/google/uuid"
)

type Repository struct {
	Queries *db.Queries
	Tx      *sql.Tx
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

func NewGameRepository(container *di.Container) *Repository {
	return &Repository{
		Queries: container.Queries.GameDb,
	}
}

func (r *Repository) GetGameById(ctx context.Context, id uuid.UUID) (values.ReadValue, *errLib.CommonError) {
	dbGame, err := r.Queries.GetGameById(ctx, id)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return values.ReadValue{}, errLib.New("Game not found", http.StatusNotFound)
		}
		log.Printf("Error getting game: %v", err)
		return values.ReadValue{}, errLib.New("Internal server error", http.StatusInternalServerError)
	}

	game := values.ReadValue{
		ID: dbGame.ID,
		BaseValue: values.BaseValue{
			Name:         dbGame.Name,
			Description:  dbGame.Description,
			WinTeamID:    dbGame.WinTeamID,
			WinTeamName:  dbGame.WinTeamName,
			LoseTeamID:   dbGame.LoseTeamID,
			LoseTeamName: dbGame.LoseTeamName,
			WinScore:     dbGame.WinScore,
			LoseScore:    dbGame.LoseScore,
		},
		CreatedAt: dbGame.CreatedAt,
		UpdatedAt: dbGame.UpdatedAt,
	}

	return game, nil
}

func (r *Repository) UpdateGame(ctx context.Context, value values.UpdateGameValue) *errLib.CommonError {

	updateParams := db.UpdateGameParams{
		ID:        value.ID,
		Name:      value.Name,
		WinTeam:   value.WinTeamID,
		LoseTeam:  value.LoseTeamID,
		WinScore:  value.WinScore,
		LoseScore: value.LoseScore,
	}

	affectedRows, err := r.Queries.UpdateGame(ctx, updateParams)

	if err != nil {
		// Check if the error is a unique violation (duplicate name)
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == databaseErrors.UniqueViolation {
			return errLib.New("Game name already exists", http.StatusConflict)
		}
		return errLib.New("Internal server error", http.StatusInternalServerError)
	}

	if affectedRows == 0 {
		return errLib.New("Game not found", http.StatusNotFound)
	}

	return nil
}

func (r *Repository) GetGames(ctx context.Context) ([]values.ReadValue, *errLib.CommonError) {

	dbGames, err := r.Queries.GetGames(ctx)

	if err != nil {

		log.Println("Error getting games: ", err)
		dbErr := errLib.New("Internal server error", http.StatusInternalServerError)

		return nil, dbErr
	}

	games := make([]values.ReadValue, len(dbGames))

	for i, dbGame := range dbGames {
		games[i] = values.ReadValue{
			ID: dbGame.ID,
			BaseValue: values.BaseValue{
				Name:         dbGame.Name,
				Description:  dbGame.Description,
				WinTeamID:    dbGame.WinTeamID,
				WinTeamName:  dbGame.WinTeamName,
				LoseTeamID:   dbGame.LoseTeamID,
				LoseTeamName: dbGame.LoseTeamName,
				WinScore:     dbGame.WinScore,
				LoseScore:    dbGame.LoseScore,
			},
			CreatedAt: dbGame.CreatedAt,
			UpdatedAt: dbGame.UpdatedAt,
		}
	}

	return games, nil
}

func (r *Repository) DeleteGame(c context.Context, id uuid.UUID) *errLib.CommonError {
	row, err := r.Queries.DeleteGame(c, id)

	if err != nil {
		return errLib.New("Internal server error", http.StatusInternalServerError)
	}

	if row == 0 {
		return errLib.New("Game not found", http.StatusNotFound)
	}

	return nil
}

func (r *Repository) CreateGame(c context.Context, details values.CreateGameValue) *errLib.CommonError {

	params := db.CreateGameParams{
		Name:      details.Name,
		WinTeam:   details.WinTeamID,
		LoseTeam:  details.LoseTeamID,
		WinScore:  details.WinScore,
		LoseScore: details.LoseScore,
	}

	affectedRows, err := r.Queries.CreateGame(c, params)

	if err != nil {
		// Check if the error is a unique violation (error code 23505)
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == databaseErrors.UniqueViolation {
			// Return a custom error for unique violation
			return errLib.New("Game name already exists", http.StatusConflict)
		}

		// Return a generic internal server error for other cases
		log.Println("error creating createdGame: ", err)
		return errLib.New("Internal server error", http.StatusInternalServerError)
	}

	if affectedRows == 0 {
		return errLib.New("Game not created for unknown reason", http.StatusInternalServerError)
	}

	return nil
}
