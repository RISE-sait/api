package waiver_signing

import (
	"context"
	"database/sql"

	"api/internal/domains/identity/persistence/sqlc/generated"
	"api/internal/libs/errors"
	"github.com/google/uuid"
)

type RepositoryInterface interface {
	GetWaiver(ctx context.Context, url string) (*db.Waiver, *errLib.CommonError)
	CreateWaiverSigningRecordTx(ctx context.Context, tx *sql.Tx, userId uuid.UUID, waiverUrl string, isSigned bool) *errLib.CommonError
}
