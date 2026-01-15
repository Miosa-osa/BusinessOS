-- Seed Workspace Roles for Testing
-- Run this to add roles to an existing workspace

-- First, find the workspace for the test user
DO $$
DECLARE
    v_workspace_id UUID;
    v_user_id UUID;
BEGIN
    -- Get the test user ID
    SELECT id INTO v_user_id FROM "user" WHERE email = 'testuser@businessos.dev' LIMIT 1;
    
    IF v_user_id IS NULL THEN
        RAISE NOTICE 'Test user not found, creating one...';
        INSERT INTO "user" (id, name, email, "emailVerified") 
        VALUES (gen_random_uuid(), 'Test User', 'testuser@businessos.dev', true)
        RETURNING id INTO v_user_id;
    END IF;
    
    -- Check if workspace exists for this user
    SELECT id INTO v_workspace_id FROM workspaces WHERE owner_id = v_user_id LIMIT 1;
    
    IF v_workspace_id IS NULL THEN
        RAISE NOTICE 'No workspace found, creating one...';
        INSERT INTO workspaces (id, name, slug, owner_id, plan_type, settings, created_at, updated_at)
        VALUES (
            gen_random_uuid(),
            'Test Workspace',
            'test-workspace',
            v_user_id,
            'professional',
            '{"features": {"ai_agents": true, "custom_roles": true}}',
            NOW(),
            NOW()
        )
        RETURNING id INTO v_workspace_id;
        
        -- Add owner as member
        INSERT INTO workspace_members (id, workspace_id, user_id, role_id, role_name, status, joined_at, created_at, updated_at)
        VALUES (
            gen_random_uuid(),
            v_workspace_id,
            v_user_id,
            NULL,
            'owner',
            'active',
            NOW(),
            NOW(),
            NOW()
        );
    END IF;
    
    RAISE NOTICE 'Using workspace: %', v_workspace_id;
    
    -- Delete existing roles for this workspace (clean slate)
    DELETE FROM workspace_roles WHERE workspace_id = v_workspace_id;
    
    -- Insert system roles
    INSERT INTO workspace_roles (id, workspace_id, name, display_name, description, color, icon, hierarchy_level, is_system, is_default, permissions, created_at, updated_at)
    VALUES
    -- Owner role (highest level)
    (
        gen_random_uuid(),
        v_workspace_id,
        'owner',
        'Owner',
        'Full control over the workspace. Can manage all settings, members, and billing.',
        '#10B981',
        'crown',
        100,
        true,
        false,
        '{
            "workspace": {"view": true, "manage": true, "delete": true, "transfer": true},
            "members": {"view": true, "invite": true, "remove": true, "manage": true},
            "roles": {"view": true, "create": true, "edit": true, "delete": true, "assign": true},
            "projects": {"view": true, "create": true, "edit": true, "delete": true, "archive": true},
            "tasks": {"view": true, "create": true, "edit": true, "delete": true, "assign": true},
            "contexts": {"view": true, "create": true, "edit": true, "delete": true, "share": true},
            "clients": {"view": true, "create": true, "edit": true, "delete": true},
            "artifacts": {"view": true, "create": true, "edit": true, "delete": true},
            "ai_agents": {"view": true, "create": true, "edit": true, "delete": true, "execute": true}
        }'::jsonb,
        NOW(),
        NOW()
    ),
    -- Admin role
    (
        gen_random_uuid(),
        v_workspace_id,
        'admin',
        'Admin',
        'Can manage most workspace settings and members. Cannot delete workspace or transfer ownership.',
        '#3B82F6',
        'shield',
        80,
        true,
        false,
        '{
            "workspace": {"view": true, "manage": true, "delete": false, "transfer": false},
            "members": {"view": true, "invite": true, "remove": true, "manage": true},
            "roles": {"view": true, "create": true, "edit": true, "delete": false, "assign": true},
            "projects": {"view": true, "create": true, "edit": true, "delete": true, "archive": true},
            "tasks": {"view": true, "create": true, "edit": true, "delete": true, "assign": true},
            "contexts": {"view": true, "create": true, "edit": true, "delete": true, "share": true},
            "clients": {"view": true, "create": true, "edit": true, "delete": true},
            "artifacts": {"view": true, "create": true, "edit": true, "delete": true},
            "ai_agents": {"view": true, "create": true, "edit": true, "delete": true, "execute": true}
        }'::jsonb,
        NOW(),
        NOW()
    ),
    -- Manager role
    (
        gen_random_uuid(),
        v_workspace_id,
        'manager',
        'Manager',
        'Can manage projects, tasks, and team members within their assigned areas.',
        '#8B5CF6',
        'users',
        60,
        true,
        false,
        '{
            "workspace": {"view": true, "manage": false, "delete": false, "transfer": false},
            "members": {"view": true, "invite": true, "remove": false, "manage": false},
            "roles": {"view": true, "create": false, "edit": false, "delete": false, "assign": false},
            "projects": {"view": true, "create": true, "edit": true, "delete": false, "archive": true},
            "tasks": {"view": true, "create": true, "edit": true, "delete": true, "assign": true},
            "contexts": {"view": true, "create": true, "edit": true, "delete": false, "share": true},
            "clients": {"view": true, "create": true, "edit": true, "delete": false},
            "artifacts": {"view": true, "create": true, "edit": true, "delete": false},
            "ai_agents": {"view": true, "create": false, "edit": false, "delete": false, "execute": true}
        }'::jsonb,
        NOW(),
        NOW()
    ),
    -- Member role (default)
    (
        gen_random_uuid(),
        v_workspace_id,
        'member',
        'Member',
        'Standard workspace member. Can create and edit their own content.',
        '#F59E0B',
        'user',
        40,
        true,
        true,
        '{
            "workspace": {"view": true, "manage": false, "delete": false, "transfer": false},
            "members": {"view": true, "invite": false, "remove": false, "manage": false},
            "roles": {"view": true, "create": false, "edit": false, "delete": false, "assign": false},
            "projects": {"view": true, "create": true, "edit": true, "delete": false, "archive": false},
            "tasks": {"view": true, "create": true, "edit": true, "delete": false, "assign": false},
            "contexts": {"view": true, "create": true, "edit": true, "delete": false, "share": false},
            "clients": {"view": true, "create": true, "edit": false, "delete": false},
            "artifacts": {"view": true, "create": true, "edit": true, "delete": false},
            "ai_agents": {"view": true, "create": false, "edit": false, "delete": false, "execute": true}
        }'::jsonb,
        NOW(),
        NOW()
    ),
    -- Viewer role (read-only)
    (
        gen_random_uuid(),
        v_workspace_id,
        'viewer',
        'Viewer',
        'Read-only access to workspace content. Cannot create or modify anything.',
        '#6B7280',
        'eye',
        20,
        true,
        false,
        '{
            "workspace": {"view": true, "manage": false, "delete": false, "transfer": false},
            "members": {"view": true, "invite": false, "remove": false, "manage": false},
            "roles": {"view": true, "create": false, "edit": false, "delete": false, "assign": false},
            "projects": {"view": true, "create": false, "edit": false, "delete": false, "archive": false},
            "tasks": {"view": true, "create": false, "edit": false, "delete": false, "assign": false},
            "contexts": {"view": true, "create": false, "edit": false, "delete": false, "share": false},
            "clients": {"view": true, "create": false, "edit": false, "delete": false},
            "artifacts": {"view": true, "create": false, "edit": false, "delete": false},
            "ai_agents": {"view": true, "create": false, "edit": false, "delete": false, "execute": false}
        }'::jsonb,
        NOW(),
        NOW()
    ),
    -- Custom role example: Contractor
    (
        gen_random_uuid(),
        v_workspace_id,
        'contractor',
        'Contractor',
        'External contractor with limited access to assigned projects only.',
        '#EC4899',
        'briefcase',
        30,
        false,
        false,
        '{
            "workspace": {"view": true, "manage": false, "delete": false, "transfer": false},
            "members": {"view": false, "invite": false, "remove": false, "manage": false},
            "roles": {"view": false, "create": false, "edit": false, "delete": false, "assign": false},
            "projects": {"view": true, "create": false, "edit": true, "delete": false, "archive": false},
            "tasks": {"view": true, "create": true, "edit": true, "delete": false, "assign": false},
            "contexts": {"view": true, "create": true, "edit": true, "delete": false, "share": false},
            "clients": {"view": false, "create": false, "edit": false, "delete": false},
            "artifacts": {"view": true, "create": true, "edit": true, "delete": false},
            "ai_agents": {"view": false, "create": false, "edit": false, "delete": false, "execute": false}
        }'::jsonb,
        NOW(),
        NOW()
    );
    
    RAISE NOTICE 'Successfully seeded 6 workspace roles!';
END $$;

-- Verify the roles were created
SELECT 
    name, 
    display_name, 
    color, 
    hierarchy_level, 
    is_system, 
    is_default 
FROM workspace_roles 
ORDER BY hierarchy_level DESC;
