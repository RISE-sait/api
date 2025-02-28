-- +goose Up
-- +goose StatementBegin
CREATE SCHEMA IF NOT EXISTS users;

-- Create the 'users' table
CREATE TABLE users.users
(
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(), -- Auto-generate UUID for primary key
    hubspot_id TEXT UNIQUE,                                -- Unique identifier from HubSpot
    profile_pic_url TEXT,
    wins            INT  NOT NULL DEFAULT 0, -- Number of games won
    losses          INT  NOT NULL DEFAULT 0, -- Number of games lost
    points          INT  NOT NULL DEFAULT 0, -- Total points scored
    steals          INT  NOT NULL DEFAULT 0, -- Total steals
    assists         INT  NOT NULL DEFAULT 0, -- Total assists
    rebounds        INT  NOT NULL DEFAULT 0, -- Total rebounds
    created_at TIMESTAMPTZ  NOT NULL    DEFAULT CURRENT_TIMESTAMP, -- Timestamp with time zone
    updated_at TIMESTAMPTZ   NOT NULL   DEFAULT CURRENT_TIMESTAMP  -- Track last update time
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE users.users;

DROP SCHEMA IF EXISTS users;
-- +goose StatementEnd