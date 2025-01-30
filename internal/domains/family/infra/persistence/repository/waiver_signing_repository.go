package repository

import (
	"api/internal/domains/family/entities"
	db "api/internal/domains/family/infra/persistence/sqlc/generated"
	errLib "api/internal/libs/errors"
	"context"
	"database/sql"
	"errors"
	"log"
	"net/http"
)

type WaiverSigningRepository struct {
	Queries *db.Queries
}

func NewWaiverSigningRepository(q *db.Queries) *WaiverSigningRepository {
	return &WaiverSigningRepository{
		Queries: q,
	}
}

func (r *WaiverSigningRepository) GetWaiverSignings(ctx context.Context, email string) ([]entities.PendingAccountsWaiverSigning, *errLib.CommonError) {

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
