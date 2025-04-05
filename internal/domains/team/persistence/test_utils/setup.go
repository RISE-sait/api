package test_utils

import (
	db "api/internal/domains/program/persistence/sqlc/generated"
	"database/sql"
	"testing"

	"github.com/stretchr/testify/require"
)

func SetupTeamTestDbQueries(t *testing.T, testDb *sql.DB) (*db.Queries, func()) {

	migrationScript := `

create schema if not exists athletic;

create table if not exists athletic.teams
(
    id         uuid                     default gen_random_uuid() not null
        primary key,
    name       varchar(100)                                       not null,
    capacity   integer                                            not null,
    created_at timestamp with time zone default CURRENT_TIMESTAMP not null,
    updated_at timestamp with time zone default CURRENT_TIMESTAMP not null,
    coach_id   uuid
        constraint fk_coach
            references staff.staff
            on delete set null
);

alter table athletic.teams
    owner to postgres;

`

	_, err := testDb.Exec(migrationScript)
	require.NoError(t, err)

	// Return the repo and cleanup function
	repo := db.New(testDb)
	cleanUpScript := `DELETE FROM athletic.teams`

	// Cleanup function to delete data after test
	return repo, func() {
		_, err := testDb.Exec(cleanUpScript)
		require.NoError(t, err)
	}
}
