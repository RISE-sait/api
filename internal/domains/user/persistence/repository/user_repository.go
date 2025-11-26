package user

import (
	"api/internal/di"
	db "api/internal/domains/user/persistence/sqlc/generated"
	values "api/internal/domains/user/values"
	errLib "api/internal/libs/errors"
	"context"
	"database/sql"
	"errors"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"log"
	"net/http"
)

// UsersRepository provides methods to interact with the user data in the database.
type UsersRepository struct {
	Queries *db.Queries
}

// NewUsersRepository creates a new instance of UsersRepository with the provided dependency injection container.
func NewUsersRepository(container *di.Container) *UsersRepository {
	return &UsersRepository{
		Queries: container.Queries.UserDb,
	}
}

func (r *UsersRepository) UpdateUser(ctx context.Context, details values.UpdateValue) *errLib.CommonError {

	impactedRows, err := r.Queries.UpdateUserInfo(ctx, db.UpdateUserInfoParams{
		ParentID: uuid.NullUUID{
			UUID:  details.ParentID,
			Valid: details.ParentID != uuid.Nil,
		},
		FirstName: details.FirstName,
		LastName:  details.LastName,
		Email: sql.NullString{
			String: details.Email,
			Valid:  details.Email != "",
		},
		Phone: sql.NullString{
			String: details.Phone,
			Valid:  details.Phone != "",
		},
		Dob:                      details.Dob,
		CountryAlpha2Code:        details.CountryAlpha2Code,
		HasMarketingEmailConsent: details.HasMarketingEmailConsent,
		HasSmsConsent:            details.HasSmsConsent,
		Gender: sql.NullString{
			String: details.Gender,
			Valid:  details.Gender != "",
		},
		ID: details.ID,
		EmergencyContactName: sql.NullString{
			String: details.EmergencyContactName,
			Valid:  details.EmergencyContactName != "",
		},
		EmergencyContactPhone: sql.NullString{
			String: details.EmergencyContactPhone,
			Valid:  details.EmergencyContactPhone != "",
		},
		EmergencyContactRelationship: sql.NullString{
			String: details.EmergencyContactRelationship,
			Valid:  details.EmergencyContactRelationship != "",
		},
	})

	if err != nil {
		var pqErr *pq.Error

		if errors.As(err, &pqErr) {
			switch pqErr.Constraint {
			case "users_parent_id_fkey":
				return errLib.New("parent id does not exist", http.StatusNotFound)
			case "users_email_key":
				return errLib.New("email already exists", http.StatusConflict)
			case "users_hubspot_id_key":
				return errLib.New("hubspot id already exists", http.StatusConflict)
			}
		}
		log.Printf("Error updating user: %s", err)
		return errLib.New("internal error while updating user", http.StatusInternalServerError)
	}

	if impactedRows == 0 {
		return errLib.New("user not found", http.StatusNotFound)
	}

	return nil
}
