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
	"log"
	"net/http"

	"github.com/lib/pq"
)

// UsersRepository provides methods to interact with the user data in the database.
type UsersRepository struct {
	Queries *db.Queries
}

// NewUserRepository creates a new instance of UserRepository with the provided dependency injection container.
func NewUserRepository(db *db.Queries) *UsersRepository {
	return &UsersRepository{
		Queries: db,
	}
}

func (r *UsersRepository) CreateAthleteTx(ctx context.Context, tx *sql.Tx, input values.AthleteRegistrationRequestInfo) (values.UserReadInfo, error) {

	var response values.UserReadInfo

	queries := r.Queries

	if tx != nil {
		queries = queries.WithTx(tx)
	}

	args := db.CreateUserParams{
		HubspotID: sql.NullString{
			String: "",
			Valid:  false,
		},
		CountryAlpha2Code:        input.CountryCode,
		Age:                      input.Age,
		HasMarketingEmailConsent: input.HasConsentToEmailMarketing,
		HasSmsConsent:            input.HasConsentToSms,
		ParentID: uuid.NullUUID{
			UUID:  uuid.Nil,
			Valid: false,
		},
		FirstName: input.FirstName,
		LastName:  input.LastName,
	}

	user, err := queries.CreateUser(ctx, args)

	if err != nil {
		var pqErr *pq.Error

		if errors.As(err, &pqErr) {
			// Handle unique constraint violation (e.g., duplicate email)
			if pqErr.Code == databaseErrors.UniqueViolation { // Unique violation error code
				log.Printf("Unique constraint violation: %v", pqErr.Message)
				return response, errLib.New("Email or hubspot id already exists", http.StatusConflict)
			}
		}
		log.Printf("Unhandled error: %v", err)
		return response, errLib.New("Internal server error", http.StatusInternalServerError)
	}

	return values.UserReadInfo{
		Age:         user.Age,
		CountryCode: user.CountryAlpha2Code,
		FirstName:   user.FirstName,
		LastName:    user.LastName,
		Email:       user.E
		Role:        "",
		Phone:       nil,
	}, nil
}

func (r *UsersRepository) GetUserInfoByID(ctx context.Context, id uuid.UUID) (values.UserReadInfo, error) {

	user, err := r.Queries.GetUserByUserID(ctx, id)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return values.UserReadInfo{}, errLib.New("User not found", http.StatusNotFound)
		}

		log.Printf("Unhandled error: %v", err)
		return values.UserReadInfo{}, errLib.New("Internal server error", http.StatusInternalServerError)
	}

	return values.UserReadInfo{
		Age:                        user.Age,
		HasConsentToSms:            user.Hassmsconsent,
		HasConsentToEmailMarketing: user.Hasmarketingemailconsent,
		CountryCode:                user.Countryalpha2code,
		FirstName:,
		LastName:                   "",
		Email:                      nil,
		Role:                       "",
		Phone:                      nil,
	}, nil
}

func (r *UsersRepository) GetUserByHubspotID(ctx context.Context, id string) (uuid.UUID, *errLib.CommonError) {

	user, err := r.Queries.GetUserByHubSpotID(ctx, sql.NullString{String: id, Valid: true})

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return uuid.Nil, errLib.New("User not found", http.StatusNotFound)
		}

		log.Printf("Unhandled error: %v", err)
		return uuid.Nil, errLib.New("Internal server error", http.StatusInternalServerError)
	}

	return user.ID, nil
}

func (r *UsersRepository) CreateAthlete(ctx context.Context, tx *sql.Tx, id uuid.UUID) *errLib.CommonError {

	_, err := r.Queries.WithTx(tx).CreateAthleteInfo(ctx, id)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errLib.New("User not found", http.StatusNotFound)
		}

		log.Printf("Unhandled error: %v", err)
		return errLib.New("Internal server error", http.StatusInternalServerError)
	}

	return nil
}

func (r *UsersRepository) GetIsAthleteByID(ctx context.Context, id uuid.UUID) (bool, error) {

	_, err := r.Queries.GetAthleteInfoByUserID(ctx, id)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}

		log.Printf("Unhandled error: %v", err)
		return false, errLib.New("Internal server error", http.StatusInternalServerError)
	}

	return true, nil
}

func (r *UsersRepository) UpdateUserHubspotIdTx(ctx context.Context, tx *sql.Tx, userId uuid.UUID, hubspotId string) error {

	queries := r.Queries

	if tx != nil {
		queries = queries.WithTx(tx)
	}

	updatedRows, err := queries.UpdateUserHubspotId(ctx, db.UpdateUserHubspotIdParams{
		HubspotID: sql.NullString{String: hubspotId, Valid: true},
		ID:        userId,
	})

	if err != nil {
		var pqErr *pq.Error

		if errors.As(err, &pqErr) {
			// Handle unique constraint violation (e.g., duplicate email)
			if pqErr.Code == databaseErrors.UniqueViolation { // Unique violation error code
				log.Printf("Unique constraint violation: %v", pqErr.Message)
				return errLib.New("Hubspot id already exists", http.StatusConflict)
			}
		}
		log.Printf("Unhandled error: %v", err)
		return errLib.New("Internal server error", http.StatusInternalServerError)
	}

	if updatedRows == 0 {
		return errLib.New("User not found", http.StatusNotFound)
	}

	return nil
}
