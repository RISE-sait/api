package user

import (
	databaseErrors "api/internal/constants"
	"api/internal/di"
	"api/internal/domains/identity/persistence/sqlc/generated"
	errLib "api/internal/libs/errors"
	"context"
	"database/sql"
	"errors"
	"github.com/google/uuid"
	"log"
	"net/http"

	"github.com/lib/pq"
)

// Repository provides methods to interact with the user data in the database.
type Repository struct {
	Queries *db.Queries
}

var _ IRepository = (*Repository)(nil)

// NewUserRepository creates a new instance of UserRepository with the provided dependency injection container.
func NewUserRepository(container *di.Container) *Repository {
	return &Repository{
		Queries: container.Queries.IdentityDb,
	}
}

func (r *Repository) CreateUserTx(ctx context.Context, tx *sql.Tx) (*uuid.UUID, *errLib.CommonError) {

	queries := r.Queries

	if tx != nil {
		queries = queries.WithTx(tx)
	}

	user, err := queries.CreateUser(ctx, sql.NullString{
		String: uuid.Nil.String(),
		Valid:  false,
	})

	if err != nil {
		var pqErr *pq.Error

		if errors.As(err, &pqErr) {
			// Handle unique constraint violation (e.g., duplicate email)
			if pqErr.Code == databaseErrors.UniqueViolation { // Unique violation error code
				log.Printf("Unique constraint violation: %v", pqErr.Message)
				return nil, errLib.New("Email or hubspot id already exists", http.StatusConflict)
			}
		}
		log.Printf("Unhandled error: %v", err)
		return nil, errLib.New("Internal server error", http.StatusInternalServerError)
	}

	return &user.ID, nil
}

func (r *Repository) GetUserIdByHubspotId(ctx context.Context, id string) (uuid.UUID, *errLib.CommonError) {

	user, err := r.Queries.GetUserByHubSpotId(ctx, sql.NullString{
		String: id,
		Valid:  true,
	})

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return uuid.Nil, errLib.New("User not found", http.StatusNotFound)
		}

		log.Printf("Unhandled error: %v", err)
		return uuid.Nil, errLib.New("Internal server error", http.StatusInternalServerError)
	}

	return user.ID, nil
}

func (r *Repository) UpdateUserHubspotIdTx(ctx context.Context, tx *sql.Tx, userId uuid.UUID, hubspotId string) *errLib.CommonError {

	queries := r.Queries

	if tx != nil {
		queries = queries.WithTx(tx)
	}

	updatedRows, err := queries.UpdateUserHubspotId(ctx, db.UpdateUserHubspotIdParams{
		HubspotID: sql.NullString{
			String: hubspotId,
			Valid:  true,
		},
		ID: userId,
	})

	if err != nil {
		var pqErr *pq.Error

		if errors.As(err, &pqErr) {
			// Handle unique constraint violation (e.g., duplicate email)
			if pqErr.Code == databaseErrors.UniqueViolation { // Unique violation error code
				log.Printf("Unique constraint violation: %v", pqErr.Message)
				return errLib.New("Hubspot id already exists", http.StatusConflict)
			}
		}
		log.Printf("Unhandled error: %v", err)
		return errLib.New("Internal server error", http.StatusInternalServerError)
	}

	if updatedRows == 0 {
		return errLib.New("User not found", http.StatusNotFound)
	}

	return nil
}
