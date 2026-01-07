-- +goose Up
-- +goose StatementBegin

-- Audit log for credit refunds when admins remove customers from events
-- Captures full event context snapshot for compliance and auditing
CREATE TABLE IF NOT EXISTS audit.credit_refund_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    customer_id UUID NOT NULL REFERENCES users.users(id) ON DELETE CASCADE,
    event_id UUID REFERENCES events.events(id) ON DELETE SET NULL,
    performed_by UUID NOT NULL REFERENCES users.users(id) ON DELETE CASCADE,

    -- Refund details
    credits_refunded INTEGER NOT NULL,

    -- Event snapshot (preserved if event is later deleted)
    event_name TEXT,
    event_start_at TIMESTAMPTZ,
    program_name TEXT,
    location_name TEXT,

    -- Admin context
    staff_role VARCHAR(50),
    reason TEXT,
    ip_address TEXT,

    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Indexes for efficient querying
CREATE INDEX idx_credit_refund_logs_customer ON audit.credit_refund_logs(customer_id);
CREATE INDEX idx_credit_refund_logs_event ON audit.credit_refund_logs(event_id);
CREATE INDEX idx_credit_refund_logs_performed_by ON audit.credit_refund_logs(performed_by);
CREATE INDEX idx_credit_refund_logs_created_at ON audit.credit_refund_logs(created_at DESC);

COMMENT ON TABLE audit.credit_refund_logs IS 'Audit trail for credit refunds when customers are removed from events';
COMMENT ON COLUMN audit.credit_refund_logs.credits_refunded IS 'Number of credits refunded to customer';
COMMENT ON COLUMN audit.credit_refund_logs.event_name IS 'Snapshot of event name at time of refund';
COMMENT ON COLUMN audit.credit_refund_logs.staff_role IS 'Role of admin who processed the refund';

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP INDEX IF EXISTS idx_credit_refund_logs_customer;
DROP INDEX IF EXISTS idx_credit_refund_logs_event;
DROP INDEX IF EXISTS idx_credit_refund_logs_performed_by;
DROP INDEX IF EXISTS idx_credit_refund_logs_created_at;
DROP TABLE IF EXISTS audit.credit_refund_logs;

-- +goose StatementEnd
