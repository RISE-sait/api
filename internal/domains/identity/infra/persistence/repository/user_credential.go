package repository

import (
	db "api/internal/domains/identity/infra/persistence/sqlc/generated"
	errLib "api/internal/libs/errors"
	"context"
	"database/sql"
)

type UserCredentialsRepository struct {
	Queries *db.Queries
}

func NewUserCredentialsRepository(q *db.Queries) *UserCredentialsRepository {
	return &UserCredentialsRepository{
		Queries: q,
	}
}

func (r *UserRepository) IsValidUser(ctx context.Context, email, password string) bool {

	params := db.GetUserByEmailPasswordParams{
		Email: email,
		HashedPassword: sql.NullString{
			String: password,
			Valid:  true,
		},
	}

	if _, err := r.Queries.GetUserByEmailPassword(ctx, params); err != nil {
		return false
	}
	return true
}

func (r *UserCredentialsRepository) CreatePasswordTx(ctx context.Context, tx *sql.Tx, email, password string) *errLib.CommonError {

	params := db.CreatePasswordParams{
		Email: email,
		HashedPassword: sql.NullString{
			String: password,
			Valid:  password != "",
		},
	}

	rows, err := r.Queries.WithTx(tx).CreatePassword(ctx, params)

	if err != nil {
		return errLib.TranslateDBErrorToCommonError(err)
	}

	if rows != 1 {
		return errLib.New("Failed to create username password", 500)
	}

	return nil
}
