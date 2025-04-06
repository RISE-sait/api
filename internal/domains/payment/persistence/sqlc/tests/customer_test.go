package payments_test

import (
	identityDb "api/internal/domains/identity/persistence/sqlc/generated"
	teamDb "api/internal/domains/team/persistence/sqlc/generated"
	userDb "api/internal/domains/user/persistence/sqlc/generated"
	dbTestUtils "api/utils/test_utils"

	"context"
	"database/sql"
	"github.com/google/uuid"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetCustomerTeam(t *testing.T) {

	identityQ, _, _, _, _, _, paymentQ, teamQ, userQ, cleanup := dbTestUtils.SetupTestDbQueries(t, "../../../../../../db/migrations")

	defer cleanup()

	createdCustomer, err := identityQ.CreateUser(context.Background(), identityDb.CreateUserParams{
		CountryAlpha2Code: "CA",
		Age:               20,
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

	createdStaffRole, err := userQ.CreateStaffRole(context.Background(), "coach")

	require.NoError(t, err)

	createdPendingStaff, err := identityQ.CreateUser(context.Background(), identityDb.CreateUserParams{
		CountryAlpha2Code: "CA",
		Age:               20,
		Phone: sql.NullString{
			String: "+14141234567",
			Valid:  true,
		},
		HasMarketingEmailConsent: false,
		HasSmsConsent:            false,
		FirstName:                "John",
		LastName:                 "Doe",
	})

	require.NoError(t, err)

	createdStaff, err := identityQ.CreateApprovedStaff(context.Background(), identityDb.CreateApprovedStaffParams{
		ID:       createdPendingStaff.ID,
		RoleName: createdStaffRole.RoleName,
		IsActive: true,
	})

	require.NoError(t, err)

	createdTeam, err := teamQ.CreateTeam(context.Background(), teamDb.CreateTeamParams{
		Name:     "Go Team",
		Capacity: 20,
		CoachID: uuid.NullUUID{
			UUID:  createdStaff.ID,
			Valid: true,
		},
	})

	require.NoError(t, err)

	impactedRows, err := userQ.AddAthleteToTeam(context.Background(), userDb.AddAthleteToTeamParams{
		CustomerID: createdCustomer.ID,
		TeamID: uuid.NullUUID{
			UUID:  createdTeam.ID,
			Valid: true,
		},
	})

	require.NoError(t, err)
	require.Equal(t, int64(1), impactedRows)

	team, err := paymentQ.GetCustomersTeam(context.Background(), createdCustomer.ID)

	require.NoError(t, err)

	require.Equal(t, createdTeam.ID, team.ID)
}

func TestGetCustomerNonExistingTeam(t *testing.T) {

	identityQ, _, _, _, _, _, paymentQ, _, _, cleanup := dbTestUtils.SetupTestDbQueries(t, "../../../../../../db/migrations")

	defer cleanup()

	createdCustomer, err := identityQ.CreateUser(context.Background(), identityDb.CreateUserParams{
		CountryAlpha2Code: "CA",
		Age:               20,
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

	identityQ, _, _, _, _, _, paymentQ, _, _, cleanup := dbTestUtils.SetupTestDbQueries(t, "../../../../../../db/migrations")

	defer cleanup()

	createdCustomer, err := identityQ.CreateUser(context.Background(), identityDb.CreateUserParams{
		CountryAlpha2Code:        "CA",
		Age:                      20,
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

	_, _, _, _, _, _, paymentQ, _, _, cleanup := dbTestUtils.SetupTestDbQueries(t, "../../../../../../db/migrations")

	defer cleanup()

	isExist, err := paymentQ.IsCustomerExist(context.Background(), uuid.Nil)

	require.NoError(t, err)
	require.False(t, isExist)
}
