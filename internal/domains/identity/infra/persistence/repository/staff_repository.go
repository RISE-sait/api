package repository

import (
	db "api/internal/domains/identity/infra/persistence/sqlc/generated"
	errLib "api/internal/libs/errors"
	"context"
	"database/sql"
	"log"
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

func (r *StaffRepository) CreateStaffTx(ctx context.Context, tx *sql.Tx, email, role string, isActive bool) *errLib.CommonError {

	params := db.CreateStaffParams{
		Email:    email,
		Role:     db.StaffRoleEnum(role),
		IsActive: isActive,
	}

	txQueries := r.Queries.WithTx(tx)

	rows, err := txQueries.CreateStaff(ctx, params)

	if err != nil {
		log.Println("Error creating staff ", err)
		return errLib.TranslateDBErrorToCommonError(err)
	}

	if rows != 1 {
		log.Println("Error creating staff ", err)
		return errLib.New("Failed to create staff", 500)
	}

	return nil
}
