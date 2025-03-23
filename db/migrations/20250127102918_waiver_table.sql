-- +goose Up

-- +goose StatementBegin

CREATE SCHEMA IF NOT EXISTS waiver;

CREATE TABLE IF NOT EXISTS waiver.waiver
(
    id          UUID PRIMARY KEY     DEFAULT gen_random_uuid(),
    waiver_url  TEXT        NOT NULL UNIQUE,
    waiver_name VARCHAR(30) NOT NULL UNIQUE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS waiver.waiver_signing
(
    user_id    UUID        NOT NULL,
    waiver_id  UUID        NOT NULL,
    is_signed  BOOLEAN     NOT NULL DEFAULT FALSE,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, waiver_id),
    FOREIGN KEY (user_id) REFERENCES users.users (id) ON DELETE CASCADE,
    FOREIGN KEY (waiver_id) REFERENCES waiver.waiver (id) ON DELETE CASCADE
);

-- +goose StatementEnd

-- +goose Down

-- +goose StatementBegin
DROP SCHEMA IF EXISTS waiver cascade;
-- +goose StatementEnd
