package repository

import (
	database_errors "api/internal/constants"
	db "api/internal/domains/identity/persistence/sqlc/generated"
	errLib "api/internal/libs/errors"
	"context"
	"database/sql"
	"errors"
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

type UserRepository struct {
	Queries *db.Queries
}

func NewUserRepository(q *db.Queries) *UserRepository {
	return &UserRepository{
		Queries: q,
	}
}

func (r *UserRepository) CreateUserTx(ctx context.Context, tx *sql.Tx, email string) (uuid.UUID, *errLib.CommonError) {

	id, err := r.Queries.WithTx(tx).CreateUser(ctx, email)

	if err != nil {
		var pqErr *pq.Error

		if errors.As(err, &pqErr) {
			// Handle unique constraint violation (e.g., duplicate email)
			if pqErr.Code == database_errors.UniqueViolation { // Unique violation error code
				log.Printf("Unique constraint violation: %v", pqErr.Message)
				return uuid.Nil, errLib.New("Email already exists", http.StatusConflict)
			}
		}
		log.Printf("Unhandled error: %v", err)
		return uuid.Nil, errLib.New("Internal server error", http.StatusInternalServerError)
	}

	if id == uuid.Nil {
		return uuid.Nil, errLib.New("Failed to create user", 500)
	}

	return id, nil
}
