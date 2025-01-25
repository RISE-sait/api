package repository

import (
	db "api/internal/domains/identity/registration/infra/persistence/sqlc/generated"
	errLib "api/internal/libs/errors"
	"context"
	"database/sql"
)

type UserRepository struct {
	Queries *db.Queries
}

func NewUserRepository(q *db.Queries) *UserRepository {
	return &UserRepository{
		Queries: q,
	}
}

func (r *UserRepository) CreateUserTx(ctx context.Context, tx *sql.Tx, email string) *errLib.CommonError {

	rows, err := r.Queries.CreateUser(ctx, email)

	if err != nil {
		return errLib.TranslateDBErrorToCommonError(err)
	}

	if rows != 1 {
		return errLib.New("Failed to create user", 500)
	}

	return nil
}
