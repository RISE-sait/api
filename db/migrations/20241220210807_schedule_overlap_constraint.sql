-- +goose Up
-- +goose StatementBegin

CREATE OR REPLACE FUNCTION check_schedule_constraint() 
RETURNS TRIGGER 
AS $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM schedules s
        WHERE s.facility_id = NEW.facility_id
        AND s.day = NEW.day
        AND (
            (NEW.begin_datetime < s.end_datetime AND NEW.end_datetime > s.begin_datetime)
        )

    ) THEN
        RAISE EXCEPTION 'Overlapping schedules for the same facility on the same day are not allowed';
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_check_schedule_constraint
BEFORE INSERT OR UPDATE ON schedules
FOR EACH ROW 
EXECUTE FUNCTION check_schedule_constraint();

-- +goose StatementEnd
-- +goose Down
-- SQL for rolling back the migration

DROP TRIGGER IF EXISTS trg_check_schedule_constraint ON schedules;

DROP FUNCTION IF EXISTS check_schedule_constraint;