package test_utils

import (
	db "api/internal/domains/payment/persistence/sqlc/generated"
	"database/sql"
	"github.com/stretchr/testify/require"
	"testing"
)

func SetupPostPaymentTestDbQueries(t *testing.T, testDb *sql.DB) (*db.Queries, func()) {

	migrationScript := `

create schema if not exists program;

create table if not exists program.customer_enrollment
(
    id           uuid                     default gen_random_uuid() not null
        primary key,
    customer_id  uuid                                               not null
        constraint fk_customer
            references users.users
            on delete cascade,
    program_id   uuid                                               not null
        constraint fk_program
            references program.programs
            on delete cascade,
    created_at   timestamp with time zone default CURRENT_TIMESTAMP not null,
    updated_at   timestamp with time zone default CURRENT_TIMESTAMP not null,
    is_cancelled boolean                  default false             not null,
    constraint unique_customer_program
        unique (customer_id, program_id)
);

alter table program.customer_enrollment
    owner to postgres;



CREATE SCHEMA IF NOT EXISTS events;

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
`

	_, err := testDb.Exec(migrationScript)
	require.NoError(t, err)

	// Return the repo and cleanup function
	repo := db.New(testDb)
	cleanUpScript := `DELETE FROM events.customer_enrollment cascade;
	`

	// Cleanup function to delete data after test
	return repo, func() {
		_, err = testDb.Exec(cleanUpScript)
		require.NoError(t, err)
	}
}
