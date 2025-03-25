-- +goose Up
-- +goose StatementBegin
CREATE TABLE athletic.coach_stats
(
    id         UUID PRIMARY KEY     DEFAULT gen_random_uuid(),
    wins       INT         NOT NULL DEFAULT 0, -- Number of games won
    losses     INT         NOT NULL DEFAULT 0, -- Number of games lost
    coach_id   uuid        NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_coach FOREIGN KEY (coach_id) REFERENCES staff.staff (id)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE athletic.coach_stats;
-- +goose StatementEnd
