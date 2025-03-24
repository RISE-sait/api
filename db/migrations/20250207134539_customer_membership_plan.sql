-- +goose Up
-- +goose StatementBegin
CREATE TYPE membership_status AS ENUM ('active', 'inactive', 'canceled', 'expired');

CREATE TABLE IF NOT EXISTS customer_membership_plans
(
    id uuid NOT NULL PRIMARY KEY DEFAULT gen_random_uuid(),
    customer_id uuid NOT NULL,
    membership_plan_id uuid NOT NULL,
    start_date TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    renewal_date TIMESTAMPTZ,
    status membership_status NOT NULL DEFAULT 'active',
    created_at timestamp with time zone NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp with time zone NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_customer FOREIGN KEY(customer_id) REFERENCES users.users(id),
    CONSTRAINT fk_membership_plan FOREIGN KEY(membership_plan_id) REFERENCES membership.membership_plans(id)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS customer_membership_plans;
DROP TYPE IF EXISTS membership_status;
-- +goose StatementEnd
