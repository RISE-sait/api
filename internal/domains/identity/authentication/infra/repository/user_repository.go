package repository

import (
	db2 "api/internal/domains/identity/authentication/infra/sqlc/generated"
	"context"
	"database/sql"
)

type UserRepository struct {
	Queries *db2.Queries
}

func NewUserRepository(q *db2.Queries) *UserRepository {
	return &UserRepository{
		Queries: q,
	}
}

func (r *UserRepository) IsValidUser(ctx context.Context, email, password string) bool {

	params := db2.GetUserByEmailPasswordParams{
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
