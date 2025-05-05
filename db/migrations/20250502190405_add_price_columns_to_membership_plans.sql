-- +goose Up
ALTER TABLE membership.membership_plans
ADD COLUMN unit_amount INTEGER,
ADD COLUMN currency VARCHAR(10),
ADD COLUMN interval VARCHAR(10);

-- +goose Down
ALTER TABLE membership.membership_plans
DROP COLUMN unit_amount,
DROP COLUMN currency,
DROP COLUMN interval;
