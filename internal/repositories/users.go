package repositories

import (
	"api/internal/utils"
	db "api/sqlc"
	"context"
	"net/http"
)

type UsersRepository struct {
	Queries *db.Queries
}

func (r *UsersRepository) CreateUser(c context.Context, email string) *utils.HTTPError {
	row, err := r.Queries.CreateUser(c, email)

	if err != nil {
		return utils.MapDatabaseError(err)
	}

	if row == 0 {
		return utils.CreateHTTPError("No row inserted", http.StatusInternalServerError)
	}

	return nil
}

func (r *UsersRepository) GetUser(c context.Context, email string) (*db.User, *utils.HTTPError) {

	user, err := r.Queries.GetUserByEmail(c, email)

	if err != nil {
		return nil, utils.MapDatabaseError(err)
	}
	return &user, nil
}

func (r *UsersRepository) UpdateUserEmail(c context.Context, params db.UpdateUserEmailParams) *utils.HTTPError {
	row, err := r.Queries.UpdateUserEmail(c, params)

	if err != nil {
		return utils.MapDatabaseError(err)
	}

	if row == 0 {
		return utils.CreateHTTPError("No user found with the associated id", http.StatusNotFound)
	}

	return nil
}
