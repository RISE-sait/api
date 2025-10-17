-- +goose Up
-- +goose StatementBegin

-- Create junction table for multiple membership plans per event
CREATE TABLE events.event_membership_access (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_id UUID NOT NULL REFERENCES events.events(id) ON DELETE CASCADE,
    membership_plan_id UUID NOT NULL REFERENCES membership.membership_plans(id) ON DELETE CASCADE,
    created_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(event_id, membership_plan_id)
);

-- Create index for faster lookups
CREATE INDEX idx_event_membership_access_event_id ON events.event_membership_access(event_id);
CREATE INDEX idx_event_membership_access_membership_plan_id ON events.event_membership_access(membership_plan_id);

-- Migrate existing data from required_membership_plan_id to the new junction table
INSERT INTO events.event_membership_access (event_id, membership_plan_id)
SELECT id, required_membership_plan_id
FROM events.events
WHERE required_membership_plan_id IS NOT NULL;

-- Drop the old single membership column
ALTER TABLE events.events DROP COLUMN required_membership_plan_id;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Restore the old column
ALTER TABLE events.events ADD COLUMN required_membership_plan_id UUID REFERENCES membership.membership_plans(id);

-- Migrate back the first membership plan from junction table (if multiple exist, only one will be kept)
UPDATE events.events e
SET required_membership_plan_id = (
    SELECT membership_plan_id
    FROM events.event_membership_access
    WHERE event_id = e.id
    LIMIT 1
);

-- Drop junction table
DROP TABLE IF EXISTS events.event_membership_access;

-- +goose StatementEnd

