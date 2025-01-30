package repository

import (
	"api/cmd/server/di"
	database_errors "api/internal/constants"
	"api/internal/domains/identity/entities"
	db "api/internal/domains/identity/persistence/sqlc/generated"
	errLib "api/internal/libs/errors"
	"context"
	"database/sql"
	"errors"
	"log"
	"net/http"

	"github.com/lib/pq"
)

type PendingChildAccountWaiverSigningRepository struct {
	Queries *db.Queries
}

func NewPendingChildAccountWaiverSigningRepository(container *di.Container) *PendingChildAccountWaiverSigningRepository {
	return &PendingChildAccountWaiverSigningRepository{
		Queries: container.Queries.IdentityDb,
	}
}

func (r *PendingChildAccountWaiverSigningRepository) GetWaiverSignings(ctx context.Context, email string) ([]entities.PendingAccountsWaiverSigning, *errLib.CommonError) {

	// Insert the waiver record
	waiverSignings, err := r.Queries.GetPendingChildAccountWaiverSigning(ctx, email)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return []entities.PendingAccountsWaiverSigning{}, errLib.New("Waivers with associated email not found", http.StatusNotFound)
		}

		// Log and handle any other database error
		log.Printf("Unhandled database error: %v", err)
		return []entities.PendingAccountsWaiverSigning{}, errLib.New("Internal server error", http.StatusInternalServerError)
	}

	results := make([]entities.PendingAccountsWaiverSigning, 0, len(waiverSignings))

	for _, waiverSigning := range waiverSignings {
		results = append(results, entities.PendingAccountsWaiverSigning{
			UserID:    waiverSigning.UserID,
			WaiverID:  waiverSigning.WaiverID,
			IsSigned:  waiverSigning.IsSigned,
			UpdatedAt: waiverSigning.UpdatedAt,
			WaiverUrl: waiverSigning.WaiverUrl,
		})
	}
	return results, nil

}

func (r *PendingChildAccountWaiverSigningRepository) CreateWaiverSigningRecordTx(ctx context.Context, tx *sql.Tx, email string, waiverUrl string, isSigned bool) *errLib.CommonError {

	txQueries := r.Queries.WithTx(tx)

	params := db.CreatePendingChildAccountWaiverSigningParams{
		UserEmail: email,
		WaiverUrl: waiverUrl,
		IsSigned:  isSigned,
	}

	// Insert the waiver record
	_, err := txQueries.CreatePendingChildAccountWaiverSigning(ctx, params)

	if err != nil {
		// Check if error is pq.Error (PostgreSQL specific errors)
		var pqErr *pq.Error
		if errors.As(err, &pqErr) {
			switch pqErr.Code {
			case database_errors.ForeignKeyViolation:
				log.Printf("Foreign key violation: %v", pqErr.Message)
				return errLib.New("User not found for the provided email. Or waiver not found", http.StatusBadRequest)
			case database_errors.UniqueViolation:
				log.Printf("Unique violation: %v", pqErr.Message)
				return errLib.New("Waiver for this email already exists", http.StatusConflict)
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

func (r *PendingChildAccountWaiverSigningRepository) DeletePendingWaiverSigningRecordByChildEmailTx(ctx context.Context, tx *sql.Tx, email string) *errLib.CommonError {

	txQueries := r.Queries.WithTx(tx)

	// Insert the waiver record
	_, err := txQueries.DeletePendingChildAccountWaiverSigning(ctx, email)

	if err != nil {
		// Log and handle any other database error
		log.Printf("Error deleting waiver signing record. Email : %v. Err: %v", email, err)
		return errLib.New("Internal server error", http.StatusInternalServerError)
	}

	return nil
}
