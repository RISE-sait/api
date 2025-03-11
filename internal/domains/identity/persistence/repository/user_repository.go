package identity

import (
	databaseErrors "api/internal/constants"
	dbIdentity "api/internal/domains/identity/persistence/sqlc/generated"
	values "api/internal/domains/identity/values"
	errLib "api/internal/libs/errors"
	dbOutbox "api/internal/services/outbox/generated"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"log"
	"net/http"

	"github.com/lib/pq"
)

// UsersRepository provides methods to interact with the user data in the database.
type UsersRepository struct {
	IdentityQueries *dbIdentity.Queries
	OutboxQueries   *dbOutbox.Queries
}

// NewUserRepository creates a new instance of UserRepository with the provided dependency injection container.
func NewUserRepository(identityDb *dbIdentity.Queries, outboxDb *dbOutbox.Queries) *UsersRepository {
	return &UsersRepository{
		IdentityQueries: identityDb,
		OutboxQueries:   outboxDb,
	}
}

func (r *UsersRepository) createCustomerTx(ctx context.Context, tx *sql.Tx, input dbIdentity.CreateUserParams, role string) (values.UserReadInfo, *errLib.CommonError) {
	queries := r.IdentityQueries
	if tx != nil {
		queries = queries.WithTx(tx)
	}

	sqlStatement := fmt.Sprintf(
		"CREATE user (first_name, last_name, age, email, phone, role_name, is_active, country) VALUES ('%s', '%s', '%v', '%v', '%v', '%s', '%v', '%v')",
		input.FirstName, input.LastName, input.Age, input.Email, input.Phone,
		role, false, input.CountryAlpha2Code,
	)

	args := dbOutbox.InsertIntoOutboxParams{
		Status:       dbOutbox.AuditStatusPENDING,
		SqlStatement: sqlStatement,
	}

	rows, err := r.OutboxQueries.InsertIntoOutbox(ctx, args)

	if err != nil {
		log.Println(err.Error())
		return values.UserReadInfo{}, errLib.New("Failed to insert to outbox", http.StatusInternalServerError)
	}

	if rows == 0 {
		return values.UserReadInfo{}, errLib.New("Failed to insert to outbox", http.StatusInternalServerError)
	}

	user, err := queries.CreateUser(ctx, input)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == databaseErrors.UniqueViolation {
			log.Printf("Unique constraint violation: %v", pqErr.Message)
			return values.UserReadInfo{}, errLib.New("Email already exists", http.StatusConflict)
		}
		log.Printf("Unhandled error: %v", err)
		return values.UserReadInfo{}, errLib.New("Internal server error", http.StatusInternalServerError)
	}

	return values.UserReadInfo{
		ID:          user.ID,
		Age:         user.Age,
		CountryCode: user.CountryAlpha2Code,
		FirstName:   user.FirstName,
		LastName:    user.LastName,
		Email:       &user.Email.String,
		Role:        role,
		Phone:       &user.Phone.String,
	}, nil
}

func (r *UsersRepository) CreateAthleteTx(ctx context.Context, tx *sql.Tx, input values.AthleteRegistrationRequestInfo) (values.UserReadInfo, *errLib.CommonError) {
	customer, qErr := r.createCustomerTx(ctx, tx, dbIdentity.CreateUserParams{
		Email:                    sql.NullString{String: input.Email, Valid: true},
		HubspotID:                sql.NullString{},
		Phone:                    sql.NullString{String: input.Phone, Valid: true},
		CountryAlpha2Code:        input.CountryCode,
		Age:                      input.Age,
		HasMarketingEmailConsent: input.HasConsentToEmailMarketing,
		HasSmsConsent:            input.HasConsentToSms,
		ParentEmail:              sql.NullString{Valid: false},
		FirstName:                input.FirstName,
		LastName:                 input.LastName,
	}, "Athlete")

	if qErr != nil {
		log.Println(qErr.Error())
		return values.UserReadInfo{}, errLib.New("Failed to insert to athlete", http.StatusInternalServerError)
	}

	affectedRows, err := r.IdentityQueries.WithTx(tx).CreateAthlete(ctx, customer.ID)

	if err != nil {
		log.Println(err.Error())
	}

	if err != nil || affectedRows == 0 {
		return values.UserReadInfo{}, errLib.New("Failed to insert to athlete", http.StatusInternalServerError)
	}

	return values.UserReadInfo{
		ID:          customer.ID,
		Age:         customer.Age,
		CountryCode: customer.CountryCode,
		FirstName:   customer.FirstName,
		LastName:    customer.LastName,
		Email:       customer.Email,
		Role:        "Athlete",
		Phone:       customer.Phone,
	}, nil

}

