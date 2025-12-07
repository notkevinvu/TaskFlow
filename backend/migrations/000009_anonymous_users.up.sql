-- Migration: Anonymous User Support
-- Enables users to try TaskFlow without registration
-- Anonymous users have limited features and expire after 30 days

-- Add user_type enum
CREATE TYPE user_type AS ENUM ('registered', 'anonymous');

-- Add user_type column to users table (default to 'registered' for existing users)
ALTER TABLE users
    ADD COLUMN user_type user_type NOT NULL DEFAULT 'registered';

-- Add expiry tracking for anonymous users (NULL for registered users)
ALTER TABLE users
    ADD COLUMN expires_at TIMESTAMP WITH TIME ZONE;

-- Make auth fields nullable for anonymous users
ALTER TABLE users
    ALTER COLUMN email DROP NOT NULL,
    ALTER COLUMN name DROP NOT NULL,
    ALTER COLUMN password_hash DROP NOT NULL;

-- Add constraint: registered users MUST have email/password/name
-- Anonymous users do NOT need these fields
ALTER TABLE users ADD CONSTRAINT users_type_credentials_check
    CHECK (
        (user_type = 'registered' AND email IS NOT NULL AND password_hash IS NOT NULL AND name IS NOT NULL)
        OR (user_type = 'anonymous')
    );

-- Add constraint: anonymous users MUST have expires_at, registered users must NOT
ALTER TABLE users ADD CONSTRAINT users_type_expiry_check
    CHECK (
        (user_type = 'registered' AND expires_at IS NULL)
        OR (user_type = 'anonymous' AND expires_at IS NOT NULL)
    );

-- Index for cleanup job to efficiently find expired anonymous users
CREATE INDEX idx_users_anonymous_expires_at ON users(expires_at)
    WHERE user_type = 'anonymous' AND expires_at IS NOT NULL;

-- Index for user type filtering
CREATE INDEX idx_users_user_type ON users(user_type);

-- Create cleanup audit table to track deleted anonymous users
CREATE TABLE anonymous_user_cleanups (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL,
    task_count INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,  -- When the anonymous user was created
    deleted_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Index for cleanup audit queries
CREATE INDEX idx_anonymous_cleanups_deleted_at ON anonymous_user_cleanups(deleted_at DESC);

-- Note: Existing users are automatically 'registered' due to DEFAULT value
-- No data migration needed - default value handles backfill
