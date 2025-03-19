package identity_tests

import (
	databaseErrors "api/internal/constants"
	identityTestUtils "api/internal/domains/identity/persistence/test_utils"
	"api/utils/test_utils"
	"context"
	"database/sql"
	"errors"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"testing"

	"github.com/stretchr/testify/require"

	db "api/internal/domains/identity/persistence/sqlc/generated"
)

func TestCreateValidUser(t *testing.T) {

	dbConn, _ := test_utils.SetupTestDB(t)

	queries, cleanup := identityTestUtils.SetupIdentityTestDb(t, dbConn)

	defer cleanup()

	firstName := "John"
	lastName := "Doe"
	age := int32(20)
	countryCode := "CA"
	email := sql.NullString{String: "john.doe@example.com", Valid: true}
	phone := sql.NullString{String: "+15141234567", Valid: true}

	// Create user parameters
	createUserParams := db.CreateUserParams{
		CountryAlpha2Code:        countryCode,
		Email:                    email,
		Age:                      age,
		Phone:                    phone,
		HasMarketingEmailConsent: false,
		HasSmsConsent:            false,
		FirstName:                firstName,
		LastName:                 lastName,
	}

	createdUser, err := queries.CreateUser(context.Background(), createUserParams)

	require.NoError(t, err)

	require.Equal(t, firstName, createdUser.FirstName)
	require.Equal(t, lastName, createdUser.LastName)
	require.Equal(t, age, createdUser.Age)
	require.Equal(t, countryCode, createdUser.CountryAlpha2Code)
	require.Equal(t, email, createdUser.Email)
	require.Equal(t, phone, createdUser.Phone)
	require.Equal(t, false, createdUser.HasMarketingEmailConsent)
	require.Equal(t, false, createdUser.HasSmsConsent)
}

func TestCreateUserViolateUniqueEmailConstraint(t *testing.T) {
	dbConn, _ := test_utils.SetupTestDB(t)
	queries, cleanup := identityTestUtils.SetupIdentityTestDb(t, dbConn)
	defer cleanup()

	// Define test data
	firstName := "John"
	lastName := "Doe"
	age := int32(20)
	countryCode := "CA"
	email := sql.NullString{String: "john.doe@example.com", Valid: true}
	phone := sql.NullString{String: "+15141234567", Valid: true}

	// Create the first user
	createUserParams := db.CreateUserParams{
		CountryAlpha2Code:        countryCode,
		Email:                    email,
		Age:                      age,
		Phone:                    phone,
		HasMarketingEmailConsent: false,
		HasSmsConsent:            false,
		FirstName:                firstName,
		LastName:                 lastName,
	}

	_, err := queries.CreateUser(context.Background(), createUserParams)
	require.NoError(t, err)

	// Attempt to create another user with the same email
	_, err = queries.CreateUser(context.Background(), createUserParams)
	require.Error(t, err)

	// Check for unique constraint violation
	var pgErr *pq.Error
	require.True(t, errors.As(err, &pgErr))
	require.Equal(t, databaseErrors.UniqueViolation, string(pgErr.Code)) // 23505 is the error code for unique violation
}

func TestUpdateUser(t *testing.T) {
	dbConn, _ := test_utils.SetupTestDB(t)
	queries, cleanup := identityTestUtils.SetupIdentityTestDb(t, dbConn)
	defer cleanup()

	// Define test data for creating a user
	firstName := "John"
	lastName := "Doe"
	age := int32(20)
	countryCode := "CA"
	email := sql.NullString{String: "john.doe@example.com", Valid: true}
	phone := sql.NullString{String: "+15141234567", Valid: true}

	// Create user parameters
	createUserParams := db.CreateUserParams{
		CountryAlpha2Code:        countryCode,
		HubspotID:                sql.NullString{String: "abcde", Valid: true},
		Email:                    email,
		Age:                      age,
		Phone:                    phone,
		HasMarketingEmailConsent: false,
		HasSmsConsent:            false,
		FirstName:                firstName,
		LastName:                 lastName,
	}

	// Create the user
	createdUser, err := queries.CreateUser(context.Background(), createUserParams)
	require.NoError(t, err)

	// Define updated data
	newFirstName := "John"
	newLastName := "Doe"
	newAge := int32(20)
	newCountryCode := "CA"
	newEmail := sql.NullString{String: "john.doe@example.com", Valid: true}
	newPhone := sql.NullString{String: "+15141234567", Valid: true}

	// Update the user
	updateUserParams := db.UpdateUserHubspotIdParams{
		ID:        createdUser.ID,
		HubspotID: sql.NullString{String: "abcd", Valid: true},
	}

	_, err = queries.UpdateUserHubspotId(context.Background(), updateUserParams)
	require.NoError(t, err)

	// Fetch the updated user
	updatedUser, err := queries.GetUserByIdOrEmail(context.Background(), db.GetUserByIdOrEmailParams{
		ID: uuid.NullUUID{
			UUID:  createdUser.ID,
			Valid: true,
		},
		Email: sql.NullString{Valid: false},
	})
	require.NoError(t, err)

	// Assert the updated fields
	require.Equal(t, newFirstName, updatedUser.FirstName)
	require.Equal(t, newLastName, updatedUser.LastName)
	require.Equal(t, newAge, updatedUser.Age)
	require.Equal(t, newCountryCode, updatedUser.CountryAlpha2Code)
	require.Equal(t, newEmail, updatedUser.Email)
	require.Equal(t, newPhone, updatedUser.Phone)
	require.Equal(t, false, updatedUser.HasMarketingEmailConsent)
	require.Equal(t, false, updatedUser.HasSmsConsent)
}

func TestGetNonExistingUser(t *testing.T) {

	dbConn, _ := test_utils.SetupTestDB(t)

	queries, cleanup := identityTestUtils.SetupIdentityTestDb(t, dbConn)

	defer cleanup()

	_, err := queries.GetUserByIdOrEmail(context.Background(), db.GetUserByIdOrEmailParams{
		ID: uuid.NullUUID{
			UUID:  uuid.Nil,
			Valid: true,
		},
		Email: sql.NullString{Valid: false},
	})

	require.Equal(t, sql.ErrNoRows, err)
}

func TestUpdateNonExistentUser(t *testing.T) {

	dbConn, _ := test_utils.SetupTestDB(t)

	queries, cleanup := identityTestUtils.SetupIdentityTestDb(t, dbConn)

	defer cleanup()

	// Attempt to update a course that doesn't exist
	nonExistentId := uuid.New() // Random UUID

	updateParams := db.UpdateUserHubspotIdParams{
		ID:        nonExistentId,
		HubspotID: sql.NullString{String: "abcde", Valid: true},
	}

	rows, err := queries.UpdateUserHubspotId(context.Background(), updateParams)

	require.Equal(t, int64(0), rows)

	require.Nil(t, err)
}
