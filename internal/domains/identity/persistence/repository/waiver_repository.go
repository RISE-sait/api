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

type WaiverRepository struct {
	Queries *db.Queries
}

func NewWaiverRepository(q *db.Queries) *WaiverRepository {
	return &WaiverRepository{
		Queries: q,
	}
}

func (r *WaiverRepository) CreateWaiverRecordTx(ctx context.Context, tx *sql.Tx, userId uuid.UUID, waiverUrl string, isSigned bool) *errLib.CommonError {

	txQueries := r.Queries.WithTx(tx)

	params := db.CreateWaiverSignedStatusParams{
		UserID:    userId,
		WaiverUrl: waiverUrl,
		IsSigned:  isSigned,
	}

	// Insert the waiver record
	_, err := txQueries.CreateWaiverSignedStatus(ctx, params)

	if err != nil {
		// Check if error is pq.Error (PostgreSQL specific errors)
		var pqErr *pq.Error
		if errors.As(err, &pqErr) {
			switch pqErr.Code {
			case database_errors.ForeignKeyViolation:
				log.Printf("Foreign key violation: %v", pqErr.Message)
				return errLib.New("User not found for the provided userId", http.StatusBadRequest)
			case database_errors.UniqueViolation:
				log.Printf("Unique violation: %v", pqErr.Message)
				return errLib.New("Waiver for this user already exists", http.StatusConflict)
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

func (r *WaiverRepository) GetWaiver(ctx context.Context, url string) (*db.Waiver, *errLib.CommonError) {
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
