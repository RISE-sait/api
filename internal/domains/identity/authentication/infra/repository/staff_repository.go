package repository

import (
	db "api/internal/domains/identity/authentication/infra/sqlc/generated"
	"api/internal/libs/errors"
	"context"
)

type StaffRepository struct {
	Queries *db.Queries
}

func NewStaffRepository(q *db.Queries) *StaffRepository {
	return &StaffRepository{
		Queries: q,
	}
}

func (r *StaffRepository) GetStaffByEmail(ctx context.Context, email string) (*db.GetStaffByEmailRow, *errLib.CommonError) {
	staff, err := r.Queries.GetStaffByEmail(ctx, email)

	if err != nil {
		return nil, errLib.TranslateDBErrorToCommonError(err)
	}

	return &staff, nil
}
