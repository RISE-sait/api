package user_info_temp_repo

import (
	values "api/internal/domains/identity/values"
	errLib "api/internal/libs/errors"
	"context"
	"database/sql"
	"github.com/google/uuid"
)

type IPendingUsersRepository interface {
	CreatePendingUserInfoTx(ctx context.Context, tx *sql.Tx, firstName, lastName string, email, parentHubspotId *string, age int) (uuid.UUID, *errLib.CommonError)
	DeletePendingUserInfoTx(ctx context.Context, tx *sql.Tx, id uuid.UUID) *errLib.CommonError
	GetPendingUserInfoByEmail(ctx context.Context, email string) (values.PendingUserReadValues, *errLib.CommonError)
}
