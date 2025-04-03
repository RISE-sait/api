-- +goose Up
-- +goose StatementBegin

CREATE SCHEMA IF NOT EXISTS events;

CREATE EXTENSION IF NOT EXISTS btree_gist; -- Ensure btree_gist extension is available

CREATE TABLE events.events
(
    id          UUID PRIMARY KEY     DEFAULT gen_random_uuid(),
    location_id UUID NOT NULL REFERENCES location.locations (id) ON DELETE RESTRICT,
    program_id          UUID,
    team_id             UUID,
    start_at    TIMESTAMPTZ NOT NULL,
    end_at      TIMESTAMPTZ NOT NULL,
    created_by  UUID        NOT NULL,
    updated_by  UUID        NOT NULL,
    capacity    int,
    is_cancelled        bool NOT NULL default false,
    cancellation_reason text,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT current_timestamp,
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT current_timestamp,
    CONSTRAINT fk_created_by FOREIGN KEY (created_by) REFERENCES users.users (id) ON DELETE CASCADE,
    CONSTRAINT fk_updated_by FOREIGN KEY (updated_by) REFERENCES users.users (id) ON DELETE CASCADE,
    CONSTRAINT fk_program FOREIGN KEY (program_id) REFERENCES program.programs (id) ON DELETE SET NULL,
    CONSTRAINT fk_team FOREIGN KEY (team_id) REFERENCES athletic.teams (id) ON DELETE SET NULL,

    CONSTRAINT no_overlapping_events
        EXCLUDE USING GIST (
        location_id WITH =,
        tstzrange(start_at, end_at) WITH &&
        ),

    CONSTRAINT check_start_end CHECK (start_at < end_at)
);

CREATE TABLE IF NOT EXISTS events.customer_enrollment
(
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    customer_id UUID NOT NUll,
    event_id UUID NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    checked_in_at TIMESTAMPTZ,
    is_cancelled bool NOT NULL DEFAULT false,
    CONSTRAINT fk_event FOREIGN KEY (event_id) REFERENCES events.events (id) ON DELETE CASCADE,
    CONSTRAINT fk_customer FOREIGN KEY (customer_id) REFERENCES users.users (id) ON DELETE CASCADE,
    CONSTRAINT unique_customer_event UNIQUE (customer_id, event_id)
);

CREATE TABLE events.attendance
(
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_id      UUID REFERENCES events.events (id) ON DELETE CASCADE NOT NULL,
    user_id       UUID REFERENCES users.users (id)                     NOT NULL,
    check_in_time TIMESTAMPTZ,
    CONSTRAINT unique_event_attendance UNIQUE (event_id, user_id)
);

CREATE TABLE IF NOT EXISTS events.staff
(
    event_id UUID NOT NULL,
    staff_id UUID NOT NULL,
    PRIMARY KEY (event_id, staff_id),
    CONSTRAINT fk_event FOREIGN KEY (event_id) REFERENCES events.events (id) ON DELETE CASCADE,
    CONSTRAINT fk_staff FOREIGN KEY (staff_id) REFERENCES staff.staff (id) ON DELETE CASCADE
);

CREATE INDEX idx_events_date_range ON events.events (start_at, end_at);
CREATE INDEX idx_events_location ON events.events (location_id);
CREATE INDEX idx_events_program ON events.events (program_id);
CREATE INDEX idx_events_team ON events.events (team_id);
CREATE INDEX idx_events_created_by ON events.events (created_by);
CREATE INDEX idx_attendance_user_history ON events.attendance (user_id, check_in_time DESC) WHERE check_in_time IS NOT NULL;
CREATE INDEX idx_customer_enrollment_event ON events.customer_enrollment (event_id);
CREATE INDEX idx_customer_enrollment_customer ON events.customer_enrollment (customer_id);
CREATE INDEX idx_staff_staff_id ON events.staff (staff_id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP schema if exists events cascade;
-- +goose StatementEnd
