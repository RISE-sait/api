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
CREATE SCHEMA IF NOT EXISTS athletic;
CREATE SCHEMA IF NOT EXISTS membership;
CREATE SCHEMA IF NOT EXISTS staff;

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

create table if not exists staff.staff_roles
(
    id        uuid default gen_random_uuid() not null
        primary key,
    role_name text                           not null
        unique
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

create table if not exists athletic.athletes
(
    id         uuid                                               not null
        primary key
        references users.users,
    wins       integer                  default 0                 not null,
    losses     integer                  default 0                 not null,
    points     integer                  default 0                 not null,
    steals     integer                  default 0                 not null,
    assists    integer                  default 0                 not null,
    rebounds   integer                  default 0                 not null,
    created_at timestamp with time zone default CURRENT_TIMESTAMP not null,
    updated_at timestamp with time zone default CURRENT_TIMESTAMP not null,
    team_id    uuid
        constraint fk_team
            references athletic.teams
            on delete set null
);

alter table athletic.athletes
    owner to postgres;

create table if not exists membership.memberships
(
    id          uuid                     default gen_random_uuid()     not null
        primary key,
    name        varchar(150)                                           not null,
    description text,
    created_at  timestamp with time zone default CURRENT_TIMESTAMP     not null,
    updated_at  timestamp with time zone default CURRENT_TIMESTAMP     not null,
    benefits    varchar(750)             default ''::character varying not null
);

alter table membership.memberships
    owner to postgres;

DO $$ 
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'payment_frequency') THEN
        CREATE TYPE public.payment_frequency AS ENUM ('once', 'week', 'month', 'day');
    END IF;
END $$;

alter type public.payment_frequency owner to postgres;

create table if not exists  membership.membership_plans
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

DO $$ 
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'membership_status') THEN
        create type public.membership_status as enum ('active', 'inactive', 'canceled', 'expired');
    END IF;
END $$;

alter type public.membership_status owner to postgres;

create table if not exists  users.customer_membership_plans
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

alter table users.customer_membership_plans
    owner to postgres;

`
	_, err := testDb.Exec(migrationScript)
	require.NoError(t, err)

	// Return the repo and cleanup function
	repo := db.New(testDb)
	cleanUpScript := `TRUNCATE table users.users, staff.staff_roles, athletic.teams, membership.membership_plans cascade;`

	// Cleanup function to delete data after test
	return repo, func() {
		_, err := testDb.Exec(cleanUpScript)
		require.NoError(t, err)
	}
}
