-- +goose Up
-- +goose StatementBegin

CREATE OR REPLACE FUNCTION check_event_constraint() 
RETURNS TRIGGER 
AS $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM events e
        WHERE e.facility_id = NEW.facility_id
        AND e.day = NEW.day
        AND (
            (NEW.begin_time < e.begin_time AND NEW.end_time > e.end_time)
        )
    ) THEN
        RAISE EXCEPTION 'An event at this facility on the selected day overlaps with an existing event. Please choose a different time.';
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_check_event_constraint
BEFORE INSERT OR UPDATE ON events
FOR EACH ROW 
EXECUTE FUNCTION check_event_constraint();

-- +goose StatementEnd

-- +goose Down
-- SQL for rolling back the migration

DROP TRIGGER IF EXISTS trg_check_event_constraint ON events;

DROP FUNCTION IF EXISTS check_event_constraint;
