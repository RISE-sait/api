package identity

import (
	databaseErrors "api/internal/constants"
	db "api/internal/domains/identity/persistence/sqlc/generated"
	values "api/internal/domains/identity/values"
	errLib "api/internal/libs/errors"
	"context"
	"database/sql"
	"errors"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"log"
	"net/http"
)

type WaiverSigningRepository struct {
	Queries *db.Queries
}

func NewWaiverSigningRepository(db *db.Queries) *WaiverSigningRepository {
	return &WaiverSigningRepository{
		Queries: db,
	}
}

func (r *WaiverSigningRepository) GetRequiredWaivers(ctx context.Context) ([]values.Waiver, *errLib.CommonError) {

	// Insert the waiver record
	waivers, err := r.Queries.GetRequiredWaivers(ctx)

	if err != nil {
		// Check if error is pq.Error (PostgreSQL specific errors)
		var pqErr *pq.Error
		if errors.As(err, &pqErr) {
			log.Printf("Unhandled database error: %v", pqErr)
			return nil, errLib.New("Internal server error", http.StatusInternalServerError)
		}

		// Log and handle any other database error
		log.Printf("Unhandled database error: %v", err)
		return nil, errLib.New("Internal server error", http.StatusInternalServerError)
	}

	var response []values.Waiver

	for _, waiver := range waivers {
		response = append(response, values.Waiver{
			ID:        waiver.ID,
			URL:       waiver.WaiverUrl,
			Name:      waiver.WaiverName,
			CreatedAt: waiver.CreatedAt,
			UpdatedAt: waiver.UpdatedAt,
		})
	}

	return response, nil
}

func (r *WaiverSigningRepository) CreateWaiversSigningRecordTx(ctx context.Context, tx *sql.Tx, userId []uuid.UUID, waiverUrl []string, isSigned []bool) *errLib.CommonError {

	txQueries := r.Queries.WithTx(tx)

	params := db.CreateWaiverSignedStatusParams{
		UserIDArray:    userId,
		WaiverUrlArray: waiverUrl,
		IsSignedArray:  isSigned,
	}

	// Insert the waiver record
	affectedRows, err := txQueries.CreateWaiverSignedStatus(ctx, params)

	if err != nil {
		// Check if error is pq.Error (PostgreSQL specific errors)
		var pqErr *pq.Error
		if errors.As(err, &pqErr) {
			switch pqErr.Code {
			case databaseErrors.ForeignKeyViolation:
				log.Printf("Foreign key violation: %v", pqErr.Message)
				return errLib.New("User or waiver not found.", http.StatusBadRequest)
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

	if affectedRows == 0 {
		return errLib.New("Failed to create waiver signing record for unknown reason. Contact support.", http.StatusInternalServerError)
	}

	return nil
}
