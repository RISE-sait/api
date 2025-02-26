package customer

import (
	db "api/internal/domains/customer/persistence/sqlc/generated"
	errLib "api/internal/libs/errors"
	"context"
	"database/sql"
	"errors"
	"github.com/google/uuid"
	"log"
	"net/http"
)

// Repository provides methods to interact with the user data in the database.
type Repository struct {
	Queries *db.Queries
}

var _ RepositoryInterface = (*Repository)(nil)

// NewUserRepository creates a new instance of UserRepository with the provided dependency injection container.
func NewUserRepository(queries *db.Queries) *Repository {
	return &Repository{
		Queries: queries,
	}
}

func (r *Repository) GetUserIDByHubSpotId(ctx context.Context, id string) (*uuid.UUID, *errLib.CommonError) {

	dbUserId, err := r.Queries.GetUserIDByHubSpotId(ctx, sql.NullString{
		String: id,
		Valid:  true,
	})

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errLib.New("User not found", http.StatusNotFound)
		}

		log.Printf("Unhandled error: %v", err)
		return nil, errLib.New("Internal server error", http.StatusInternalServerError)
	}

	return &dbUserId, nil
}
