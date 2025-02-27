-- +goose Up

-- +goose StatementBegin
CREATE TABLE events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
    begin_date_time TIMESTAMP WITH TIME ZONE NOT NULL,
    end_date_time TIMESTAMP WITH TIME ZONE NOT NULL,
    practice_id UUID,
    course_id UUID,
    game_id UUID,
    location_id UUID NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    CONSTRAINT fk_game FOREIGN KEY (game_id) REFERENCES games (id),
    CONSTRAINT fk_course FOREIGN KEY (course_id) REFERENCES courses (id),
    CONSTRAINT fk_practice FOREIGN KEY (practice_id) REFERENCES practices (id),
    CONSTRAINT fk_location FOREIGN KEY (location_id) REFERENCES locations (id),
    CONSTRAINT check_end_time CHECK (end_date_time > begin_date_time) -- Prevent invalid schedules
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
            (NEW.begin_date_time < e.end_date_time AND NEW.end_date_time > e.begin_date_time)
            )
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

-- +goose StatementEnd

-- +goose Down
DROP TRIGGER IF EXISTS trg_check_event_constraint ON events;
DROP FUNCTION IF EXISTS check_event_constraint;

-- Drop the events table
DROP TABLE IF EXISTS events;