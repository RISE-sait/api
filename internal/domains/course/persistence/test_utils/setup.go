package test_utils

import (
	db "api/internal/domains/course/persistence/sqlc/generated"
	"database/sql"
	"github.com/stretchr/testify/require"
	"testing"
)

func SetupCourseTestDb(t *testing.T, testDb *sql.DB) (*db.Queries, func()) {

	migrationScript := `

create table if not exists public.courses
(
    id          uuid                     default gen_random_uuid() not null
        primary key,
    name        varchar(50)                                        not null
        unique,
    description text                                               not null,
    capacity    integer                                            not null,
    created_at  timestamp with time zone default CURRENT_TIMESTAMP not null,
    updated_at  timestamp with time zone default CURRENT_TIMESTAMP not null
);

alter table public.courses
    owner to postgres;

`

	_, err := testDb.Exec(migrationScript)
	require.NoError(t, err)

	// Return the repo and cleanup function
	repo := db.New(testDb)
	cleanUpScript := `DELETE FROM courses;`

	// Cleanup function to delete data after test
	return repo, func() {
		_, err = testDb.Exec(cleanUpScript)
		require.NoError(t, err)
	}
}
