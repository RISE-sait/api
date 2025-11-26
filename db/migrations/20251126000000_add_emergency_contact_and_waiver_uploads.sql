-- +goose Up
-- Add emergency contact fields to users table
ALTER TABLE users.users
ADD COLUMN IF NOT EXISTS emergency_contact_name VARCHAR(100),
ADD COLUMN IF NOT EXISTS emergency_contact_phone VARCHAR(25),
ADD COLUMN IF NOT EXISTS emergency_contact_relationship VARCHAR(50);

-- Create table for uploaded waiver documents (signed physical waivers scanned and uploaded)
CREATE TABLE IF NOT EXISTS waiver.waiver_uploads (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users.users(id) ON DELETE CASCADE,
    file_url TEXT NOT NULL,
    file_name VARCHAR(255) NOT NULL,
    file_type VARCHAR(50) NOT NULL,
    file_size_bytes BIGINT,
    uploaded_by UUID REFERENCES users.users(id),
    notes TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_waiver_uploads_user_id ON waiver.waiver_uploads(user_id);

-- +goose Down
DROP INDEX IF EXISTS waiver.idx_waiver_uploads_user_id;
DROP TABLE IF EXISTS waiver.waiver_uploads;
ALTER TABLE users.users
DROP COLUMN IF EXISTS emergency_contact_name,
DROP COLUMN IF EXISTS emergency_contact_phone,
DROP COLUMN IF EXISTS emergency_contact_relationship;
