package repository

import (
	db "api/internal/domains/identity/registration/infra/persistence/sqlc/generated"
	errLib "api/internal/libs/errors"
	"context"
	"database/sql"
)

type StaffRepository struct {
	Queries *db.Queries
}

func NewStaffRepository(q *db.Queries) *StaffRepository {
	return &StaffRepository{
		Queries: q,
	}
}

func (r *StaffRepository) CreateStaffTx(ctx context.Context, tx *sql.Tx, role string, isActive bool) *errLib.CommonError {

	params := db.CreateStaffParams{
		Role:     db.StaffRoleEnum(role),
		IsActive: isActive,
	}

	txQueries := r.Queries.WithTx(tx)

	rows, err := txQueries.CreateStaff(ctx, params)

	if err != nil {
		return errLib.TranslateDBErrorToCommonError(err)
	}

	if rows != 1 {
		return errLib.New("Failed to create staff", 500)
	}

	return nil
}
