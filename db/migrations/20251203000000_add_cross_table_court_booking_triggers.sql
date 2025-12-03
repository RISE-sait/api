-- +goose Up
-- +goose StatementBegin

-- Function to check court availability across all booking tables
CREATE OR REPLACE FUNCTION location.check_court_availability()
RETURNS TRIGGER AS $$
DECLARE
    v_court_id UUID;
    v_start_time TIMESTAMPTZ;
    v_end_time TIMESTAMPTZ;
    v_table_name TEXT;
    v_conflict_type TEXT;
BEGIN
    -- Get the table name that triggered this
    v_table_name := TG_TABLE_SCHEMA || '.' || TG_TABLE_NAME;

    -- Handle different column names across tables
    IF v_table_name = 'events.events' THEN
        v_court_id := NEW.court_id;
        v_start_time := NEW.start_at;
        v_end_time := NEW.end_at;
    ELSIF v_table_name = 'practice.practices' THEN
        v_court_id := NEW.court_id;
        v_start_time := NEW.start_time;
        -- Handle NULL end_time (default to 2 hours)
        v_end_time := COALESCE(NEW.end_time, NEW.start_time + interval '2 hours');
    ELSIF v_table_name = 'game.games' THEN
        v_court_id := NEW.court_id;
        v_start_time := NEW.start_time;
        -- Handle NULL end_time (default to 2 hours)
        v_end_time := COALESCE(NEW.end_time, NEW.start_time + interval '2 hours');
    END IF;

    -- Skip check if no court assigned
    IF v_court_id IS NULL THEN
        RETURN NEW;
    END IF;

    -- Check events table (skip if we're inserting into events)
    IF v_table_name != 'events.events' THEN
        IF EXISTS (
            SELECT 1 FROM events.events e
            WHERE e.court_id = v_court_id
              AND e.is_cancelled = false
              AND tstzrange(e.start_at, e.end_at, '[)') && tstzrange(v_start_time, v_end_time, '[)')
        ) THEN
            RAISE EXCEPTION 'Court is already booked by an event during this time';
        END IF;
    END IF;

    -- Check practices table (skip if we're inserting into practices)
    IF v_table_name != 'practice.practices' THEN
        IF EXISTS (
            SELECT 1 FROM practice.practices p
            WHERE p.court_id = v_court_id
              AND (p.status IS NULL OR p.status != 'canceled')
              AND tstzrange(p.start_time, COALESCE(p.end_time, p.start_time + interval '2 hours'), '[)') && tstzrange(v_start_time, v_end_time, '[)')
              AND (TG_OP = 'INSERT' OR p.id != NEW.id)
        ) THEN
            RAISE EXCEPTION 'Court is already booked by a practice during this time';
        END IF;
    END IF;

    -- Check games table (skip if we're inserting into games)
    IF v_table_name != 'game.games' THEN
        IF EXISTS (
            SELECT 1 FROM game.games g
            WHERE g.court_id = v_court_id
              AND (g.status IS NULL OR g.status != 'canceled')
              AND tstzrange(g.start_time, COALESCE(g.end_time, g.start_time + interval '2 hours'), '[)') && tstzrange(v_start_time, v_end_time, '[)')
              AND (TG_OP = 'INSERT' OR g.id != NEW.id)
        ) THEN
            RAISE EXCEPTION 'Court is already booked by a game during this time';
        END IF;
    END IF;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create triggers for each table
CREATE TRIGGER check_event_court_availability
    BEFORE INSERT OR UPDATE ON events.events
    FOR EACH ROW
    WHEN (NEW.court_id IS NOT NULL)
    EXECUTE FUNCTION location.check_court_availability();

CREATE TRIGGER check_practice_court_availability
    BEFORE INSERT OR UPDATE ON practice.practices
    FOR EACH ROW
    WHEN (NEW.court_id IS NOT NULL)
    EXECUTE FUNCTION location.check_court_availability();

CREATE TRIGGER check_game_court_availability
    BEFORE INSERT OR UPDATE ON game.games
    FOR EACH ROW
    WHEN (NEW.court_id IS NOT NULL)
    EXECUTE FUNCTION location.check_court_availability();

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TRIGGER IF EXISTS check_event_court_availability ON events.events;
DROP TRIGGER IF EXISTS check_practice_court_availability ON practice.practices;
DROP TRIGGER IF EXISTS check_game_court_availability ON game.games;
DROP FUNCTION IF EXISTS location.check_court_availability();

-- +goose StatementEnd
