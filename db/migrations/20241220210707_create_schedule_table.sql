-- +goose Up
CREATE TYPE day_enum AS ENUM ('MONDAY', 'TUESDAY', 'WEDNESDAY', 'THURSDAY', 'FRIDAY', 'SATURDAY', 'SUNDAY');

CREATE TABLE schedules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
    begin_time TIME(0) NOT NULL,
    end_time TIME(0) NOT NULL,
    course_id UUID,
    facility_id UUID NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    day day_enum NOT NULL,
    CONSTRAINT fk_course FOREIGN KEY (course_id) REFERENCES courses (id),
    CONSTRAINT fk_facility FOREIGN KEY (facility_id) REFERENCES facilities (id)
);

-- +goose Down
DROP TABLE IF EXISTS schedules;

DROP TYPE IF EXISTS day_enum;