package repositories

import (
	"api/internal/types"
	"api/internal/utils"
	db "api/sqlc"
	"context"
	"database/sql"
	"net/http"
)

type UserOptionalInfoRepository struct {
	Queries *db.Queries
}

func (r *UserOptionalInfoRepository) CreateUserOptionalInfo(c context.Context, params *db.CreateUserOptionalInfoParams) *types.HTTPError {
	row, err := r.Queries.CreateUserOptionalInfo(c, *params)

	if err != nil {
		return utils.MapDatabaseError(err)
	}

	if row == 0 {
		return utils.CreateHTTPError("Error creating user optional info", 500)
	}

	return nil
}

func (r *UserOptionalInfoRepository) IsUserOptionalInfoExist(c context.Context, param *db.GetUserOptionalInfoParams) bool {

	if _, err := r.Queries.GetUserOptionalInfo(c, *param); err != nil {
		return false
	}
	return true
}

func (r *UserOptionalInfoRepository) UpdateUsername(c context.Context, param *db.UpdateUsernameParams) *types.HTTPError {
	row, err := r.Queries.UpdateUsername(c, *param)

	if err != nil {
		return utils.MapDatabaseError(err)
	}

	if row == 0 {
		return utils.CreateHTTPError("No username updated", http.StatusNotFound)
	}

	return nil
}

func (r *UserOptionalInfoRepository) UpdateUserPassword(c context.Context, param *db.UpdateUserPasswordParams) *types.HTTPError {
	row, err := r.Queries.UpdateUserPassword(c, *param)

	if err != nil {
		return utils.MapDatabaseError(err)
	}

	if row == 0 {
		return utils.CreateHTTPError("No password updated", http.StatusNotFound)
	}

	return nil
}

func (r *UserOptionalInfoRepository) CreateUserOptionalInfoTx(ctx context.Context, tx *sql.Tx, params db.CreateUserOptionalInfoParams) *types.HTTPError {
	// Create a new Queries instance bound to the transaction.
	txQueries := r.Queries.WithTx(tx)

	row, err := txQueries.CreateUserOptionalInfo(ctx, params)

	if err != nil {
		return utils.MapDatabaseError(err)
	}

	if row == 0 {
		return utils.CreateHTTPError("Error creating user optional info", 500)
	}

	return nil
}
