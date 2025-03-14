-- +goose Up
-- +goose StatementBegin

CREATE SCHEMA IF NOT EXISTS membership;

CREATE TABLE membership.memberships
(
    id          UUID PRIMARY KEY         DEFAULT gen_random_uuid(),
    name       VARCHAR(150)             NOT NULL,
    description TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TYPE payment_frequency AS ENUM ('once', 'week', 'month', 'day');

CREATE TABLE membership.membership_plans
(
    id                UUID PRIMARY KEY         DEFAULT gen_random_uuid(),
    name              VARCHAR(150)             NOT NULL,
    price             decimal(6, 2)            NOT NULL,
    joining_fee       decimal(6, 2)            NOT NULL,
    auto_renew        BOOLEAN                  NOT NULL DEFAULT FALSE,
    membership_id     UUID                     NOT NULL,
    payment_frequency payment_frequency        NOT NULL,
    amt_periods       INT,
    created_at        TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at        TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_membership
        FOREIGN KEY (membership_id)
            REFERENCES membership.memberships (id),
    CONSTRAINT unique_membership_name
        UNIQUE (membership_id, name)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS membership.membership_plans;
DROP TYPE IF EXISTS payment_frequency;

DROP TABLE IF EXISTS membership.memberships;

DROP SCHEMA IF EXISTS membership;
-- +goose StatementEnd
