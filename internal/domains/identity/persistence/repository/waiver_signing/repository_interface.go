package waiver_signing

import (
	"context"
	"database/sql"

	"api/internal/libs/errors"
	"github.com/google/uuid"
)

type IRepository interface {
	CreateWaiverSigningRecordTx(ctx context.Context, tx *sql.Tx, userId uuid.UUID, waiverUrl string, isSigned bool) *errLib.CommonError
}
