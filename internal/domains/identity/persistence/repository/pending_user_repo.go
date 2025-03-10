package identity

import (
	database_errors "api/internal/constants"
	"api/internal/di"
	"api/internal/domains/identity/persistence/sqlc/generated"
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

type PendingUsersRepo struct {
	Queries *db.Queries
}

func NewPendingUserInfoRepository(container *di.Container) *PendingUsersRepo {
	return &PendingUsersRepo{
		Queries: container.Queries.IdentityDb,
	}
}

func (r *PendingUsersRepo) CreatePendingUserInfoTx(ctx context.Context, tx *sql.Tx, firstName, lastName string, hasSmsConsent, hasEmailConsent, isParent bool, countryCode, phone, email, parentHubspotId *string, age int) (uuid.UUID, *errLib.CommonError) {

	queries := r.Queries
	if tx != nil {
		queries = queries.WithTx(tx)
	}

	dbTempUserInfo := db.CreatePendingUserParams{
		FirstName:                firstName,
		LastName:                 lastName,
		Age:                      int32(age),
		HasMarketingEmailConsent: hasEmailConsent,
		HasSmsConsent:            hasSmsConsent,
		IsParent:                 isParent,
	}

	if countryCode != nil {
		dbTempUserInfo.Alpha2CountryCode = sql.NullString{String: *countryCode, Valid: true}
	}

	if phone != nil {
		dbTempUserInfo.Phone = sql.NullString{String: *phone, Valid: true}
	}

	if email != nil {
		dbTempUserInfo.Email = sql.NullString{String: *email, Valid: true}
	}

	if parentHubspotId != nil {
		dbTempUserInfo.ParentHubspotID = sql.NullString{String: *parentHubspotId, Valid: true}
	}

	user, err := queries.CreatePendingUser(ctx, dbTempUserInfo)

	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) {
			// Handle unique constraint violation
			if pqErr.Code == database_errors.UniqueViolation { // Unique violation
				log.Printf("Unique constraint violation: %v", pqErr.Message)
				return uuid.Nil, errLib.New("Email already exists", http.StatusConflict)
			}
		}
		log.Printf("Unhandled error: %v", err)
		return uuid.Nil, errLib.New("Internal server error", http.StatusInternalServerError)
	}

	return user.ID, nil
}

func (r *PendingUsersRepo) DeletePendingUserInfoTx(ctx context.Context, tx *sql.Tx, id uuid.UUID) *errLib.CommonError {
	queries := r.Queries
	if tx != nil {
		queries = queries.WithTx(tx)
	}

	deletedRows, err := queries.DeletePendingUser(ctx, id)
	if err != nil {
		log.Printf("Error deleting temp user info: %v", err)
		return errLib.New("Internal server error", http.StatusInternalServerError)
	}

	if deletedRows == 0 {
		return errLib.New("Temp user not found", http.StatusNotFound)
	}

	return nil
}

func (r *PendingUsersRepo) GetPendingUserInfoByEmail(ctx context.Context, email string) (values.PendingUserReadValues, *errLib.CommonError) {

	info, err := r.Queries.GetPendingUserByEmail(ctx, sql.NullString{
		String: email,
		Valid:  true,
	})

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// Return a "not found" error if no rows are returned
			return values.PendingUserReadValues{}, errLib.New("User info not found", http.StatusNotFound)
		}
		log.Printf("Error fetching temp user info: %v", err)
		return values.PendingUserReadValues{}, errLib.New("Internal server error", http.StatusInternalServerError)
	}

	response := values.PendingUserReadValues{
		ID:                         info.ID,
		FirstName:                  info.FirstName,
		LastName:                   info.LastName,
		HasConsentToSms:            info.HasSmsConsent,
		HasConsentToEmailMarketing: info.HasMarketingEmailConsent,
	}

	if info.Email.Valid {
		response.Email = &info.Email.String
	}

	if info.Phone.Valid {
		response.Phone = &info.Phone.String
	}

	return response, nil
}
