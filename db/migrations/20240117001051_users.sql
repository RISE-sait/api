-- +goose Up
-- +goose StatementBegin
CREATE SCHEMA IF NOT EXISTS users;

-- Create the 'users' table
CREATE TABLE users.users
(
    id                          UUID PRIMARY KEY     DEFAULT gen_random_uuid(), -- Auto-generate UUID for primary key
    hubspot_id                  TEXT UNIQUE,                                    -- Unique identifier from HubSpot
    country_alpha2_code         char(2)     NOT NUll,
    gender CHAR(1) CHECK (gender IN ('M', 'F')) NULL,
    first_name                  varchar(20) NOT NULL,
    last_name                   varchar(20) NOT NULL,
    age                         int         NOT NULL,
    parent_id                   UUID REFERENCES users.users (id),
    phone  varchar(25),
    email  varchar(255) UNIQUE,
    has_marketing_email_consent bool        NOT NULL,
    has_sms_consent             bool        NOT NULL,
    created_at                  TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP, -- Timestamp with time zone
    updated_at                  TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP  -- Track last update time
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE users.users;

DROP SCHEMA IF EXISTS users;
-- +goose StatementEnd