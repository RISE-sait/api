package test_utils

import (
	db "api/internal/domains/location/persistence/sqlc/generated"
	"database/sql"
	"github.com/stretchr/testify/require"
	"testing"
)

func SetupFacilityTestDbQueries(t *testing.T, testDb *sql.DB) (*db.Queries, func()) {

	migrationScript := `

CREATE SCHEMA IF NOT EXISTS location;

CREATE TABLE location.facility_categories
(
    id   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL
);

CREATE TABLE location.facilities
(
    id                   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name                 VARCHAR(255) UNIQUE NOT NULL,
    address              VARCHAR(255)        NOT NULL,
    facility_category_id UUID                NOT NULL,
    FOREIGN KEY (facility_category_id) REFERENCES location.facility_categories (id) ON DELETE CASCADE
);

CREATE TABLE location.locations
(
    id                   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name                 VARCHAR(100) UNIQUE NOT NULL,
    facility_id UUID                NOT NULL,
    FOREIGN KEY (facility_id) REFERENCES location.facilities (id) ON DELETE CASCADE
);`

	_, err := testDb.Exec(migrationScript)
	require.NoError(t, err)

	// Return the repo and cleanup function
	repo := db.New(testDb)
	cleanUpScript := `DELETE FROM location.facilities`

	// Cleanup function to delete data after test
	return repo, func() {
		_, err := testDb.Exec(cleanUpScript)
		require.NoError(t, err)
	}
}
