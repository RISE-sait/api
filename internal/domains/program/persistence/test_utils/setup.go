package test_utils

import (
	db "api/internal/domains/program/persistence/sqlc/generated"
	"database/sql"
	"testing"

	"github.com/stretchr/testify/require"
)

func SetupProgramTestDbQueries(t *testing.T, testDb *sql.DB) (*db.Queries, func()) {

	migrationScript := `

create schema if not exists program;

DO $$ 
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'program_level') THEN
        CREATE TYPE program.program_level AS ENUM ('beginner', 'intermediate', 'advanced', 'all');
    END IF;
END $$;

DO $$ 
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'program_type') THEN
create type program.program_type as enum ('practice', 'course', 'game', 'others');
    END IF;
END $$;

create table if not exists program.programs
(
    id          uuid                     default gen_random_uuid()            not null
        primary key,
    name        varchar(150)                                                  not null
        unique,
    description text                                                          not null,
    level       program.program_level    default 'all'::program.program_level not null,
    type        program.program_type                                          not null,
    created_at  timestamp with time zone default CURRENT_TIMESTAMP            not null,
    updated_at  timestamp with time zone default CURRENT_TIMESTAMP            not null,
    capacity    integer
);

alter table program.programs
    owner to postgres;

`

	_, err := testDb.Exec(migrationScript)
	require.NoError(t, err)

	// Return the repo and cleanup function
	repo := db.New(testDb)
	cleanUpScript := `DELETE FROM program.programs`

	// Cleanup function to delete data after test
	return repo, func() {
		_, err := testDb.Exec(cleanUpScript)
		require.NoError(t, err)
	}
}
