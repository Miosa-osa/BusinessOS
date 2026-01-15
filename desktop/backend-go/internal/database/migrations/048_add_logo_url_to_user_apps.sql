-- Migration 048: Add logo_url to user_external_apps
-- Allows storing actual app logos (favicons) instead of just Lucide icons

ALTER TABLE user_external_apps
ADD COLUMN logo_url TEXT;

COMMENT ON COLUMN user_external_apps.logo_url IS 'URL to app logo/favicon - fetched automatically from app URL';
