package staffs

import (
	"api/internal/types"
	"api/internal/utils"
	db "api/sqlc"
	"context"
)

type Repository struct {
	Queries *db.Queries
}

func (r *Repository) CreateStaff(c context.Context, params *db.CreateStaffParams) *types.HTTPError {
	row, err := r.Queries.CreateStaff(c, *params)

	if err != nil {
		return utils.MapDatabaseError(err)
	}

	if row == 0 {
		return utils.CreateHTTPError("Error creating staff", 500)
	}

	return nil
}

func (r *Repository) GetStaffByEmail(c context.Context, email string) (*db.GetStaffByEmailRow, *types.HTTPError) {
	staff, err := r.Queries.GetStaffByEmail(c, email)

	if err != nil {
		return nil, utils.MapDatabaseError(err)
	}

	return &staff, nil
}

func (r *Repository) GetAllStaff(c context.Context) (*[]db.Staff, *types.HTTPError) {
	staff, err := r.Queries.GetAllStaff(c)

	if err != nil {
		return nil, utils.MapDatabaseError(err)
	}

	return &staff, nil
}

func (r *Repository) UpdateStaff(c context.Context, params db.UpdateStaffParams) *types.HTTPError {
	row, err := r.Queries.UpdateStaff(c, params)

	if err != nil {
		return utils.MapDatabaseError(err)
	}

	if row == 0 {
		return utils.CreateHTTPError("No staff found with the associated id", 404)
	}

	return nil
}

func (r *Repository) RemoveStaff(c context.Context, email string) *types.HTTPError {
	row, err := r.Queries.DeleteStaff(c, email)

	if err != nil {
		return utils.MapDatabaseError(err)
	}

	if row == 0 {
		return utils.CreateHTTPError("No staff found with the associated email", 404)
	}

	return nil
}
