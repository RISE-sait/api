package test_utils

import (
	db "api/internal/domains/event/persistence/sqlc/generated"
	"database/sql"
	"github.com/stretchr/testify/require"
	"testing"
)

func SetupEventTestDbQueries(t *testing.T, testDb *sql.DB) (*db.Queries, func()) {

	migrationScript := `

CREATE SCHEMA IF NOT EXISTS events;

CREATE EXTENSION IF NOT EXISTS btree_gist;

	CREATE TYPE day_enum AS ENUM ('MONDAY', 'TUESDAY', 'WEDNESDAY', 'THURSDAY', 'FRIDAY', 'SATURDAY', 'SUNDAY');

create table if not exists events.events
(
    id                  uuid                     default gen_random_uuid() not null
        primary key,
    location_id         uuid                                               not null
        references location.locations
            on delete restrict,
    program_id          uuid
        constraint fk_program
            references program.programs
            on delete set null,
    team_id             uuid
        constraint fk_team
            references athletic.teams
            on delete set null,
    start_at            timestamp with time zone                           not null,
    end_at              timestamp with time zone                           not null,
    created_by          uuid                                               not null
        constraint fk_created_by
            references users.users
            on delete cascade,
    updated_by          uuid                                               not null
        constraint fk_updated_by
            references users.users
            on delete cascade,
    capacity            integer,
    is_cancelled        boolean                  default false             not null,
    cancellation_reason text,
    created_at          timestamp with time zone default CURRENT_TIMESTAMP not null,
    updated_at          timestamp with time zone default CURRENT_TIMESTAMP not null,
    constraint no_overlapping_events
        exclude using gist (location_id with =, tstzrange(start_at, end_at) with &&),
    constraint check_start_end
        check (start_at < end_at)
);

alter table events.events
    owner to postgres;

create table if not exists events.customer_enrollment
(
    id            uuid                     default gen_random_uuid() not null
        primary key,
    customer_id   uuid                                               not null
        constraint fk_customer
            references users.users
            on delete cascade,
    event_id      uuid                                               not null
        constraint fk_event
            references events.events
            on delete cascade,
    created_at    timestamp with time zone default CURRENT_TIMESTAMP not null,
    updated_at    timestamp with time zone default CURRENT_TIMESTAMP not null,
    checked_in_at timestamp with time zone,
    is_cancelled  boolean                  default false             not null,
    constraint unique_customer_event
        unique (customer_id, event_id)
);

alter table events.customer_enrollment
    owner to postgres;

create table if not exists events.staff
(
    event_id uuid not null
        constraint fk_event
            references events.events
            on delete cascade,
    staff_id uuid not null
        constraint fk_staff
            references staff.staff
            on delete cascade,
    primary key (event_id, staff_id)
);

alter table events.staff
    owner to postgres;
`

	_, err := testDb.Exec(migrationScript)
	require.NoError(t, err)

	// Return the repo and cleanup function
	repo := db.New(testDb)
	cleanUpScript := `DELETE FROM events.customer_enrollment cascade;
DELETE FROM events.staff cascade;
    DELETE FROM events.events cascade;
	`

	// Cleanup function to delete data after test
	return repo, func() {
		_, err = testDb.Exec(cleanUpScript)
		require.NoError(t, err)
	}
}
