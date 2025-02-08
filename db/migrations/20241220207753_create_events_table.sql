-- +goose Up

-- +goose StatementBegin
CREATE TYPE day_enum AS ENUM ('MONDAY', 'TUESDAY', 'WEDNESDAY', 'THURSDAY', 'FRIDAY', 'SATURDAY', 'SUNDAY');

CREATE TABLE events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
    begin_time TIME(0) NOT NULL,
    end_time TIME(0) NOT NULL,
    course_id UUID NULL,
    facility_id UUID NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    day day_enum NOT NULL,
    CONSTRAINT fk_course FOREIGN KEY (course_id) REFERENCES courses (id),
    CONSTRAINT fk_facility FOREIGN KEY (facility_id) REFERENCES facilities (id),
    CONSTRAINT check_end_time CHECK (end_time > begin_time) -- Prevent invalid schedules
);

CREATE FUNCTION update_timestamp() RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_events_timestamp
BEFORE UPDATE ON events
FOR EACH ROW EXECUTE FUNCTION update_timestamp();

-- +goose StatementEnd

-- +goose Down
DROP TABLE IF EXISTS events;

DROP TYPE IF EXISTS day_enum;
DROP FUNCTION IF EXISTS update_timestamp;
