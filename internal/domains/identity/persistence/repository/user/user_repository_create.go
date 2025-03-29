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
	"github.com/lib/pq"
	"log"
	"net/http"
)

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
		return values.UserReadInfo{}, qErr
	}

	if err := r.IdentityQueries.WithTx(tx).CreateAthlete(ctx, customer.ID); err != nil {
		var pqErr *pq.Error
		if errors.As(qErr, &pqErr) && pqErr.Code == databaseErrors.UniqueViolation {
			return values.UserReadInfo{}, errLib.New("Athlete with that email already exists", http.StatusConflict)
		}
		log.Println(err.Error())
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
		return values.UserReadInfo{}, err
	}

	return createdCustomer, err
}
