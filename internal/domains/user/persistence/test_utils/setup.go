package user

import (
	db "api/internal/domains/user/persistence/sqlc/generated"
	"database/sql"
	"github.com/stretchr/testify/require"
	"testing"
)

func SetupUsersTestDb(t *testing.T, testDb *sql.DB) (*db.Queries, func()) {

	migrationScript := `

create schema if not exists users;

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
		_, err = testDb.Exec(cleanUpScript)
		require.NoError(t, err)
	}
}

func SetupStaffsTestDb(t *testing.T, testDb *sql.DB) (*db.Queries, func()) {

	migrationScript := `

create schema if not exists staff;

create table if not exists staff.staff_roles
(
    id         uuid                     default gen_random_uuid() not null
        primary key,
    role_name  text                                               not null
        unique,
    created_at timestamp with time zone default CURRENT_TIMESTAMP not null,
    updated_at timestamp with time zone default CURRENT_TIMESTAMP not null
);

alter table staff.staff_roles
    owner to postgres;

create table if not exists staff.staff
(
    id         uuid                                               not null
        primary key
        references users.users,
    is_active  boolean                  default true              not null,
    created_at timestamp with time zone default CURRENT_TIMESTAMP not null,
    updated_at timestamp with time zone default CURRENT_TIMESTAMP not null,
    role_id    uuid                                               not null
        references staff.staff_roles
);

alter table staff.staff
    owner to postgres;
`

	_, err := testDb.Exec(migrationScript)
	require.NoError(t, err)

	// Return the repo and cleanup function
	repo := db.New(testDb)
	cleanUpScript := `
DELETE FROM staff.staff; 
DELETE  FROM staff.staff_roles;
`

	// Cleanup function to delete data after test
	return repo, func() {
		_, err = testDb.Exec(cleanUpScript)
		require.NoError(t, err)
	}
}
