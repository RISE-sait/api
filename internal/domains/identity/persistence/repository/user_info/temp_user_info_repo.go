package user_info_temp_repo

import (
	"api/internal/di"
	"api/internal/domains/identity/entity"
	"api/internal/domains/identity/persistence/sqlc/generated"
	errLib "api/internal/libs/errors"
	"context"
	"database/sql"
	"errors"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"log"
	"net/http"
)

type InfoTempRepo struct {
	Queries *db.Queries
}

func NewInfoTempRepository(container *di.Container) *InfoTempRepo {
	return &InfoTempRepo{
		Queries: container.Queries.IdentityDb,
	}
}

var _ InfoTempRepositoryInterface = (*InfoTempRepo)(nil)

func (r *InfoTempRepo) CreateTempUserInfoTx(ctx context.Context, tx *sql.Tx, userId uuid.UUID, firstName, lastName string, email, parentHubspotId *string, age int) *errLib.CommonError {

	queries := r.Queries
	if tx != nil {
		queries = queries.WithTx(tx)
	}

	dbTempUserInfo := db.CreateTempUserInfoParams{
		ID:        userId,
		FirstName: firstName,
		LastName:  lastName,
		Age:       int32(age),
	}

	if email != nil {
		dbTempUserInfo.Email = sql.NullString{String: *email, Valid: true}
	}

	if parentHubspotId != nil {
		dbTempUserInfo.ParentHubspotID = sql.NullString{String: *parentHubspotId, Valid: true}
	}

	_, err := queries.CreateTempUserInfo(ctx, dbTempUserInfo)

	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) {
			// Handle unique constraint violation
			if pqErr.Code == "23505" { // Unique violation
				log.Printf("Unique constraint violation: %v", pqErr.Message)
				return errLib.New("Email already exists", http.StatusConflict)
			}
		}
		log.Printf("Unhandled error: %v", err)
		return errLib.New("Internal server error", http.StatusInternalServerError)
	}

	return nil
}

func (r *InfoTempRepo) DeleteTempUserInfoTx(ctx context.Context, tx *sql.Tx, id uuid.UUID) *errLib.CommonError {
	queries := r.Queries
	if tx != nil {
		queries = queries.WithTx(tx)
	}

	deletedRows, err := queries.DeleteTempUserInfo(ctx, id)
	if err != nil {
		log.Printf("Error deleting temp user info: %v", err)
		return errLib.New("Internal server error", http.StatusInternalServerError)
	}

	if deletedRows == 0 {
		return errLib.New("Temp user not found", http.StatusNotFound)
	}

	return nil
}

func (r *InfoTempRepo) GetTempUserInfoByEmail(ctx context.Context, email string) (*entity.UserInfo, *errLib.CommonError) {

	info, err := r.Queries.GetTempUserInfoByEmail(ctx, sql.NullString{
		String: email,
		Valid:  true,
	})

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// Return a "not found" error if no rows are returned
			return nil, errLib.New("User info not found", http.StatusNotFound)
		}
		log.Printf("Error fetching temp user info: %v", err)
		return nil, errLib.New("Internal server error", http.StatusInternalServerError)
	}

	return &entity.UserInfo{
		ID:        info.ID,
		FirstName: info.FirstName,
		LastName:  info.LastName,
		Email:     info.Email.String,
	}, nil
}
