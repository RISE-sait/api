package game

import (
	databaseErrors "api/internal/constants"
	db "api/internal/domains/game/persistence/sqlc/generated"
	values "api/internal/domains/game/values"
	errLib "api/internal/libs/errors"
	"api/internal/services/gcp"
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

func NewGameRepository(dbQueries *db.Queries) *Repository {
	return &Repository{
		Queries: dbQueries,
	}
}

func (r *Repository) GetGameById(ctx context.Context, id uuid.UUID) (values.ReadValue, *errLib.CommonError) {
	dbGame, err := r.Queries.GetGameById(ctx, id)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return values.ReadValue{}, errLib.New("Course not found", http.StatusNotFound)
		}
		return values.ReadValue{}, errLib.New("Internal server error", http.StatusInternalServerError)
	}

	game := values.ReadValue{
		ID: dbGame.ID,
		BaseValue: values.BaseValue{
			Name: dbGame.Name,
		},
	}

	videoLink := gcp.GeneratePublicFileURL(fmt.Sprintf("games/%v", game.ID))

	game.VideoLink = &videoLink

	return game, nil
}

func (r *Repository) UpdateGame(ctx context.Context, value values.UpdateGameValue) (values.ReadValue, *errLib.CommonError) {

	updateParams := db.UpdateGameParams{
		ID: value.ID,
		Name: sql.NullString{
			String: value.Name,
			Valid:  true,
		},
	}

	updatedGame, err := r.Queries.UpdateGame(ctx, updateParams)

	if err != nil {
		// Check if the error is a unique violation (duplicate name)
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" {
			return values.ReadValue{}, errLib.New("Game name already exists", http.StatusConflict)
		}
		return values.ReadValue{}, errLib.New("Internal server error", http.StatusInternalServerError)
	}

	return values.ReadValue{
		ID: updatedGame.ID,
		BaseValue: values.BaseValue{
			Name: updatedGame.Name,
		},
	}, nil
}

func (r *Repository) GetGames(ctx context.Context) ([]values.ReadValue, *errLib.CommonError) {

	dbGames, err := r.Queries.GetGames(ctx, sql.NullString{
		Valid: false,
	})

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
				Name: dbGame.Name,
			},
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

func (r *Repository) CreateGame(c context.Context, name string) (values.ReadValue, *errLib.CommonError) {

	//params := db.CreateGameParams{
	//	Name: details.Name,
	//}
	//
	//if details.VideoLink != nil {
	//	params.VideoLink = sql.NullString{
	//		String: *details.VideoLink,
	//		Valid:  true,
	//	}
	//}

	createdGame, err := r.Queries.CreateGame(c, name)

	if err != nil {
		// Check if the error is a unique violation (error code 23505)
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == databaseErrors.UniqueViolation {
			// Return a custom error for unique violation
			return values.ReadValue{}, errLib.New("Game name already exists", http.StatusConflict)
		}

		// Return a generic internal server error for other cases
		log.Println("error creating createdGame: ", err)
		return values.ReadValue{}, errLib.New("Internal server error", http.StatusInternalServerError)
	}

	return values.ReadValue{
		ID: createdGame.ID,
		BaseValue: values.BaseValue{
			Name: createdGame.Name,
			//VideoLink: &createdGame.VideoLink.String,
		},
	}, nil
}
