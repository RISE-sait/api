package repository

import (
	database_errors "api/internal/constants"
	"api/internal/di"
	db "api/internal/domains/identity/persistence/sqlc/generated"
	"api/internal/domains/identity/values"
	errLib "api/internal/libs/errors"
	"context"
	"database/sql"
	"errors"
	"log"
	"net/http"

	"github.com/lib/pq"
)

type UserOptionalInfoRepository struct {
	Queries *db.Queries
}

func NewUserOptionalInfoRepository(container *di.Container) *UserOptionalInfoRepository {
	return &UserOptionalInfoRepository{
		Queries: container.Queries.IdentityDb,
	}
}

func (r *UserOptionalInfoRepository) GetUser(ctx context.Context, email, password string) *values.UserInfo {

	params := db.GetUserByEmailPasswordParams{
		Email: email,
		HashedPassword: sql.NullString{
			String: password,
			Valid:  true,
		},
	}

	user, err := r.Queries.GetUserByEmailPassword(ctx, params)

	if err != nil {
		log.Printf("Failed to validate user: %v", err)
		return nil
	}

	userInfo := &values.UserInfo{
		Email: email,
	}

	if user.FirstName.Valid {
		userInfo.FirstName = &user.FirstName.String
	}

	if user.LastName.Valid {
		userInfo.LastName = &user.LastName.String
	}

	if user.Phone.Valid {
		userInfo.Phone = &user.Phone.String
	}

	return userInfo
}

func (r *UserOptionalInfoRepository) CreateUserOptionalInfoTx(ctx context.Context, tx *sql.Tx, userInfo values.UserInfo, pwd *string) *errLib.CommonError {

	params := db.CreateUserOptionalInfoParams{
		Email: userInfo.Email,
	}

	if userInfo.FirstName != nil {
		params.FirstName = sql.NullString{
			String: *userInfo.FirstName,
			Valid:  *userInfo.FirstName != "",
		}
	}

	if userInfo.LastName != nil {
		params.LastName = sql.NullString{
			String: *userInfo.LastName,
			Valid:  *userInfo.LastName != "",
		}
	}

	if userInfo.Phone != nil {
		params.Phone = sql.NullString{
			String: *userInfo.Phone,
			Valid:  *userInfo.Phone != "",
		}
	}

	if pwd != nil {
		params.HashedPassword = sql.NullString{
			String: *pwd,
			Valid:  *pwd != "",
		}
	}

	rows, err := r.Queries.WithTx(tx).CreateUserOptionalInfo(ctx, params)

	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) {

			if pqErr.Code == database_errors.NotNullViolation {
				return errLib.New("Either user with the email is not found, or password is null", http.StatusBadRequest)
			}

			if pqErr.Code == database_errors.UniqueViolation {
				return errLib.New("Email already exists for the credentials", http.StatusBadRequest)
			}
		}
	}

	if rows == 0 {
		return errLib.New("Failed to create email password", 500)
	}

	return nil
}
