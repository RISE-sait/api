package user

import (
	values "api/internal/domains/user/values"
	errLib "api/internal/libs/errors"
	"context"
	"github.com/google/uuid"
)

type RepositoryInterface interface {
	GetUserIDByHubSpotId(ctx context.Context, id string) (*uuid.UUID, *errLib.CommonError)
	GetUsers(ctx context.Context) ([]values.ReadValue, *errLib.CommonError)
	UpdateStats(ctx context.Context, valuesToUpdate values.StatsUpdateValue) *errLib.CommonError
}
