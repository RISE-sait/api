-- +goose Up
-- +goose StatementBegin

-- Add credit_cost column to events table to allow events to have optional credit pricing
ALTER TABLE events.events ADD COLUMN IF NOT EXISTS credit_cost INTEGER NULL;

-- Add index for efficient querying of events with credit costs
CREATE INDEX IF NOT EXISTS idx_events_credit_cost ON events.events (credit_cost) WHERE credit_cost IS NOT NULL;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Remove the index
DROP INDEX IF EXISTS idx_events_credit_cost;

-- Remove the credit_cost column
ALTER TABLE events.events DROP COLUMN IF EXISTS credit_cost;

-- +goose StatementEnd