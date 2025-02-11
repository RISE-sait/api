-- +goose Up
-- +goose StatementBegin
CREATE TYPE payment_frequency AS ENUM ('once', 'week', 'month', 'day');

CREATE TABLE membership_plans (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(50) NOT NULL,
    price int NOT NULL,
    joining_fee int,
    auto_renew BOOLEAN NOT NULL DEFAULT FALSE,
    membership_id UUID NOT NULL,
    payment_frequency payment_frequency,
    amt_periods INT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_membership
        FOREIGN KEY(membership_id) 
        REFERENCES memberships(id),
    CONSTRAINT unique_membership_name
        UNIQUE (membership_id, name)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS membership_plans;
DROP TYPE IF EXISTS payment_frequency;
-- +goose StatementEnd