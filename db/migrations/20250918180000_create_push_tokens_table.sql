-- +goose Up
-- Create notifications schema if it doesn't exist
CREATE SCHEMA IF NOT EXISTS notifications;

-- Create push_tokens table for storing Expo push notification tokens
CREATE TABLE IF NOT EXISTS notifications.push_tokens (
    id SERIAL PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users.users(id) ON DELETE CASCADE,
    expo_push_token TEXT NOT NULL,
    device_type VARCHAR(10) CHECK (device_type IN ('ios', 'android')),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, device_type)
);

-- Add index for faster lookups
CREATE INDEX IF NOT EXISTS idx_push_tokens_user_id ON notifications.push_tokens(user_id);
CREATE INDEX IF NOT EXISTS idx_push_tokens_device_type ON notifications.push_tokens(device_type);

-- +goose Down
DROP TABLE IF EXISTS notifications.push_tokens;
DROP SCHEMA IF EXISTS notifications CASCADE;