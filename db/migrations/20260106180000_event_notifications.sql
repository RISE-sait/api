-- +goose Up
-- +goose StatementBegin

-- Event Notification History
-- Tracks notifications sent to event attendees for auditing purposes

CREATE TABLE IF NOT EXISTS events.notification_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_id UUID NOT NULL REFERENCES events.events(id) ON DELETE CASCADE,
    sent_by UUID NOT NULL REFERENCES users.users(id),
    channel VARCHAR(20) NOT NULL CHECK (channel IN ('email', 'push', 'both')),
    subject VARCHAR(255),
    message TEXT NOT NULL,
    include_event_details BOOLEAN NOT NULL DEFAULT false,
    recipient_count INT NOT NULL DEFAULT 0,
    email_success_count INT NOT NULL DEFAULT 0,
    email_failure_count INT NOT NULL DEFAULT 0,
    push_success_count INT NOT NULL DEFAULT 0,
    push_failure_count INT NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Index for querying notification history by event
CREATE INDEX IF NOT EXISTS idx_notification_history_event_id ON events.notification_history(event_id);

-- Index for querying notification history by sender
CREATE INDEX IF NOT EXISTS idx_notification_history_sent_by ON events.notification_history(sent_by);

COMMENT ON TABLE events.notification_history IS 'Tracks notifications sent to event attendees';
COMMENT ON COLUMN events.notification_history.channel IS 'Notification channel: email, push, or both';
COMMENT ON COLUMN events.notification_history.include_event_details IS 'Whether event details were automatically included in the message';

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS events.notification_history;

-- +goose StatementEnd
