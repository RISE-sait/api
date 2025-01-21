package users

import (
	"api/internal/types"
	"api/internal/utils"
	db "api/sqlc"
	"context"
	"database/sql"
	"net/http"
)

type Repository struct {
	Queries *db.Queries
}

func (r *Repository) CreateUser(c context.Context, email string) *types.HTTPError {
	row, err := r.Queries.CreateUser(c, email)

	if err != nil {
		return utils.MapDatabaseError(err)
	}

	if row == 0 {
		return utils.CreateHTTPError("No row inserted", http.StatusInternalServerError)
	}

	return nil
}

func (r *Repository) GetUser(c context.Context, email string) (*db.User, *types.HTTPError) {

	user, err := r.Queries.GetUserByEmail(c, email)

	if err != nil {
		return nil, utils.MapDatabaseError(err)
	}
	return &user, nil
}

func (r *Repository) UpdateUserEmail(c context.Context, params db.UpdateUserEmailParams) *types.HTTPError {
	row, err := r.Queries.UpdateUserEmail(c, params)

	if err != nil {
		return utils.MapDatabaseError(err)
	}

	if row == 0 {
		return utils.CreateHTTPError("No user found with the associated id", http.StatusNotFound)
	}

	return nil
}

func (r *Repository) CreateUserTx(ctx context.Context, tx *sql.Tx, email string) *types.HTTPError {
	txQueries := r.Queries.WithTx(tx)
	row, err := txQueries.CreateUser(ctx, email)

	if err != nil {
		return utils.MapDatabaseError(err)
	}

	if row == 0 {
		return utils.CreateHTTPError("No row inserted", http.StatusInternalServerError)
	}

	return nil
}
