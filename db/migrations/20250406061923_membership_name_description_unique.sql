-- +goose Up
-- +goose StatementBegin
-- Safely drop constraint if it exists
ALTER TABLE membership.memberships
    ADD CONSTRAINT unique_membership_table_name UNIQUE (name);

-- Make description not null
ALTER TABLE membership.memberships
    ALTER COLUMN description SET NOT NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE membership.memberships
    DROP CONSTRAINT IF EXISTS unique_membership_table_name;

ALTER TABLE membership.memberships
    ALTER COLUMN description DROP NOT NULL;
-- +goose StatementEnd