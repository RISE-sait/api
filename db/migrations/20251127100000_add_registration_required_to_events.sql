-- +goose Up
-- Add registration_required column to events table
-- When false, the event cannot be enrolled in (informational only)
ALTER TABLE events.events ADD COLUMN registration_required BOOLEAN NOT NULL DEFAULT true;

-- +goose Down
ALTER TABLE events.events DROP COLUMN registration_required;
