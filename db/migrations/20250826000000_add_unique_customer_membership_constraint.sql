-- +goose Up
-- +goose StatementBegin
ALTER TABLE users.customer_membership_plans 
ADD CONSTRAINT unique_customer_membership_plan 
UNIQUE (customer_id, membership_plan_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE users.customer_membership_plans 
DROP CONSTRAINT IF EXISTS unique_customer_membership_plan;
-- +goose StatementEnd