package test_utils

import (
	db "api/internal/domains/practice/persistence/sqlc/generated"
	"database/sql"
	"github.com/stretchr/testify/require"
	"testing"
)

func SetupPracticeTestDbQueries(t *testing.T, testDb *sql.DB) (*db.Queries, func()) {

	migrationScript := `
DO $$ 
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'practice_level') THEN
        CREATE TYPE practice_level AS ENUM ('beginner', 'intermediate', 'advanced', 'all');
    END IF;
END $$;

create table public.practices
(
    id          uuid                     default gen_random_uuid()     not null
        primary key,
    name        varchar(150)                                           not null
        unique,
    description text                                                   not null,
    level       practice_level           default 'all'::practice_level not null,
    created_at  timestamp with time zone default CURRENT_TIMESTAMP     not null,
    updated_at  timestamp with time zone default CURRENT_TIMESTAMP     not null
);

alter table public.practices
    owner to postgres;

`

	_, err := testDb.Exec(migrationScript)
	require.NoError(t, err)

	// Return the repo and cleanup function
	repo := db.New(testDb)
	cleanUpScript := `DELETE FROM practices`

	// Cleanup function to delete data after test
	return repo, func() {
		_, err := testDb.Exec(cleanUpScript)
		require.NoError(t, err)
	}
}
