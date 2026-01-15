-- Migration 049: Fix user_external_apps workspace foreign key
-- The workspace_id references 'workspaces' but the frontend uses workspace IDs
-- from different sources. Removing the FK constraint for flexibility.

-- Drop the existing foreign key constraint if it exists
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.table_constraints
        WHERE constraint_name = 'user_external_apps_workspace_id_fkey'
        AND table_name = 'user_external_apps'
    ) THEN
        ALTER TABLE user_external_apps DROP CONSTRAINT user_external_apps_workspace_id_fkey;
    END IF;
END $$;

-- Create a default workspace if it doesn't exist
-- This workspace will be used as a fallback for the 3D desktop
INSERT INTO workspaces (id, name, slug, owner_id, settings)
SELECT
    'a438da8e-d245-4e7a-bde2-3a053499ab87'::uuid,
    'Default Desktop Workspace',
    'default-desktop',
    'J4OG6wYzz6Mz4YcBQee8luJhxeNw2z9A', -- roberto@accelerants.ai
    '{}'::jsonb
WHERE NOT EXISTS (
    SELECT 1 FROM workspaces WHERE id = 'a438da8e-d245-4e7a-bde2-3a053499ab87'::uuid
);

-- Add a comment explaining the change
COMMENT ON COLUMN user_external_apps.workspace_id IS 'Workspace UUID - no longer constrained to workspaces table for flexibility';
