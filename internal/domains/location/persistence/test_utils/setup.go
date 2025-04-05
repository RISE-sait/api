package test_utils

import (
	db "api/internal/domains/location/persistence/sqlc/generated"
	"database/sql"
	"testing"

	"github.com/stretchr/testify/require"
)

func SetupLocationTestDbQueries(t *testing.T, testDb *sql.DB) (*db.Queries, func()) {

	migrationScript := `

create schema if not exists location;

create table if not exists location.locations
(
    id         uuid                     default gen_random_uuid() not null
        primary key,
    name       varchar(100)                                       not null
        unique,
    address    varchar(255)                                       not null,
    created_at timestamp with time zone default CURRENT_TIMESTAMP not null,
    updated_at timestamp with time zone default CURRENT_TIMESTAMP not null
);

alter table location.locations
    owner to postgres;

`

	_, err := testDb.Exec(migrationScript)
	require.NoError(t, err)

	// Return the repo and cleanup function
	repo := db.New(testDb)
	cleanUpScript := `DELETE FROM location.locations;`

	// Cleanup function to delete data after test
	return repo, func() {
		_, err := testDb.Exec(cleanUpScript)
		require.NoError(t, err)
	}
}
