package test_utils

import (
	db "api/internal/domains/identity/persistence/sqlc/generated"
	"database/sql"
	"github.com/stretchr/testify/require"
	"testing"
)

func SetupIdentityTestDb(t *testing.T, testDb *sql.DB) (*db.Queries, func()) {

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

CREATE SCHEMA IF NOT EXISTS membership;

create table membership.memberships
(
    id          uuid                     default gen_random_uuid() not null
        primary key,
    name        varchar(150)                                       not null,
    description text,
    created_at  timestamp with time zone default CURRENT_TIMESTAMP not null,
    updated_at  timestamp with time zone default CURRENT_TIMESTAMP not null
);

alter table membership.memberships
    owner to postgres;

create table membership.membership_plans
(
    id                uuid                     default gen_random_uuid() not null
        primary key,
    name              varchar(150)                                       not null,
    price             numeric(6, 2)                                      not null,
    joining_fee       numeric(6, 2)                                      not null,
    auto_renew        boolean                  default false             not null,
    membership_id     uuid                                               not null
        constraint fk_membership
            references membership.memberships,
    payment_frequency payment_frequency                                  not null,
    amt_periods       integer,
    created_at        timestamp with time zone default CURRENT_TIMESTAMP not null,
    updated_at        timestamp with time zone default CURRENT_TIMESTAMP not null,
    constraint unique_membership_name
        unique (membership_id, name)
);

alter table membership.membership_plans
    owner to postgres;

create table public.customer_membership_plans
(
    id                 uuid                     default gen_random_uuid()           not null
        primary key,
    customer_id        uuid                                                         not null
        constraint fk_customer
            references users.users,
    membership_plan_id uuid
        constraint fk_membership_plan
            references membership.membership_plans,
    start_date         timestamp with time zone default CURRENT_TIMESTAMP           not null,
    renewal_date       timestamp with time zone,
    status             membership_status        default 'active'::membership_status not null,
    created_at         timestamp with time zone default CURRENT_TIMESTAMP           not null,
    updated_at         timestamp with time zone default CURRENT_TIMESTAMP           not null
);

alter table public.customer_membership_plans
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
