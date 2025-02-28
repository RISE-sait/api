package user_info_temp_repo

import (
	"api/internal/domains/identity/entity"
	errLib "api/internal/libs/errors"
	"context"
	"database/sql"
	"github.com/google/uuid"
)

type IPendingUsersRepository interface {
	CreatePendingUserInfoTx(ctx context.Context, tx *sql.Tx, firstName, lastName string, email, parentHubspotId *string, age int) (uuid.UUID, *errLib.CommonError)
	DeletePendingUserInfoTx(ctx context.Context, tx *sql.Tx, id uuid.UUID) *errLib.CommonError
	GetPendingUserInfoByEmail(ctx context.Context, email string) (*entity.UserInfo, *errLib.CommonError)
}
