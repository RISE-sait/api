package user

import (
	userTestUtils "api/internal/domains/identity/persistence/test_utils"
	"api/utils/test_utils"
	"context"
	"database/sql"
	"fmt"
	"testing"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"
)

func SetupUserRepo(t *testing.T) *Repository {
	testDb, _ := test_utils.SetupTestDB(t)

	queries, _ := userTestUtils.SetupUsersTestDb(t, testDb)

	return NewUserRepository(queries)
}

func TestCreateUserTx(t *testing.T) {
	repo := SetupUserRepo(t)

	userId, err := repo.CreateUserTx(context.Background(), nil)
	var errToCheck error

	if err != nil {
		errToCheck = fmt.Errorf("%v", err.Message)
	}

	require.NoError(t, errToCheck)
	require.NotNil(t, userId)
}

func TestGetUserIdByHubspotId(t *testing.T) {
	repo := SetupUserRepo(t)

	hubspotId := "test-hubspot-id"

	// Manually insert a user for testing
	_, queriesErr := repo.Queries.CreateUser(context.Background(), sql.NullString{
		String: hubspotId,
		Valid:  true,
	})
	require.NoError(t, queriesErr)

	// Retrieve user by HubSpot ID
	retrievedUserId, err := repo.GetUserIdByHubspotId(context.Background(), hubspotId)
	var errToCheck error

	if err != nil {
		errToCheck = fmt.Errorf("%v", err.Message)
	}

	require.NoError(t, errToCheck)
	require.NotNil(t, *retrievedUserId)
}

func TestUpdateUserHubspotIdTx(t *testing.T) {
	repo := SetupUserRepo(t)

	createdUserId, err := repo.CreateUserTx(context.Background(), nil)
	require.Nil(t, err)
	require.NotNil(t, createdUserId)

	// Step 2: Update the user with a new HubSpot ID
	newHubspotId := "new-hubspot-id"
	err = repo.UpdateUserHubspotIdTx(context.Background(), nil, *createdUserId, newHubspotId)

	var errToCheck error
	if err != nil {
		errToCheck = fmt.Errorf("%v", err.Message)
	}

	require.NoError(t, errToCheck)

	// Step 3: Verify the update
	retrievedUserId, err := repo.GetUserIdByHubspotId(context.Background(), newHubspotId)
	if err != nil {
		errToCheck = fmt.Errorf("%v", err.Message)
	}

	require.NoError(t, errToCheck)
	require.NotNil(t, retrievedUserId)
	require.Equal(t, *createdUserId, *retrievedUserId)
}
