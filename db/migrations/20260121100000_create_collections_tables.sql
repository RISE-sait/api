-- +goose Up
-- +goose StatementBegin

-- Collection attempts audit log for tracking all collection activities
CREATE TABLE IF NOT EXISTS payments.collection_attempts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    customer_id UUID NOT NULL REFERENCES users.users(id) ON DELETE CASCADE,
    admin_id UUID NOT NULL REFERENCES users.users(id) ON DELETE CASCADE,

    -- Amount information
    amount_attempted DECIMAL(10,2) NOT NULL,
    amount_collected DECIMAL(10,2) DEFAULT 0,

    -- Collection method: 'card_charge', 'payment_link', 'manual_entry'
    collection_method TEXT NOT NULL,
    -- For card: masked card info like "Visa ending in 4242"
    -- For manual: 'cash', 'check', 'external_card', etc.
    payment_method_details TEXT,

    -- Status: 'pending', 'success', 'failed', 'disputed'
    status TEXT NOT NULL DEFAULT 'pending',
    failure_reason TEXT,

    -- Stripe references
    stripe_payment_intent_id TEXT,
    stripe_payment_link_id TEXT,
    stripe_customer_id TEXT,

    -- Related records
    membership_plan_id UUID REFERENCES membership.membership_plans(id) ON DELETE SET NULL,
    stripe_subscription_id TEXT,

    -- Audit information
    notes TEXT,
    previous_balance DECIMAL(10,2),
    new_balance DECIMAL(10,2),

    -- Timestamps
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    completed_at TIMESTAMPTZ,

    -- Constraints
    CONSTRAINT valid_collection_method CHECK (
        collection_method IN ('card_charge', 'payment_link', 'manual_entry')
    ),
    CONSTRAINT valid_collection_status CHECK (
        status IN ('pending', 'success', 'failed', 'disputed')
    ),
    CONSTRAINT positive_amounts CHECK (
        amount_attempted >= 0 AND amount_collected >= 0
    )
);

-- Indexes for collection_attempts
CREATE INDEX idx_collection_attempts_customer_id ON payments.collection_attempts(customer_id);
CREATE INDEX idx_collection_attempts_admin_id ON payments.collection_attempts(admin_id);
CREATE INDEX idx_collection_attempts_status ON payments.collection_attempts(status);
CREATE INDEX idx_collection_attempts_created_at ON payments.collection_attempts(created_at DESC);
CREATE INDEX idx_collection_attempts_method ON payments.collection_attempts(collection_method);

-- Payment links tracking for sent payment requests
CREATE TABLE IF NOT EXISTS payments.payment_links (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    customer_id UUID NOT NULL REFERENCES users.users(id) ON DELETE CASCADE,
    admin_id UUID NOT NULL REFERENCES users.users(id) ON DELETE CASCADE,

    -- Stripe payment link info
    stripe_payment_link_id TEXT NOT NULL,
    stripe_payment_link_url TEXT NOT NULL,

    -- Amount and purpose
    amount DECIMAL(10,2) NOT NULL,
    description TEXT,

    -- Related records
    membership_plan_id UUID REFERENCES membership.membership_plans(id) ON DELETE SET NULL,
    collection_attempt_id UUID REFERENCES payments.collection_attempts(id) ON DELETE SET NULL,

    -- Status: 'pending', 'sent', 'opened', 'completed', 'expired'
    status TEXT NOT NULL DEFAULT 'pending',

    -- Delivery tracking
    sent_via TEXT[], -- Array of delivery methods: 'email', 'sms'
    sent_to_email TEXT,
    sent_to_phone TEXT,

    -- Timestamps
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    sent_at TIMESTAMPTZ,
    opened_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    expires_at TIMESTAMPTZ,

    -- Constraints
    CONSTRAINT valid_link_status CHECK (
        status IN ('pending', 'sent', 'opened', 'completed', 'expired')
    ),
    CONSTRAINT positive_amount CHECK (amount >= 0)
);

-- Indexes for payment_links
CREATE INDEX idx_payment_links_customer_id ON payments.payment_links(customer_id);
CREATE INDEX idx_payment_links_admin_id ON payments.payment_links(admin_id);
CREATE INDEX idx_payment_links_status ON payments.payment_links(status);
CREATE INDEX idx_payment_links_created_at ON payments.payment_links(created_at DESC);
CREATE INDEX idx_payment_links_stripe_id ON payments.payment_links(stripe_payment_link_id);

-- Unique constraint on Stripe payment link ID
CREATE UNIQUE INDEX idx_payment_links_unique_stripe_id
ON payments.payment_links(stripe_payment_link_id)
WHERE stripe_payment_link_id IS NOT NULL;

-- Update timestamp triggers
CREATE OR REPLACE FUNCTION payments.update_collection_attempts_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_collection_attempts_updated_at
    BEFORE UPDATE ON payments.collection_attempts
    FOR EACH ROW
    EXECUTE FUNCTION payments.update_collection_attempts_updated_at();

-- Comments for documentation
COMMENT ON TABLE payments.collection_attempts IS 'Audit log of all payment collection attempts by admins';
COMMENT ON TABLE payments.payment_links IS 'Tracking of payment links sent to customers for collection';
COMMENT ON COLUMN payments.collection_attempts.collection_method IS 'Method used: card_charge, payment_link, or manual_entry';
COMMENT ON COLUMN payments.collection_attempts.payment_method_details IS 'Details like masked card info or cash/check type';

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TRIGGER IF EXISTS update_collection_attempts_updated_at ON payments.collection_attempts;
DROP FUNCTION IF EXISTS payments.update_collection_attempts_updated_at();

DROP INDEX IF EXISTS payments.idx_payment_links_unique_stripe_id;
DROP INDEX IF EXISTS payments.idx_payment_links_stripe_id;
DROP INDEX IF EXISTS payments.idx_payment_links_created_at;
DROP INDEX IF EXISTS payments.idx_payment_links_status;
DROP INDEX IF EXISTS payments.idx_payment_links_admin_id;
DROP INDEX IF EXISTS payments.idx_payment_links_customer_id;

DROP INDEX IF EXISTS payments.idx_collection_attempts_method;
DROP INDEX IF EXISTS payments.idx_collection_attempts_created_at;
DROP INDEX IF EXISTS payments.idx_collection_attempts_status;
DROP INDEX IF EXISTS payments.idx_collection_attempts_admin_id;
DROP INDEX IF EXISTS payments.idx_collection_attempts_customer_id;

DROP TABLE IF EXISTS payments.payment_links;
DROP TABLE IF EXISTS payments.collection_attempts;

-- +goose StatementEnd
