-- +goose Up
-- +goose StatementBegin
CREATE SCHEMA IF NOT EXISTS users;

-- Create the 'users' table
CREATE TABLE users.users
(
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(), -- Auto-generate UUID for primary key
    hubspot_id TEXT UNIQUE,                                -- Unique identifier from HubSpot
    created_at TIMESTAMPTZ      DEFAULT CURRENT_TIMESTAMP, -- Timestamp with time zone
    updated_at TIMESTAMPTZ      DEFAULT CURRENT_TIMESTAMP  -- Track last update time
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE users.users;

DROP SCHEMA IF EXISTS users;
-- +goose StatementEnd