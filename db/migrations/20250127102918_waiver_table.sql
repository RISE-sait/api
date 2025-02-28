-- +goose Up

-- +goose StatementBegin

CREATE SCHEMA IF NOT EXISTS waiver;

CREATE TABLE waiver.waiver
(
    id         UUID PRIMARY KEY     DEFAULT gen_random_uuid(),
    waiver_url TEXT        NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE waiver.waiver_signing
(
    user_id    UUID        NOT NULL,
    waiver_id  UUID        NOT NULL,
    is_signed  BOOLEAN     NOT NULL DEFAULT FALSE,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, waiver_id),
    FOREIGN KEY (user_id) REFERENCES users.users (id) ON DELETE CASCADE,
    FOREIGN KEY (waiver_id) REFERENCES waiver.waiver (id) ON DELETE CASCADE
);

CREATE TABLE waiver.pending_users_waiver_signing
(
    user_id    UUID        NOT NULL,
    waiver_id  UUID        NOT NULL,
    is_signed  BOOLEAN     NOT NULL DEFAULT FALSE,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, waiver_id),
    FOREIGN KEY (user_id) REFERENCES users.pending_users (id) ON DELETE CASCADE,
    FOREIGN KEY (waiver_id) REFERENCES waiver.waiver (id) ON DELETE CASCADE
);

-- +goose StatementEnd

-- +goose Down

-- +goose StatementBegin
DROP TABLE IF EXISTS waiver.pending_users_waiver_signing;

DROP TABLE IF EXISTS waiver.waiver_signing;

DROP TABLE IF EXISTS waiver.waiver;

DROP SCHEMA IF EXISTS waiver;
-- +goose StatementEnd
