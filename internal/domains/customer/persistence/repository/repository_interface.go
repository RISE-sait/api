package customer

import (
	errLib "api/internal/libs/errors"
	"context"
	"github.com/google/uuid"
)

type RepositoryInterface interface {
	GetUserIDByHubSpotId(ctx context.Context, id string) (*uuid.UUID, *errLib.CommonError)
}
