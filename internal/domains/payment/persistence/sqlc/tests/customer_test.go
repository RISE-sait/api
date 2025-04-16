package payments_test

import (
	identityDb "api/internal/domains/identity/persistence/sqlc/generated"
	paymentDb "api/internal/domains/payment/persistence/sqlc/generated"
	dbTestUtils "api/utils/test_utils"
	"time"

	"context"
	"database/sql"
	"testing"

	"github.com/google/uuid"

	"github.com/stretchr/testify/require"
)

func TestGetCustomerNonExistingTeam(t *testing.T) {

	db, cleanup := dbTestUtils.SetupTestDbQueries(t, "../../../../../../db/migrations")

	identityQ := identityDb.New(db)
	paymentQ := paymentDb.New(db)

	defer cleanup()

	createdCustomer, err := identityQ.CreateUser(context.Background(), identityDb.CreateUserParams{
		CountryAlpha2Code: "CA",
		Dob:               time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC),
		Phone: sql.NullString{
			String: "+1514123456337",
			Valid:  true,
		},
		HasMarketingEmailConsent: false,
		HasSmsConsent:            false,
		FirstName:                "John",
		LastName:                 "Doe",
	})

	require.NoError(t, err)

	err = identityQ.CreateAthlete(context.Background(), createdCustomer.ID)

	require.NoError(t, err)

	team, err := paymentQ.GetCustomersTeam(context.Background(), createdCustomer.ID)

	require.Nil(t, err)
	require.Equal(t, uuid.Nil, team.ID.UUID)
}

func TestIsCustomerExist(t *testing.T) {

	db, cleanup := dbTestUtils.SetupTestDbQueries(t, "../../../../../../db/migrations")

	identityQ := identityDb.New(db)
	paymentQ := paymentDb.New(db)

	defer cleanup()

	createdCustomer, err := identityQ.CreateUser(context.Background(), identityDb.CreateUserParams{
		CountryAlpha2Code:        "CA",
		Dob:                      time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC),
		HasMarketingEmailConsent: false,
		HasSmsConsent:            false,
		FirstName:                "John",
		LastName:                 "Doe",
	})

	require.NoError(t, err)

	err = identityQ.CreateAthlete(context.Background(), createdCustomer.ID)

	require.NoError(t, err)

	isExist, err := paymentQ.IsCustomerExist(context.Background(), createdCustomer.ID)

	require.NoError(t, err)
	require.True(t, isExist)
}

func TestIsCustomerExistFalse(t *testing.T) {

	db, cleanup := dbTestUtils.SetupTestDbQueries(t, "../../../../../../db/migrations")

	paymentQ := paymentDb.New(db)

	defer cleanup()

	isExist, err := paymentQ.IsCustomerExist(context.Background(), uuid.Nil)

	require.NoError(t, err)
	require.False(t, isExist)
}
