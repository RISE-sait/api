package location_test

import (
	databaseErrors "api/internal/constants"
	dbTestUtils "api/utils/test_utils"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/lib/pq"

	"github.com/stretchr/testify/require"

	db "api/internal/domains/location/persistence/sqlc/generated"
)

func TestCreateLocation(t *testing.T) {

	_, _, _, _, queries, cleanup := dbTestUtils.SetupTestDbQueries(t, "../../../../../db/migrations")
	defer cleanup()

	params := db.CreateLocationParams{
		Name:    "Test Location",
		Address: "123 Test St",
	}

	createdLocation, err := queries.CreateLocation(context.Background(), params)

	require.NoError(t, err)

	// Assert course data
	require.Equal(t, createdLocation.Name, params.Name)
	require.Equal(t, createdLocation.Address, params.Address)
}

func TestUpdateLocationValid(t *testing.T) {

	_, _, _, _, queries, cleanup := dbTestUtils.SetupTestDbQueries(t, "../../../../../db/migrations")
	defer cleanup()

	params := db.CreateLocationParams{
		Name:    "Test Location",
		Address: "123 Test St",
	}

	createdLocation, err := queries.CreateLocation(context.Background(), params)

	require.NoError(t, err)

	updateParams := db.UpdateLocationParams{
		ID:      createdLocation.ID,
		Name:    "Updated Location",
		Address: "456 Updated St",
	}

	impactedRows, err := queries.UpdateLocation(context.Background(), updateParams)

	require.NoError(t, err)

	require.Equal(t, impactedRows, int64(1))

	// Get the updated course and verify
	updatedLocation, err := queries.GetLocationById(context.Background(), createdLocation.ID)
	require.NoError(t, err)
	require.Equal(t, updateParams.Name, updatedLocation.Name)
	require.Equal(t, updateParams.Address, updatedLocation.Address)
}

func TestCreateLocationUniqueNameConstraint(t *testing.T) {

	_, _, _, _, queries, cleanup := dbTestUtils.SetupTestDbQueries(t, "../../../../../db/migrations")
	defer cleanup()

	params := db.CreateLocationParams{
		Name:    "Test Location",
		Address: "123 Test St",
	}

	createdLocation, err := queries.CreateLocation(context.Background(), params)
	require.NoError(t, err)
	require.Equal(t, createdLocation.Name, params.Name)
	require.Equal(t, createdLocation.Address, params.Address)

	// Attempt to create another course with the same name
	createdLocation2, err := queries.CreateLocation(context.Background(), params)
	require.Error(t, err)
	require.Equal(t, createdLocation2, db.LocationLocation{})

	var pgErr *pq.Error
	require.True(t, errors.As(err, &pgErr))
	require.Equal(t, databaseErrors.UniqueViolation, string(pgErr.Code))
}

func TestGetAllLocations(t *testing.T) {

	_, _, _, _, queries, cleanup := dbTestUtils.SetupTestDbQueries(t, "../../../../../db/migrations")
	defer cleanup()

	// Create some locations
	for i := 1; i <= 5; i++ {
		createLocationParams := db.CreateLocationParams{
			Name:    fmt.Sprintf("Course %d", i),
			Address: fmt.Sprintf("Address %d", i),
		}
		createdLocation, err := queries.CreateLocation(context.Background(), createLocationParams)
		require.NoError(t, err)
		require.Equal(t, createdLocation.Name, createLocationParams.Name)
		require.Equal(t, createdLocation.Address, createLocationParams.Address)
	}

	// Fetch all locations
	locations, err := queries.GetLocations(context.Background())
	require.NoError(t, err)
	require.EqualValues(t, 5, len(locations))
}

func TestUpdateNonExistentLocation(t *testing.T) {

	_, _, _, _, queries, cleanup := dbTestUtils.SetupTestDbQueries(t, "../../../../../db/migrations")

	defer cleanup()

	// Attempt to update a practice that doesn't exist
	nonExistentId := uuid.New() // Random UUID

	updateParams := db.UpdateLocationParams{
		ID:   nonExistentId,
		Name: "Updated Practice",
	}

	impactedRows, err := queries.UpdateLocation(context.Background(), updateParams)
	require.Nil(t, err)
	require.Equal(t, impactedRows, int64(0))
}

func TestDeleteLocation(t *testing.T) {

	_, _, _, _, queries, cleanup := dbTestUtils.SetupTestDbQueries(t, "../../../../../db/migrations")

	defer cleanup()

	// Create a course to delete
	params := db.CreateLocationParams{
		Name:    "Go Course",
		Address: "Learn Go programming",
	}

	createdLocation, err := queries.CreateLocation(context.Background(), params)
	require.NoError(t, err)

	require.Equal(t, createdLocation.Name, params.Name)
	require.Equal(t, createdLocation.Address, params.Address)

	location, err := queries.GetLocationById(context.Background(), createdLocation.ID)

	require.NoError(t, err)
	require.Equal(t, location.Name, params.Name)
	require.Equal(t, location.Address, params.Address)

	// Delete the course
	impactedRows, err := queries.DeleteLocation(context.Background(), location.ID)
	require.NoError(t, err)

	require.Equal(t, impactedRows, int64(1))

	// Attempt to fetch the deleted course (expecting error)
	_, err = queries.GetLocationById(context.Background(), location.ID)

	require.Error(t, err)

	require.Equal(t, sql.ErrNoRows, err)
}
