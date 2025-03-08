package test_utils

import (
	db "api/internal/domains/enrollment/persistence/sqlc/generated"
	"database/sql"
	"github.com/stretchr/testify/require"
	"testing"
)

func SetupEnrollmentTestDb(t *testing.T, testDb *sql.DB) (*db.Queries, func()) {

	migrationScript := `
	CREATE TABLE customer_enrollment (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    customer_id UUID NOT NUll,
    event_id UUID NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    checked_in_at TIMESTAMPTZ NULL,
    is_cancelled BOOLEAN DEFAULT FALSE,
    CONSTRAINT fk_event FOREIGN KEY (event_id) REFERENCES events(id) ON DELETE CASCADE,
    CONSTRAINT fk_customer FOREIGN KEY (customer_id) REFERENCES users.users (id) ON DELETE CASCADE,
    CONSTRAINT unique_customer_event UNIQUE (customer_id, event_id),
    CONSTRAINT chk_cancelled_checked_in CHECK (
        NOT (is_cancelled AND checked_in_at IS NOT NULL)
    )
);`

	_, err := testDb.Exec(migrationScript)
	require.NoError(t, err)

	// Return the repo and cleanup function
	repo := db.New(testDb)
	cleanUpScript := `DELETE FROM customer_enrollment`

	// Cleanup function to delete data after test
	return repo, func() {
		_, err := testDb.Exec(cleanUpScript)
		require.NoError(t, err)
	}
}
