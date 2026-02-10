-- Migration: Create parent_link_requests table for family linkage
-- This table tracks requests to link children to parents (new links and transfers)

CREATE TABLE users.parent_link_requests (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    -- The child being linked/transferred
    child_id UUID NOT NULL REFERENCES users.users(id) ON DELETE CASCADE,

    -- The new parent being linked to
    new_parent_id UUID NOT NULL REFERENCES users.users(id) ON DELETE CASCADE,

    -- The old parent (only for transfers, NULL for new links)
    old_parent_id UUID REFERENCES users.users(id) ON DELETE SET NULL,

    -- Who initiated the request: 'child' or 'parent'
    -- If 'child': verification code sent to new_parent
    -- If 'parent': verification code sent to child
    initiated_by VARCHAR(10) NOT NULL CHECK (initiated_by IN ('child', 'parent')),

    -- Verification code for the non-initiating party
    verification_code VARCHAR(6) NOT NULL,
    verified_at TIMESTAMPTZ,

    -- For transfers: old parent must also verify with separate code
    old_parent_code VARCHAR(6),
    old_parent_verified_at TIMESTAMPTZ,

    -- Lifecycle timestamps
    expires_at TIMESTAMPTZ NOT NULL,
    completed_at TIMESTAMPTZ,
    cancelled_at TIMESTAMPTZ,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- Prevent duplicate pending requests per child
    -- Uses NULLS NOT DISTINCT to treat NULL as equal for uniqueness
    CONSTRAINT unique_pending_child_request
        UNIQUE NULLS NOT DISTINCT (child_id, completed_at, cancelled_at),

    -- Child cannot be the same as either parent
    CONSTRAINT child_not_same_as_parent
        CHECK (child_id != new_parent_id AND (old_parent_id IS NULL OR child_id != old_parent_id)),

    -- New parent cannot be the same as old parent
    CONSTRAINT parents_must_differ
        CHECK (old_parent_id IS NULL OR new_parent_id != old_parent_id)
);

-- Index for efficient verification code lookups (only active requests)
CREATE INDEX idx_parent_link_verification_code
    ON users.parent_link_requests(verification_code)
    WHERE completed_at IS NULL AND cancelled_at IS NULL;

-- Index for old parent code lookups (only active requests with old parent)
CREATE INDEX idx_parent_link_old_parent_code
    ON users.parent_link_requests(old_parent_code)
    WHERE completed_at IS NULL AND cancelled_at IS NULL AND old_parent_code IS NOT NULL;

-- Index for finding pending requests by child
CREATE INDEX idx_parent_link_child_pending
    ON users.parent_link_requests(child_id)
    WHERE completed_at IS NULL AND cancelled_at IS NULL;

-- Index for finding requests involving a user (as new or old parent)
CREATE INDEX idx_parent_link_new_parent
    ON users.parent_link_requests(new_parent_id)
    WHERE completed_at IS NULL AND cancelled_at IS NULL;

CREATE INDEX idx_parent_link_old_parent
    ON users.parent_link_requests(old_parent_id)
    WHERE completed_at IS NULL AND cancelled_at IS NULL AND old_parent_id IS NOT NULL;

COMMENT ON TABLE users.parent_link_requests IS 'Tracks parent-child link requests including transfers between parents';
COMMENT ON COLUMN users.parent_link_requests.initiated_by IS 'Who started the request: child or parent';
COMMENT ON COLUMN users.parent_link_requests.verification_code IS 'Code sent to the non-initiating party for confirmation';
COMMENT ON COLUMN users.parent_link_requests.old_parent_code IS 'For transfers: separate code sent to old parent for approval';
