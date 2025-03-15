-- +goose Up
-- +goose StatementBegin
ALTER TABLE public.customer_membership_plans
    ALTER COLUMN membership_plan_id DROP NOT NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE public.customer_membership_plans
    ALTER COLUMN membership_plan_id SET NOT NULL;
-- +goose StatementEnd
