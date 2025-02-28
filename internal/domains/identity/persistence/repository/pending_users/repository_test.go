package user_info_temp_repo

import (
	"api/internal/domains/identity/persistence/repository/user"
	userTestUtils "api/internal/domains/identity/persistence/test_utils"
	"api/utils/test_utils"
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func SetupUserInfoTempRepo(t *testing.T) (*user.Repository, *PendingUserInfoRepo) {
	testDb, _ := test_utils.SetupTestDB(t)

	queries, _ := userTestUtils.SetupUsersTestDb(t, testDb)
	queries, _ = userTestUtils.SetupTempUsersInfoTestDb(t, testDb)

	return user.NewUserRepository(queries), NewPendingUserInfoRepository(queries)
}

func TestCreateTempUserInfoTx(t *testing.T) {
	userRepo, tempUserInfoRepo := SetupUserInfoTempRepo(t)

	userId, err := userRepo.CreateUserTx(context.Background(), nil)

	var errToCheck error

	if err != nil {
		errToCheck = fmt.Errorf("%v", err.Message)
	}

	require.NoError(t, errToCheck)
	require.NotNil(t, userId)

	firstName := "John"
	lastName := "Doe"
	email := "johndoe@example.com"
	parentHubspotId := "parent-hubspot-id"
	age := 25

	err = tempUserInfoRepo.CreatePendingUserInfoTx(context.Background(), nil, *userId, firstName, lastName, &email, &parentHubspotId, age)

	if err != nil {
		errToCheck = fmt.Errorf("%v", err.Message)
	}

	require.NoError(t, errToCheck)

	// Verify creation by fetching the user info
	retrievedUser, err := tempUserInfoRepo.GetPendingUserInfoByEmail(context.Background(), email)

	if err != nil {
		errToCheck = fmt.Errorf("%v", err.Message)
	}

	require.NoError(t, errToCheck)

	require.NotNil(t, retrievedUser)
	require.Equal(t, firstName, retrievedUser.FirstName)
	require.Equal(t, lastName, retrievedUser.LastName)
	require.Equal(t, email, retrievedUser.Email)
}

func TestDeleteTempUserInfoTx(t *testing.T) {
	userRepo, tempUserInfoRepo := SetupUserInfoTempRepo(t)

	// Create user first in the user repository
	userId, err := userRepo.CreateUserTx(context.Background(), nil)

	var errToCheck error

	if err != nil {
		errToCheck = fmt.Errorf("%v", err.Message)
	}

	require.NoError(t, errToCheck)
	require.NotNil(t, userId)

	// Create temp user info
	firstName := "Jane"
	lastName := "Smith"
	email := "janesmith@example.com"
	age := 30

	err = tempUserInfoRepo.CreatePendingUserInfoTx(context.Background(), nil, *userId, firstName, lastName, &email, nil, age)

	if err != nil {
		errToCheck = fmt.Errorf("%v", err.Message)
	}

	require.NoError(t, errToCheck)

	// Delete temp user info
	err = tempUserInfoRepo.DeletePendingUserInfoTx(context.Background(), nil, *userId)

	if err != nil {
		errToCheck = fmt.Errorf("%v", err.Message)
	}

	require.NoError(t, errToCheck)

	// Verify deletion by attempting to retrieve user info
	_, err = tempUserInfoRepo.GetPendingUserInfoByEmail(context.Background(), email)

	require.NotNil(t, err)
	require.Equal(t, http.StatusNotFound, err.HTTPCode)
}

func TestGetTempUserInfoByEmail_NotFound(t *testing.T) {
	_, tempUserInfoRepo := SetupUserInfoTempRepo(t)

	_, err := tempUserInfoRepo.GetPendingUserInfoByEmail(context.Background(), "nonexistent@example.com")

	require.NotNil(t, err)
	require.Equal(t, http.StatusNotFound, err.HTTPCode)
}
