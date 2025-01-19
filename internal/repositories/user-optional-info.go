package repositories

import (
	"api/internal/utils"
	db "api/sqlc"
	"context"
	"net/http"
)

type UserOptionalInfoRepository struct {
	Queries *db.Queries
}

func (r *UserOptionalInfoRepository) CreateUserOptionalInfo(c context.Context, params *db.CreateUserOptionalInfoParams) *utils.HTTPError {
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

func (r *UserOptionalInfoRepository) UpdateUsername(c context.Context, param *db.UpdateUsernameParams) *utils.HTTPError {
	row, err := r.Queries.UpdateUsername(c, *param)

	if err != nil {
		return utils.MapDatabaseError(err)
	}

	if row == 0 {
		return utils.CreateHTTPError("No username updated", http.StatusNotFound)
	}

	return nil
}

func (r *UserOptionalInfoRepository) UpdateUserPassword(c context.Context, param *db.UpdateUserPasswordParams) *utils.HTTPError {
	row, err := r.Queries.UpdateUserPassword(c, *param)

	if err != nil {
		return utils.MapDatabaseError(err)
	}

	if row == 0 {
		return utils.CreateHTTPError("No password updated", http.StatusNotFound)
	}

	return nil
}
