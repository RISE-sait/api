-- +goose Up

-- +goose StatementBegin
CREATE TYPE day_enum AS ENUM ('MONDAY', 'TUESDAY', 'WEDNESDAY', 'THURSDAY', 'FRIDAY', 'SATURDAY', 'SUNDAY');

CREATE TABLE events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
    begin_time TIMETZ NOT NULL,
    end_time TIMETZ NOT NULL,
    practice_id UUID NULL,
    course_id UUID,
    location_id UUID NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    day day_enum NOT NULL,
    CONSTRAINT fk_course FOREIGN KEY (course_id) REFERENCES courses (id),
    CONSTRAINT fk_practice FOREIGN KEY (practice_id) REFERENCES practices (id),
    CONSTRAINT fk_location FOREIGN KEY (location_id) REFERENCES locations (id),
    CONSTRAINT check_end_time CHECK (end_time > begin_time) -- Prevent invalid schedules
);

-- +goose StatementEnd

-- +goose Down
DROP TABLE IF EXISTS events;

DROP TYPE IF EXISTS day_enum;
DROP FUNCTION IF EXISTS update_timestamp;
