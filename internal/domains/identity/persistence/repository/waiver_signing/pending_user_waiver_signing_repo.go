package waiver_signing

import (
	databaseErrors "api/internal/constants"
	"api/internal/di"
	db "api/internal/domains/identity/persistence/sqlc/generated"
	errLib "api/internal/libs/errors"
	"context"
	"database/sql"
	"errors"
	"github.com/google/uuid"
	"log"
	"net/http"

	"github.com/lib/pq"
)

type PendingUserWaiverSigningRepository struct {
	Queries *db.Queries
}

func NewPendingUserWaiverSigningRepository(container *di.Container) *PendingUserWaiverSigningRepository {
	return &PendingUserWaiverSigningRepository{
		Queries: container.Queries.IdentityDb,
	}
}

func (r *PendingUserWaiverSigningRepository) CreateWaiverSigningRecordTx(ctx context.Context, tx *sql.Tx, userId uuid.UUID, waiverUrl string, isSigned bool) *errLib.CommonError {

	txQueries := r.Queries.WithTx(tx)

	// Get waiver
	waiver, err := txQueries.GetWaiver(ctx, waiverUrl)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Printf("Waiver not found for URL: %s", waiverUrl)
			return errLib.New("Waiver not found", http.StatusNotFound)
		}
	}

	params := db.CreatePendingUserWaiverSignedStatusParams{
		UserID:   userId,
		WaiverID: waiver.ID,
		IsSigned: isSigned,
	}

	// Insert the waiver record
	_, err = txQueries.CreatePendingUserWaiverSignedStatus(ctx, params)

	if err != nil {
		// Check if error is pq.Error (PostgreSQL specific errors)
		var pqErr *pq.Error
		if errors.As(err, &pqErr) {
			switch pqErr.Code {
			case databaseErrors.ForeignKeyViolation:
				log.Printf("Foreign key violation: %v", pqErr.Message)
				return errLib.New("User not found for the provided user id. Or waiver not found", http.StatusBadRequest)
			case databaseErrors.UniqueViolation:
				log.Printf("Unique violation: %v", pqErr.Message)
				return errLib.New("Waiver for this user id already exists", http.StatusConflict)
			default:
				log.Printf("Unhandled database error: %v", pqErr)
				return errLib.New("Internal server error", http.StatusInternalServerError)
			}
		}

		// Log and handle any other database error
		log.Printf("Unhandled database error: %v", err)
		return errLib.New("Internal server error", http.StatusInternalServerError)
	}

	return nil
}
