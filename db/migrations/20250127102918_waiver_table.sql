-- +goose Up

-- +goose StatementBegin

CREATE SCHEMA IF NOT EXISTS waiver;

CREATE TABLE waiver.waiver (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    waiver_url TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
-- +goose StatementEnd

-- +goose Down

-- +goose StatementBegin
DROP TABLE IF EXISTS waiver.waiver;

DROP SCHEMA IF EXISTS waiver;
-- +goose StatementEnd
