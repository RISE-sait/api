package repository

import (
	db "api/internal/domains/identity/registration/infra/persistence/sqlc/generated"
	errLib "api/internal/libs/errors"
	"context"
	"database/sql"
)

type UserPasswordRepository struct {
	Queries *db.Queries
}

func NewUserPasswordRepository(q *db.Queries) *UserPasswordRepository {
	return &UserPasswordRepository{
		Queries: q,
	}
}

func (r *UserPasswordRepository) CreatePasswordTx(ctx context.Context, tx *sql.Tx, email, password string) *errLib.CommonError {

	params := db.CreatePasswordParams{
		Email: email,
		HashedPassword: sql.NullString{
			String: password,
			Valid:  password != "",
		},
	}

	rows, err := r.Queries.CreatePassword(ctx, params)

	if err != nil {
		return errLib.TranslateDBErrorToCommonError(err)
	}

	if rows != 1 {
		return errLib.New("Failed to create username password", 500)
	}

	return nil
}
