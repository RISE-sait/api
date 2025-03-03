package user

import (
	db "api/internal/domains/user/persistence/sqlc/generated"
	"database/sql"
	"github.com/stretchr/testify/require"
	"testing"
)

func SetupUsersTestDb(t *testing.T, testDb *sql.DB) (*db.Queries, func()) {

	migrationScript := `
	CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    hubspot_id text unique,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);`

	_, err := testDb.Exec(migrationScript)
	require.NoError(t, err)

	// Return the repo and cleanup function
	repo := db.New(testDb)
	cleanUpScript := `DELETE FROM users`

	// Cleanup function to delete data after test
	return repo, func() {
		_, err := testDb.Exec(cleanUpScript)
		require.NoError(t, err)
	}
}
