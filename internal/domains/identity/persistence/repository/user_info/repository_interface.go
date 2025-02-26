package user_info_temp_repo

import (
	"api/internal/domains/identity/entity"
	errLib "api/internal/libs/errors"
	"context"
	"database/sql"
	"github.com/google/uuid"
)

type InfoTempRepositoryInterface interface {
	CreateTempUserInfoTx(ctx context.Context, tx *sql.Tx, userId uuid.UUID, firstName, lastName string, email, parentHubspotId *string, age int) *errLib.CommonError
	DeleteTempUserInfoTx(ctx context.Context, tx *sql.Tx, id uuid.UUID) *errLib.CommonError
	GetTempUserInfoByEmail(ctx context.Context, email string) (*entity.UserInfo, *errLib.CommonError)
}
