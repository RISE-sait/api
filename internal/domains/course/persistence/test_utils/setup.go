package test_utils

import (
	db "api/internal/domains/course/persistence/sqlc/generated"
	"database/sql"
	"github.com/stretchr/testify/require"
	"testing"
)

func SetupCourseTestDb(t *testing.T, testDb *sql.DB) (*db.Queries, func()) {

	migrationScript := `

CREATE SCHEMA IF NOT EXISTS course;

	CREATE TABLE course.courses
(
    id          UUID PRIMARY KEY         DEFAULT gen_random_uuid(),
    name        VARCHAR(50) NOT NULL UNIQUE,
    description TEXT,
    capacity    INT         NOT NULL,
    created_at  TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at  TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);`

	_, err := testDb.Exec(migrationScript)
	require.NoError(t, err)

	// Return the repo and cleanup function
	repo := db.New(testDb)
	cleanUpScript := `DELETE FROM course.courses;`

	// Cleanup function to delete data after test
	return repo, func() {
		_, err := testDb.Exec(cleanUpScript)
		require.NoError(t, err)
	}
}
