package test_utils

import (
	db "api/internal/domains/identity/persistence/sqlc/generated"
	"database/sql"
	"github.com/stretchr/testify/require"
	"testing"
)

func SetupUsersTestDb(t *testing.T, testDb *sql.DB) (*db.Queries, func()) {

	migrationScript := `
CREATE SCHEMA IF NOT EXISTS users;

-- Create the 'users' table
CREATE TABLE users.users
(
    id              UUID PRIMARY KEY     DEFAULT gen_random_uuid(), -- Auto-generate UUID for primary key
    hubspot_id      TEXT        NOT NULL UNIQUE,                    -- Unique identifier from HubSpot
    profile_pic_url TEXT,
    wins            INT         NOT NULL DEFAULT 0,                 -- Number of games won
    losses          INT         NOT NULL DEFAULT 0,                 -- Number of games lost
    points          INT         NOT NULL DEFAULT 0,                 -- Total points scored
    steals          INT         NOT NULL DEFAULT 0,                 -- Total steals
    assists         INT         NOT NULL DEFAULT 0,                 -- Total assists
    rebounds        INT         NOT NULL DEFAULT 0,                 -- Total rebounds
    created_at      TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP, -- Timestamp with time zone
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP  -- Track last update time
);`

	_, err := testDb.Exec(migrationScript)
	require.NoError(t, err)

	// Return the repo and cleanup function
	repo := db.New(testDb)
	cleanUpScript := `DELETE FROM users.users;`

	// Cleanup function to delete data after test
	return repo, func() {
		_, err := testDb.Exec(cleanUpScript)
		require.NoError(t, err)
	}
}

func SetupPendingUsersTestDb(t *testing.T, testDb *sql.DB) (*db.Queries, func()) {

	migrationScript := `
	CREATE TABLE users.pending_users
(
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    first_name      TEXT NOT NULL,
    last_name       TEXT NOT NULL,
    email           TEXT UNIQUE,
    parent_hubspot_id TEXT, -- Nullable, since not all users may have a parent
    age             INT NOT NULL,
    created_at      TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);`

	_, err := testDb.Exec(migrationScript)
	require.NoError(t, err)

	// Return the repo and cleanup function
	repo := db.New(testDb)
	cleanUpScript := `DELETE FROM users.pending_users`

	// Cleanup function to delete data after test
	return repo, func() {
		_, err := testDb.Exec(cleanUpScript)
		require.NoError(t, err)
	}
}