func (r *UsersRepository) CreateParentTx(ctx context.Context, tx *sql.Tx, input values.ParentRegistrationRequestInfo) (values.UserReadInfo, *errLib.CommonError) {
	return r.createCustomerTx(ctx, tx, dbIdentity.CreateUserParams{
		Email:                    sql.NullString{String: input.Email, Valid: true},
		HubspotID:                sql.NullString{},
		CountryAlpha2Code:        input.CountryCode,
		Age:                      input.Age,
		HasMarketingEmailConsent: input.HasConsentToEmailMarketing,
		HasSmsConsent:            input.HasConsentToSms,
		ParentEmail:              sql.NullString{},
		FirstName:                input.FirstName,
		LastName:                 input.LastName,
	}, "Parent")
}

func (r *UsersRepository) CreateChildTx(ctx context.Context, tx *sql.Tx, input values.ChildRegistrationRequestInfo) (values.UserReadInfo, *errLib.CommonError) {
	createdCustomer, err := r.createCustomerTx(ctx, tx, dbIdentity.CreateUserParams{
		HubspotID:                sql.NullString{},
		CountryAlpha2Code:        input.CountryCode,
		Age:                      input.Age,
		HasMarketingEmailConsent: false,
		HasSmsConsent:            false,
		ParentEmail:              sql.NullString{String: input.ParentEmail, Valid: true},
		FirstName:                input.FirstName,
		LastName:                 input.LastName,
	}, "Child")

	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == databaseErrors.ForeignKeyViolation {
			log.Printf("Foreign key violation: %v", pqErr.Message)
			return values.UserReadInfo{}, errLib.New("Parent not found for the provided user id", http.StatusBadRequest)
		}
	}

	return createdCustomer, err
}

func (r *UsersRepository) GetIsActualParentChild(ctx context.Context, childID uuid.UUID, parentEmail string) (bool, *errLib.CommonError) {
	isConnected, err := r.IdentityQueries.GetIsActualParentChild(ctx, dbIdentity.GetIsActualParentChildParams{
		ParentEmail: sql.NullString{
			String: parentEmail,
			Valid:  true,
		},
		ChildID: childID,
	})

	if err != nil {
		log.Printf("Error verifying parent-child relationship: %v", err.Error())
		return false, errLib.New("Internal server error", http.StatusInternalServerError)
	}

	return isConnected, nil
}

func (r *UsersRepository) GetUserInfo(ctx context.Context, email string, id uuid.UUID) (values.UserReadInfo, *errLib.CommonError) {

	if email == "" && id == uuid.Nil {
		return values.UserReadInfo{}, errLib.New("Either use email or id to get user info. One must be present", http.StatusBadRequest)
	}

	if email != "" && id != uuid.Nil {
		return values.UserReadInfo{}, errLib.New("Either use email or id to get user info. Not both", http.StatusBadRequest)
	}

	var user dbIdentity.UsersUser
	var err error

	if email != "" {
		user, err = r.IdentityQueries.GetUserByEmail(ctx, sql.NullString{String: email, Valid: true})
	} else {
		user, err = r.IdentityQueries.GetUserByID(ctx, id)
	}

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return values.UserReadInfo{}, errLib.New("User not found", http.StatusNotFound)
		}
		log.Println(err.Error())
		return values.UserReadInfo{}, errLib.New("Internal server error", http.StatusInternalServerError)
	}

	response := values.UserReadInfo{
		ID:          user.ID,
		Age:         user.Age,
		CountryCode: user.CountryAlpha2Code,
		FirstName:   user.FirstName,
		LastName:    user.LastName,
		Email:       &email,
		Phone:       &user.Phone.String,
	}

	if user.ParentID.Valid {
		response.Role = "Child"
		return response, nil
	}

	if staffInfo, err := r.IdentityQueries.GetStaffById(ctx, user.ID); err != nil {

		if !errors.Is(err, sql.ErrNoRows) {

			// Got error, but it's not because the staff doesn't exist
			log.Println(err.Error())
			return values.UserReadInfo{}, errLib.New("Internal server error", http.StatusInternalServerError)
		}
	} else {

		// staff exists

		response.Role = staffInfo.RoleName
		return response, nil
	}

	if isParent, err := r.IdentityQueries.GetIsUserAParent(ctx, user.ID); err != nil {
		log.Println(err.Error())
		return values.UserReadInfo{}, errLib.New("Internal server error", http.StatusInternalServerError)
	} else if isParent {
		response.Role = "Parent"
		return response, nil
	}

	if isAthlete, err := r.IdentityQueries.GetIsAthleteByID(ctx, user.ID); err != nil {
		log.Println(err.Error())
		return values.UserReadInfo{}, errLib.New("Internal server error", http.StatusInternalServerError)
	} else if isAthlete {
		response.Role = "Athlete"
		return response, nil
	}

	log.Printf("error in getting user role with email %v", email)
	return values.UserReadInfo{}, errLib.New("Internal server error in getting user role with email", http.StatusInternalServerError)
}

func (r *UsersRepository) UpdateUserHubspotIdTx(ctx context.Context, tx *sql.Tx, userId uuid.UUID, hubspotId string) *errLib.CommonError {

	queries := r.IdentityQueries

	if tx != nil {
		queries = queries.WithTx(tx)
	}

	updatedRows, err := queries.UpdateUserHubspotId(ctx, dbIdentity.UpdateUserHubspotIdParams{
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
