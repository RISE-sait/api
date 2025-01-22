package repository

import (
	db2 "api/internal/domains/identity/authentication/infra/sqlc/generated"
	"api/internal/libs/errors"
	"context"
)

type StaffRepository struct {
	Queries *db2.Queries
}

func NewStaffRepository(q *db2.Queries) *StaffRepository {
	return &StaffRepository{
		Queries: q,
	}
}

func (r *StaffRepository) GetStaffByEmail(ctx context.Context, email string) (*db2.GetStaffByEmailRow, *errors.CommonError) {
	staff, err := r.Queries.GetStaffByEmail(ctx, email)

	if err != nil {
		return nil, errors.TranslateDBErrorToCommonError(err)
	}

	return &staff, nil
}
