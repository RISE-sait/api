-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS users.athletes
(
    id              UUID PRIMARY KEY REFERENCES users.users (id),
    wins            INT         NOT NULL DEFAULT 0,                 -- Number of games won
    losses          INT         NOT NULL DEFAULT 0,                 -- Number of games lost
    points          INT         NOT NULL DEFAULT 0,                 -- Total points scored
    steals          INT         NOT NULL DEFAULT 0,                 -- Total steals
    assists         INT         NOT NULL DEFAULT 0,                 -- Total assists
    rebounds        INT         NOT NULL DEFAULT 0,                 -- Total rebounds
    created_at      TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP, -- Timestamp with time zone
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP  -- Track last update time
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE users.athletes;
-- +goose StatementEnd
