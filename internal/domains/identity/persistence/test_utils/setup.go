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
create table if not exists users.users
(
    id                          uuid                     default gen_random_uuid() not null
        primary key,
    hubspot_id                  text
        unique,
    country_alpha2_code         char(2)                                            not null,
    gender                      char
        constraint users_gender_check
            check (gender = ANY (ARRAY ['M'::bpchar, 'F'::bpchar])),
    first_name                  varchar(20)                                        not null,
    last_name                   varchar(20)                                        not null,
    age                         integer                                            not null,
    parent_id                   uuid
        references users.users,
    phone                       varchar(25),
    email                       varchar(255)
        unique,
    has_marketing_email_consent boolean                                            not null,
    has_sms_consent             boolean                                            not null,
    created_at                  timestamp with time zone default CURRENT_TIMESTAMP not null,
    updated_at                  timestamp with time zone default CURRENT_TIMESTAMP not null
);

alter table users.users
    owner to postgres;


`

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
