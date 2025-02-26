package waiver_signing

import (
	databaseerrors "api/internal/constants"
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

type Repository struct {
	Queries *db.Queries
}

func NewWaiverSigningRepository(container *di.Container) *Repository {
	return &Repository{
		Queries: container.Queries.IdentityDb,
	}
}

var _ RepositoryInterface = (*Repository)(nil)

func (r *Repository) GetWaiver(ctx context.Context, url string) (*db.Waiver, *errLib.CommonError) {
	waiver, err := r.Queries.GetWaiver(ctx, url)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Printf("Waiver not found for URL: %s", url)
			return nil, errLib.New("Waiver not found", http.StatusNotFound)
		}

		log.Printf("Error fetching waiver for URL %s: %v", url, err)
		return nil, errLib.New("Internal server error", http.StatusInternalServerError)
	}

	return &waiver, nil
}

func (r *Repository) CreateWaiverSigningRecordTx(ctx context.Context, tx *sql.Tx, userId uuid.UUID, waiverUrl string, isSigned bool) *errLib.CommonError {

	txQueries := r.Queries.WithTx(tx)

	// Get waiver
	waiver, err := txQueries.GetWaiver(ctx, waiverUrl)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Printf("Waiver not found for URL: %s", waiverUrl)
			return errLib.New("Waiver not found", http.StatusNotFound)
		}
	}

	params := db.CreateWaiverSignedStatusParams{
		UserID:   userId,
		WaiverID: waiver.ID,
		IsSigned: isSigned,
	}

	// Insert the waiver record
	_, err = txQueries.CreateWaiverSignedStatus(ctx, params)

	if err != nil {
		// Check if error is pq.Error (PostgreSQL specific errors)
		var pqErr *pq.Error
		if errors.As(err, &pqErr) {
			switch pqErr.Code {
			case databaseerrors.ForeignKeyViolation:
				log.Printf("Foreign key violation: %v", pqErr.Message)
				return errLib.New("User not found for the provided user id. Or waiver not found", http.StatusBadRequest)
			case databaseerrors.UniqueViolation:
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
