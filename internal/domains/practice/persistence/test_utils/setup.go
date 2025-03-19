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

CREATE TABLE IF NOT EXISTS practices (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(50) NOT NULL UNIQUE, 
    description TEXT,
    level practice_level NOT NULL DEFAULT 'all',
    should_email_booking_notification BOOLEAN DEFAULT True,
    capacity INT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);`

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
