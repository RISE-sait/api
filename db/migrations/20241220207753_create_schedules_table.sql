-- +goose Up

-- +goose StatementBegin
CREATE TYPE day_enum AS ENUM ('MONDAY', 'TUESDAY', 'WEDNESDAY', 'THURSDAY', 'FRIDAY', 'SATURDAY', 'SUNDAY');

DO
$$
    BEGIN
        IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'timetzrange') THEN
            CREATE TYPE timetzrange AS RANGE
            (
                subtype = timetz
            );
        END IF;
    END
$$;

CREATE EXTENSION IF NOT EXISTS btree_gist;

CREATE TABLE IF NOT EXISTS schedules
(
    id                  UUID PRIMARY KEY                  DEFAULT gen_random_uuid(),
    program_id          UUID,
    team_id             uuid,
    location_id         UUID                     NOT NULL,

    recurrence_start_at TIMESTAMP WITH TIME ZONE NOT NULL,
    recurrence_end_at   TIMESTAMP WITH TIME ZONE,

    -- Recurrence pattern
    day                 day_enum                 NOT NULL,
    event_start_time    TIMETZ                   NOT NULL,
    event_end_time      TIMETZ                   NOT NULL,

    created_at          TIMESTAMPTZ              NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ              NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_program FOREIGN KEY (program_id) REFERENCES program.programs (id) ON DELETE SET NULL,
    CONSTRAINT fk_location FOREIGN KEY (location_id) REFERENCES location.locations (id) ON DELETE cascade,
    CONSTRAINT fk_team FOREIGN KEY (team_id) REFERENCES athletic.teams (id) ON DELETE SET NULL,

    CONSTRAINT valid_recurrence_dates CHECK (
        recurrence_end_at IS NULL OR
        recurrence_end_at > recurrence_start_at
        ),

    CONSTRAINT check_event_end_after_start CHECK (event_start_time < event_end_time)
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS schedules CASCADE;
DO
$$
    BEGIN
        IF EXISTS (SELECT 1 FROM pg_type WHERE typname = 'timetzrange') THEN
            DROP TYPE timetzrange;
        END IF;
    END
$$;
DROP TYPE IF EXISTS day_enum;
-- +goose StatementEnd