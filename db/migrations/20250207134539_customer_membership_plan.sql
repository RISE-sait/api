-- +goose Up
-- +goose StatementBegin
CREATE TYPE membership_status AS ENUM ('active', 'inactive', 'canceled', 'expired');

CREATE TABLE customer_memberships (
    id uuid NOT NULL DEFAULT gen_random_uuid(),
    customer_id uuid NOT NULL,
    membership_id uuid NOT NULL,
    start_date TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    renewal_date TIMESTAMPTZ,
    status membership_status NOT NULL DEFAULT 'active',
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY(id),
    CONSTRAINT fk_customer FOREIGN KEY(customer_id) REFERENCES customers(user_id),
    CONSTRAINT fk_membership FOREIGN KEY(membership_id) REFERENCES memberships(id),
    CONSTRAINT unique_customer_membership UNIQUE(customer_id, membership_id)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS customer_memberships;
DROP TYPE IF EXISTS membership_status;
-- +goose StatementEnd
