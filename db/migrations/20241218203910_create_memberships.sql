-- +goose Up
-- +goose StatementBegin
CREATE SCHEMA IF NOT EXISTS membership;

CREATE TABLE IF NOT EXISTS membership.memberships
(
    id          UUID PRIMARY KEY         DEFAULT gen_random_uuid(),
    name       VARCHAR(150)             NOT NULL,
    description TEXT,
    benefits varchar(750) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS membership.membership_plans
(
    id                UUID PRIMARY KEY         DEFAULT gen_random_uuid(),
    name              VARCHAR(150)             NOT NULL,
    stripe_price_id       varchar(50) NOT NULL,
    stripe_joining_fee_id varchar(50),
    membership_id     UUID                     NOT NULL,
    amt_periods       INT,
    created_at        TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at        TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_membership
        FOREIGN KEY (membership_id)
            REFERENCES membership.memberships (id),
    CONSTRAINT unique_membership_name
        UNIQUE (membership_id, name),
    CONSTRAINT unique_stripe_price_id
        UNIQUE (stripe_price_id)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP SCHEMA IF EXISTS membership cascade;
-- +goose StatementEnd
