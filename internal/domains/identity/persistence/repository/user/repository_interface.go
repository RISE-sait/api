package user

import (
	errLib "api/internal/libs/errors"
	"context"
	"database/sql"
	"github.com/google/uuid"
)

type IRepository interface {
	CreateUserTx(ctx context.Context, tx *sql.Tx) (*uuid.UUID, *errLib.CommonError)
	GetUserIdByHubspotId(ctx context.Context, id string) (uuid.UUID, *errLib.CommonError)
	UpdateUserHubspotIdTx(ctx context.Context, tx *sql.Tx, userId uuid.UUID, hubspotId string) *errLib.CommonError
}
