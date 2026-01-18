-- Migration 050: Add OSA Build onboarding columns
-- Adds support for user analysis and starter apps generation

-- Add columns for OSA Build onboarding analysis and starter apps
ALTER TABLE workspace_onboarding_profiles
ADD COLUMN IF NOT EXISTS user_id TEXT,
ADD COLUMN IF NOT EXISTS analysis_data JSONB,
ADD COLUMN IF NOT EXISTS starter_apps_data JSONB;

-- Add index for user lookups
CREATE INDEX IF NOT EXISTS idx_workspace_onboarding_user ON workspace_onboarding_profiles(user_id);

-- Add comment
COMMENT ON COLUMN workspace_onboarding_profiles.analysis_data IS 'User analysis results from OSA Build onboarding (insights, interests, tools)';
COMMENT ON COLUMN workspace_onboarding_profiles.starter_apps_data IS 'Personalized starter apps generated during onboarding';
