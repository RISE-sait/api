package user

import (
	databaseErrors "api/internal/constants"
	dbIdentity "api/internal/domains/identity/persistence/sqlc/generated"
	errLib "api/internal/libs/errors"
	"context"
	"database/sql"
	"errors"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"log"
	"net/http"
)

func (r *UsersRepository) UpdateUserHubspotIdTx(ctx context.Context, tx *sql.Tx, userId uuid.UUID, hubspotId string) *errLib.CommonError {
	queries := r.IdentityQueries

	if tx != nil {
		queries = queries.WithTx(tx)
	}

	updatedRows, err := queries.UpdateUserHubspotId(ctx, dbIdentity.UpdateUserHubspotIdParams{
		HubspotID: sql.NullString{String: hubspotId, Valid: true},
		ID:        userId,
	})

	if err != nil {
		var pqErr *pq.Error

		if errors.As(err, &pqErr) {
			if pqErr.Code == databaseErrors.UniqueViolation {
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
