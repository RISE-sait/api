-- +goose Up
-- +goose StatementBegin
ALTER TABLE events.events
    ADD COLUMN required_membership_plan_id UUID REFERENCES membership.membership_plans (id),
    ADD COLUMN price_id TEXT;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE events.events
    DROP COLUMN IF EXISTS required_membership_plan_id,
    DROP COLUMN IF EXISTS price_id;
-- +goose StatementEnd