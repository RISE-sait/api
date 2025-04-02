-- +goose Up
-- +goose StatementBegin
DROP TABLE IF EXISTS events.attendance;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
CREATE TABLE if not exists events.attendance
(
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_id      UUID REFERENCES events.events (id) ON DELETE CASCADE NOT NULL,
    user_id       UUID REFERENCES users.users (id)                     NOT NULL,
    check_in_time TIMESTAMPTZ,
    CONSTRAINT unique_event_attendance UNIQUE (event_id, user_id)
);

CREATE INDEX if not exists idx_attendance_user_history ON events.attendance (user_id, check_in_time DESC) WHERE check_in_time IS NOT NULL;

-- +goose StatementEnd
