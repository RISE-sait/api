-- +goose Up
-- +goose StatementBegin

CREATE SCHEMA IF NOT EXISTS membership;

CREATE TABLE membership.memberships (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(50) NOT NULL,
    description TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS membership.memberships;

DROP SCHEMA IF EXISTS membership;
-- +goose StatementEnd
