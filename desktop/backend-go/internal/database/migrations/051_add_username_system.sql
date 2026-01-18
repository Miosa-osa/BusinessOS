-- Migration 051: Add username system
-- Adds username column to user table with validation and uniqueness constraints
-- Also creates reserved_usernames table to protect system names

-- Add username columns to user table
ALTER TABLE "user"
ADD COLUMN IF NOT EXISTS username VARCHAR(50),
ADD COLUMN IF NOT EXISTS username_claimed_at TIMESTAMPTZ;

-- Create reserved usernames table
CREATE TABLE IF NOT EXISTS reserved_usernames (
    username VARCHAR(50) PRIMARY KEY,
    reason TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Create case-insensitive unique index for username lookups
-- This prevents usernames that differ only by case (e.g., "John" and "john")
CREATE UNIQUE INDEX IF NOT EXISTS idx_user_username_lower ON "user"(LOWER(username));

-- Create regular index for faster username lookups
CREATE INDEX IF NOT EXISTS idx_user_username ON "user"(username) WHERE username IS NOT NULL;

-- Add constraint to validate username format (alphanumeric, underscore, hyphen)
-- Cannot start or end with hyphen
-- Regex: ^[a-zA-Z0-9]([a-zA-Z0-9_-]{1,48}[a-zA-Z0-9])?$
ALTER TABLE "user"
ADD CONSTRAINT username_format_check
CHECK (username IS NULL OR username ~ '^[a-zA-Z0-9]([a-zA-Z0-9_-]{1,48}[a-zA-Z0-9])?$');

-- Add comments
COMMENT ON COLUMN "user".username IS 'Unique username for the user (3-50 chars, alphanumeric + underscore + hyphen, cannot start/end with hyphen)';
COMMENT ON COLUMN "user".username_claimed_at IS 'Timestamp when username was first claimed';
COMMENT ON TABLE reserved_usernames IS 'Reserved usernames that cannot be claimed by users';

-- Insert reserved usernames
INSERT INTO reserved_usernames (username, reason) VALUES
    ('admin', 'System administrator'),
    ('osa', 'Operating System Agent'),
    ('api', 'API namespace'),
    ('system', 'System namespace'),
    ('app', 'Application namespace'),
    ('apps', 'Applications namespace'),
    ('root', 'Root user'),
    ('support', 'Support team'),
    ('help', 'Help namespace'),
    ('www', 'Web namespace'),
    ('mail', 'Email namespace'),
    ('ftp', 'FTP namespace'),
    ('smtp', 'SMTP namespace'),
    ('info', 'Info namespace'),
    ('login', 'Authentication flow'),
    ('register', 'Registration flow'),
    ('signup', 'Registration flow'),
    ('signin', 'Authentication flow'),
    ('profile', 'Profile namespace'),
    ('settings', 'Settings namespace'),
    ('search', 'Search namespace'),
    ('discover', 'Discovery namespace'),
    ('marketplace', 'Marketplace namespace'),
    ('workspace', 'Workspace namespace'),
    ('team', 'Team namespace'),
    ('project', 'Project namespace'),
    ('task', 'Task namespace'),
    ('dashboard', 'Dashboard namespace'),
    ('about', 'About page'),
    ('contact', 'Contact page'),
    ('terms', 'Terms of service'),
    ('privacy', 'Privacy policy'),
    ('blog', 'Blog namespace'),
    ('docs', 'Documentation'),
    ('status', 'Status page'),
    ('businessos', 'Product name'),
    ('miosa', 'Product name'),
    ('test', 'Testing namespace'),
    ('null', 'Reserved keyword'),
    ('undefined', 'Reserved keyword'),
    ('anonymous', 'Anonymous user'),
    ('guest', 'Guest user'),
    ('deleted', 'Deleted user marker'),
    ('official', 'Official account marker'),
    ('verified', 'Verified account marker')
ON CONFLICT (username) DO NOTHING;
