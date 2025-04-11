package tests_test

import (
	dbTestUtils "api/utils/test_utils"
	"github.com/lib/pq"

	db "api/internal/domains/membership/persistence/sqlc/generated"

	"database/sql"

	"context"
	"github.com/google/uuid"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCreateMembership(t *testing.T) {

	dbConn, cleanup := dbTestUtils.SetupTestDbQueries(t, "../../../../../../db/migrations")

	queries := db.New(dbConn)

	defer cleanup()

	// Create a user to be the creator of the event
	params := db.CreateMembershipParams{
		Name:        "Go Basics Practice",
		Description: "Learn Go programming",
		Benefits:    "Free access to all events",
	}

	createdMembership, err := queries.CreateMembership(context.Background(), params)

	require.NoError(t, err)
	require.Equal(t, params.Name, createdMembership.Name)
	require.Equal(t, params.Description, createdMembership.Description)
	require.Equal(t, params.Benefits, createdMembership.Benefits)
	require.NotEmpty(t, createdMembership.ID)

}

func TestCreateExistingMembership(t *testing.T) {

	dbConn, cleanup := dbTestUtils.SetupTestDbQueries(t, "../../../../../../db/migrations")

	queries := db.New(dbConn)

	defer cleanup()

	// Create a user to be the creator of the event
	params := db.CreateMembershipParams{
		Name:        "Go Basics Practice",
		Description: "Learn Go programming",
		Benefits:    "Free access to all events",
	}

	createdMembership, err := queries.CreateMembership(context.Background(), params)

	require.NoError(t, err)
	require.Equal(t, params.Name, createdMembership.Name)
	require.Equal(t, params.Description, createdMembership.Description)
	require.Equal(t, params.Benefits, createdMembership.Benefits)
	require.NotEmpty(t, createdMembership.ID)

	// Attempt to create the same membership again
	_, err = queries.CreateMembership(context.Background(), params)

	require.Error(t, err)

	// Check if the error's constraint is unique_membership_table_name
	var pqErr *pq.Error

	require.ErrorAs(t, err, &pqErr)
	require.Equal(t, "unique_membership_table_name", pqErr.Constraint)
}

func TestUpdateMembership(t *testing.T) {

	dbConn, cleanup := dbTestUtils.SetupTestDbQueries(t, "../../../../../../db/migrations")

	queries := db.New(dbConn)

	defer cleanup()

	params := db.CreateMembershipParams{
		Name:        "Go Basics Practice",
		Description: "Learn Go programming",
		Benefits:    "Free access to all events",
	}

	createdMembership, err := queries.CreateMembership(context.Background(), params)

	require.NoError(t, err)

	updateMembershipParams := db.UpdateMembershipParams{
		ID:          createdMembership.ID,
		Name:        "Updated Membership Name",
		Description: "Updated description",
		Benefits:    "Updated benefits",
	}

	updatedMembership, err := queries.UpdateMembership(context.Background(), updateMembershipParams)
	require.NoError(t, err)

	require.Equal(t, updateMembershipParams.Name, updatedMembership.Name)
	require.Equal(t, updateMembershipParams.Description, updatedMembership.Description)
	require.Equal(t, updateMembershipParams.Benefits, updatedMembership.Benefits)

}

func TestUpdateNonExistingMembership(t *testing.T) {

	dbConn, cleanup := dbTestUtils.SetupTestDbQueries(t, "../../../../../../db/migrations")

	queries := db.New(dbConn)

	defer cleanup()

	updateMembershipParams := db.UpdateMembershipParams{
		ID:          uuid.New(),
		Name:        "Updated Membership Name",
		Description: "Updated description",
		Benefits:    "Updated benefits",
	}

	_, err := queries.UpdateMembership(context.Background(), updateMembershipParams)
	require.Error(t, err)
	require.Equal(t, sql.ErrNoRows, err)
}

func TestDeleteMembership(t *testing.T) {

	dbConn, cleanup := dbTestUtils.SetupTestDbQueries(t, "../../../../../../db/migrations")

	queries := db.New(dbConn)

	defer cleanup()

	params := db.CreateMembershipParams{
		Name:        "Go Basics Practice",
		Description: "Learn Go programming",
		Benefits:    "Free access to all events",
	}

	createdMembership, err := queries.CreateMembership(context.Background(), params)

	require.NoError(t, err)

	impactedRows, err := queries.DeleteMembership(context.Background(), createdMembership.ID)

	require.NoError(t, err)
	require.Equal(t, int64(1), impactedRows)

	_, err = queries.GetMembershipById(context.Background(), createdMembership.ID)

	require.Error(t, err)
	require.Equal(t, sql.ErrNoRows, err)
}
