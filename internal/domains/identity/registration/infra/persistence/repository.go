package waiver_repository

import (
	db "api/internal/domains/identity/registration/infra/persistence/sqlc"
	errLib "api/internal/libs/errors"
	"context"
	"database/sql"
)

type WaiverRepository struct {
	Queries *db.Queries
}

func NewWaiverRepository(q *db.Queries) *WaiverRepository {
	return &WaiverRepository{
		Queries: q,
	}
}

func (r *WaiverRepository) CreateWaiverRecordTx(ctx context.Context, tx *sql.Tx, email, waiverUrl string, isSigned bool) *errLib.CommonError {

	txQueries := r.Queries.WithTx(tx)

	params := db.CreateWaiverSignedStatusParams{
		Email:     email,
		WaiverUrl: waiverUrl,
		IsSigned:  isSigned,
	}

	row, err := txQueries.CreateWaiverSignedStatus(ctx, params)

	if err != nil {
		return errLib.TranslateDBErrorToCommonError(err)
	}

	if row != 1 {
		return errLib.New("Failed to create waiver record", 500)
	}

	return nil
}
