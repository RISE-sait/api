-- +goose Up
CREATE TABLE waiver (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    waiver_url TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- +goose Down
DROP TABLE IF EXISTS waiver;