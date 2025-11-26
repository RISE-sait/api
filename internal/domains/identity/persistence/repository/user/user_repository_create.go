package user

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"

	databaseErrors "api/internal/constants"
	dbIdentity "api/internal/domains/identity/persistence/sqlc/generated"
	values "api/internal/domains/identity/values"
	errLib "api/internal/libs/errors"
	dbOutbox "api/internal/services/outbox/generated"

	"github.com/lib/pq"
)

func (r *UsersRepository) createCustomerTx(ctx context.Context, tx *sql.Tx, input dbIdentity.CreateUserParams, role string) (values.UserReadInfo, *errLib.CommonError) {
	queries := r.IdentityQueries
	if tx != nil {
		queries = queries.WithTx(tx)
	}

	/*
		i added this to insert the "create user on hubspot" into outbox table, so that an admin can
		read from the outbox table and update user info on hubspot later on manually, and set the status to "DONE"

		This is necessary if we wanna guarantee that our db is consistent update with hubspot,
		since we cant guarantee that update in our db = update in hubspot
		since that would result in distributed transactions cuz it involves external service,
		and hubspot api is not transactional, which makes distributed transactions a nightmare.

		However, we use our db as the source of truth for user info
		to simplify ACID compliance by avoiding distributed transactions,

		so if updating user info on hubspot doesnt have to be guaranteed,
		then we can just create the user on hubspot in the code directly right after creating the user in our db.
		But the hubspot update is NOT GUARANTEED to be successful.

		U can definitely do retries to reduce manual work tho,
		such as incorporating a background job to retry failed hubspot updates,
		but imo it just makes things more complicated, and also again, its NOT GUARANTEED to be in sync

		i chose this approach ultimately because it is the simplest and most straightforward way to handle this,
		even though it does require some manual work to update user info on hubspot later on,
		but at least we can guarantee that our db is the source of truth for user info which simplifies ACID,
		and we dont have to worry bout outbox stuff and distributed transactions.

		After all, the ONLY way to guarantee that our db is in sync with hubspot without manual work
		is to use distributed transactions.

		U can kinda use Eventual Consistency to help keep things in sync but its still not guaranteed,
		but I really dont think its worth it especially considering the
		size and skillset of the team. I mean i could be wrong.

		Also ngl i literally forgot this was here until i came here
		to write this docs.
	*/
	sqlStatement := fmt.Sprintf(
		"CREATE user (first_name, last_name, dob, email, phone, role_name, is_active, country) VALUES ('%s', '%s', '%v', '%v', '%v', '%s', '%v', '%v')",
		input.FirstName, input.LastName, input.Dob, input.Email, input.Phone,
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
		DOB:         user.Dob,
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
		Email:                        sql.NullString{String: input.Email, Valid: true},
		HubspotID:                    sql.NullString{},
		Phone:                        sql.NullString{String: input.Phone, Valid: true},
		CountryAlpha2Code:            input.CountryCode,
		Dob:                          input.DOB,
		HasMarketingEmailConsent:     input.HasConsentToEmailMarketing,
		HasSmsConsent:                input.HasConsentToSms,
		ParentEmail:                  sql.NullString{Valid: false},
		FirstName:                    input.FirstName,
		LastName:                     input.LastName,
		EmergencyContactName:         sql.NullString{String: input.EmergencyContactName, Valid: input.EmergencyContactName != ""},
		EmergencyContactPhone:        sql.NullString{String: input.EmergencyContactPhone, Valid: input.EmergencyContactPhone != ""},
		EmergencyContactRelationship: sql.NullString{String: input.EmergencyContactRelationship, Valid: input.EmergencyContactRelationship != ""},
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
		DOB:         customer.DOB,
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
		Email:                        sql.NullString{String: input.Email, Valid: true},
		HubspotID:                    sql.NullString{},
		CountryAlpha2Code:            input.CountryCode,
		Dob:                          input.DOB,
		HasMarketingEmailConsent:     input.HasConsentToEmailMarketing,
		HasSmsConsent:                input.HasConsentToSms,
		ParentEmail:                  sql.NullString{},
		FirstName:                    input.FirstName,
		LastName:                     input.LastName,
		EmergencyContactName:         sql.NullString{},
		EmergencyContactPhone:        sql.NullString{},
		EmergencyContactRelationship: sql.NullString{},
	}, "Parent")
}

func (r *UsersRepository) CreateChildTx(ctx context.Context, tx *sql.Tx, input values.ChildRegistrationRequestInfo) (values.UserReadInfo, *errLib.CommonError) {
	createdCustomer, err := r.createCustomerTx(ctx, tx, dbIdentity.CreateUserParams{
		HubspotID:                    sql.NullString{},
		CountryAlpha2Code:            input.CountryCode,
		Dob:                          input.DOB,
		HasMarketingEmailConsent:     false,
		HasSmsConsent:                false,
		ParentEmail:                  sql.NullString{String: input.ParentEmail, Valid: true},
		FirstName:                    input.FirstName,
		LastName:                     input.LastName,
		EmergencyContactName:         sql.NullString{},
		EmergencyContactPhone:        sql.NullString{},
		EmergencyContactRelationship: sql.NullString{},
	}, "Child")
	if err != nil {
		return values.UserReadInfo{}, err
	}

	return createdCustomer, err
}
