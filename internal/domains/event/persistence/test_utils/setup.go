package test_utils

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
    event_start_at TIMESTAMP WITH TIME ZONE NOT NULL,
    event_end_at TIMESTAMP WITH TIME ZONE NOT NULL,
    session_start_time TIMETZ NOT NULL,
    session_end_time TIMETZ NOT NULL,
    day day_enum NOT NULL,
    practice_id UUID,
    course_id UUID,
    game_id UUID,
    location_id UUID NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT fk_game FOREIGN KEY (game_id) REFERENCES games (id),
    CONSTRAINT fk_course FOREIGN KEY (course_id) REFERENCES course.courses (id),
    CONSTRAINT fk_practice FOREIGN KEY (practice_id) REFERENCES practices (id),
    CONSTRAINT fk_location FOREIGN KEY (location_id) REFERENCES location.locations (id),
    CONSTRAINT event_end_after_start CHECK (event_end_at > event_start_at),
    CONSTRAINT session_end_after_start CHECK (session_end_time > events.session_start_time)
);

CREATE OR REPLACE FUNCTION check_event_constraint()
    RETURNS TRIGGER
AS $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM events e
        WHERE e.location_id = NEW.location_id
          AND (
            (NEW.event_start_at <= e.event_end_at AND NEW.event_end_at >= e.event_start_at)
            )
          AND (
            (NEW.session_start_time <= e.session_end_time AND NEW.session_end_time >= e.session_start_time)
            )
          AND e.day = NEW.day
    ) THEN
        RAISE EXCEPTION 'An event at this location overlaps with an existing event. Please choose a different time.';
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create the trigger to enforce the constraint
CREATE TRIGGER trg_check_event_constraint
    BEFORE INSERT OR UPDATE ON events
    FOR EACH ROW
EXECUTE FUNCTION check_event_constraint();
`

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
