package repository

import (
	database_errors "api/internal/constants"
	"api/internal/di"
	db "api/internal/domains/identity/persistence/sqlc/generated"
	errLib "api/internal/libs/errors"
	"context"
	"database/sql"
	"errors"
	"log"
	"net/http"

	"github.com/lib/pq"
)

type UserCredentialsRepository struct {
	Queries *db.Queries
}

func NewUserCredentialsRepository(container *di.Container) *UserCredentialsRepository {
	return &UserCredentialsRepository{
		Queries: container.Queries.IdentityDb,
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
		log.Printf("Failed to validate user: %v", err)
		return false
	}
	return true
}

func (r *UserCredentialsRepository) CreatePasswordTx(ctx context.Context, tx *sql.Tx, email, password string) *errLib.CommonError {

	params := db.CreatePasswordParams{
		Email: email,
		HashedPassword: sql.NullString{
			String: password,
			Valid:  true,
		},
	}

	rows, err := r.Queries.WithTx(tx).CreatePassword(ctx, params)

	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) {

			if pqErr.Code == database_errors.NotNullViolation {
				return errLib.New("Either user with the email is not found, or password is null", http.StatusBadRequest)
			}

			if pqErr.Code == database_errors.UniqueViolation {
				return errLib.New("Email already exists for the credentials", http.StatusBadRequest)
			}
		}
	}

	if rows != 1 {
		return errLib.New("Failed to create email password", 500)
	}

	return nil
}
