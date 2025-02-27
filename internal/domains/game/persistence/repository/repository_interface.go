package game

import (
	values "api/internal/domains/game/values"
	errLib "api/internal/libs/errors"
	"context"
	"github.com/google/uuid"
)

type RepositoryInterface interface {
	CreateGame(ctx context.Context, input values.CreateGameValue) (values.ReadValue, *errLib.CommonError)
	GetGameById(ctx context.Context, id uuid.UUID) (values.ReadValue, *errLib.CommonError)
	GetGames(ctx context.Context) ([]values.ReadValue, *errLib.CommonError)
	UpdateGame(ctx context.Context, input values.UpdateGameValue) (values.ReadValue, *errLib.CommonError)
	DeleteGame(ctx context.Context, id uuid.UUID) *errLib.CommonError
}
