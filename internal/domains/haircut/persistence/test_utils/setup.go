package haircut

import (
	db "api/internal/domains/event/persistence/sqlc/generated"
	"database/sql"
	"github.com/stretchr/testify/require"
	"testing"
)

func SetupEventTestDbQueries(t *testing.T, testDb *sql.DB) (*db.Queries, func()) {

	migrationScript := `
	CREATE TYPE day_enum AS ENUM ('MONDAY', 'TUESDAY', 'WEDNESDAY', 'THURSDAY', 'FRIDAY', 'SATURDAY', 'SUNDAY');

CREATE TABLE events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
    begin_time TIME(0) NOT NULL,
    end_time TIME(0) NOT NULL,
    practice_id UUID NULL,
    facility_id UUID NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    day day_enum NOT NULL,
    CONSTRAINT fk_practice FOREIGN KEY (practice_id) REFERENCES practices (id),
    CONSTRAINT fk_facility FOREIGN KEY (facility_id) REFERENCES facilities (id),
    CONSTRAINT check_end_time CHECK (end_time > begin_time) -- Prevent invalid schedules
);`

	_, err := testDb.Exec(migrationScript)
	require.NoError(t, err)

	// Return the repo and cleanup function
	repo := db.New(testDb)
	cleanUpScript := `DELETE FROM events`

	// Cleanup function to delete data after test
	return repo, func() {
		_, err := testDb.Exec(cleanUpScript)
		require.NoError(t, err)
	}
}
