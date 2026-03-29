-- Seed Development Workspace
-- This creates the mock workspace used by the frontend in dev mode
-- Run with: psql $DATABASE_URL -f scripts/seed_dev_workspace.sql
-- Or: go run scripts/seed_dev_workspace.go

BEGIN;

-- The mock workspace ID used by frontend
-- This must match frontend/src/lib/stores/workspaces.ts
DO $$
DECLARE
    v_workspace_id UUID := '00000000-0000-0000-0000-000000000001';
    v_owner_role_id UUID;
    v_user_id VARCHAR(255);
BEGIN
    -- Check if workspace already exists
    IF EXISTS (SELECT 1 FROM workspaces WHERE id = v_workspace_id) THEN
        RAISE NOTICE 'Dev workspace already exists, skipping creation';
    ELSE
        -- Create the dev workspace
        INSERT INTO workspaces (
            id, name, slug, description, plan_type,
            max_members, max_projects, max_storage_gb,
            owner_id, settings, created_at, updated_at
        ) VALUES (
            v_workspace_id,
            'Development Workspace',
            'dev-workspace',
            'Auto-created workspace for local development and testing',
            'professional',
            50, 100, 50,
            'system', -- Will be updated when first user joins
            '{"dev_mode": true}'::jsonb,
            NOW(), NOW()
        );
        RAISE NOTICE 'Created dev workspace: %', v_workspace_id;
    END IF;

    -- Create owner role if it doesn't exist
    IF NOT EXISTS (
        SELECT 1 FROM workspace_roles
        WHERE workspace_id = v_workspace_id AND name = 'owner'
    ) THEN
        INSERT INTO workspace_roles (
            id, workspace_id, name, display_name, description,
            color, icon, hierarchy_level, is_system, is_default,
            permissions, created_at, updated_at
        ) VALUES (
            gen_random_uuid(),
            v_workspace_id,
            'owner',
            'Owner',
            'Full workspace access',
            '#10b981',
            'crown',
            0,
            true,
            false,
            '{
                "workspace": {"manage": true, "delete": true, "transfer": true},
                "members": {"view": true, "invite": true, "manage": true, "remove": true},
                "roles": {"view": true, "create": true, "manage": true, "delete": true},
                "projects": {"view": true, "create": true, "manage": true, "delete": true},
                "settings": {"view": true, "manage": true},
                "billing": {"view": true, "manage": true},
                "apps": {"view": true, "create": true, "manage": true, "delete": true}
            }'::jsonb,
            NOW(), NOW()
        );
        RAISE NOTICE 'Created owner role for dev workspace';
    END IF;

    -- Get the owner role ID
    SELECT id INTO v_owner_role_id
    FROM workspace_roles
    WHERE workspace_id = v_workspace_id AND name = 'owner';

    -- Add all existing users to the dev workspace if they aren't already members
    -- This ensures any authenticated user can access the dev workspace
    FOR v_user_id IN
        SELECT u.id FROM users u
        WHERE NOT EXISTS (
            SELECT 1 FROM workspace_members wm
            WHERE wm.workspace_id = v_workspace_id AND wm.user_id = u.id
        )
    LOOP
        INSERT INTO workspace_members (
            workspace_id, user_id, role_id, role_name, status, joined_at
        ) VALUES (
            v_workspace_id,
            v_user_id,
            v_owner_role_id,
            'owner',
            'active',
            NOW()
        );
        RAISE NOTICE 'Added user % to dev workspace as owner', v_user_id;

        -- Update workspace owner_id to first user if it's still 'system'
        UPDATE workspaces
        SET owner_id = v_user_id
        WHERE id = v_workspace_id AND owner_id = 'system';
    END LOOP;

END $$;

-- Verify
SELECT
    w.id,
    w.name,
    w.slug,
    COUNT(wm.user_id) as member_count
FROM workspaces w
LEFT JOIN workspace_members wm ON w.id = wm.workspace_id
WHERE w.id = '00000000-0000-0000-0000-000000000001'
GROUP BY w.id, w.name, w.slug;

COMMIT;
