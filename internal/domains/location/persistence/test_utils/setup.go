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

create table location.locations
(
    id      uuid default gen_random_uuid() not null
        primary key,
    name    varchar(100)                   not null
        unique,
    address varchar(255)                   not null
);

alter table location.locations
    owner to postgres;

`

	_, err := testDb.Exec(migrationScript)
	require.NoError(t, err)

	// Return the repo and cleanup function
	repo := db.New(testDb)
	cleanUpScript := `DELETE FROM location.locations`

	// Cleanup function to delete data after test
	return repo, func() {
		_, err := testDb.Exec(cleanUpScript)
		require.NoError(t, err)
	}
}
