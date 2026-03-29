-- ============================================================================
-- SEED TEST DATA FOR DASHBOARD TESTING
-- ============================================================================
-- This script creates sample tasks and projects for a test user
-- to enable realistic dashboard widget testing
--
-- Usage:
--   psql $DATABASE_URL -v user_id='your-user-id-here' -f seed_test_data.sql
-- ============================================================================

-- Check if user_id variable is set
DO $$
BEGIN
    IF current_setting('user_variables.user_id', true) IS NULL THEN
        RAISE EXCEPTION 'user_id variable not set. Use: psql -v user_id=''uuid'' -f seed_test_data.sql';
    END IF;
END $$;

-- Store user_id in a variable
\set user_id_var :user_id

-- ============================================================================
-- CREATE TEST WORKSPACE
-- ============================================================================

INSERT INTO workspaces (id, name, description, created_by, settings)
VALUES (
    'a0000000-0000-0000-0000-000000000001'::uuid,
    'Test Workspace',
    'Sample workspace for dashboard testing',
    :'user_id_var',
    '{}'::jsonb
)
ON CONFLICT (id) DO NOTHING;

-- Add user to workspace
INSERT INTO workspace_members (workspace_id, user_id, role, added_by)
VALUES (
    'a0000000-0000-0000-0000-000000000001'::uuid,
    :'user_id_var',
    'admin',
    :'user_id_var'
)
ON CONFLICT (workspace_id, user_id) DO NOTHING;

-- ============================================================================
-- CREATE TEST PROJECTS
-- ============================================================================

INSERT INTO projects (id, workspace_id, name, description, status, priority, created_by) VALUES
(
    'b0000000-0000-0000-0000-000000000001'::uuid,
    'a0000000-0000-0000-0000-000000000001'::uuid,
    'Website Redesign',
    'Complete overhaul of company website with modern design',
    'active',
    'high',
    :'user_id_var'
),
(
    'b0000000-0000-0000-0000-000000000002'::uuid,
    'a0000000-0000-0000-0000-000000000001'::uuid,
    'Mobile App Development',
    'Build native mobile apps for iOS and Android',
    'active',
    'high',
    :'user_id_var'
),
(
    'b0000000-0000-0000-0000-000000000003'::uuid,
    'a0000000-0000-0000-0000-000000000001'::uuid,
    'Marketing Campaign Q1',
    'Launch new product marketing campaign',
    'active',
    'medium',
    :'user_id_var'
),
(
    'b0000000-0000-0000-0000-000000000004'::uuid,
    'a0000000-0000-0000-0000-000000000001'::uuid,
    'Infrastructure Upgrade',
    'Migrate to cloud infrastructure',
    'planning',
    'medium',
    :'user_id_var'
),
(
    'b0000000-0000-0000-0000-000000000005'::uuid,
    'a0000000-0000-0000-0000-000000000001'::uuid,
    'Customer Portal',
    'Self-service portal for customer support',
    'completed',
    'low',
    :'user_id_var'
)
ON CONFLICT (id) DO NOTHING;

-- ============================================================================
-- CREATE TEST TASKS
-- ============================================================================

INSERT INTO tasks (id, workspace_id, project_id, title, description, status, priority, created_by, assigned_to, due_date) VALUES

-- Website Redesign tasks
(
    'c0000000-0000-0000-0000-000000000001'::uuid,
    'a0000000-0000-0000-0000-000000000001'::uuid,
    'b0000000-0000-0000-0000-000000000001'::uuid,
    'Design homepage mockups',
    'Create high-fidelity mockups for homepage',
    'in_progress',
    'high',
    :'user_id_var',
    :'user_id_var',
    CURRENT_DATE + INTERVAL '2 days'
),
(
    'c0000000-0000-0000-0000-000000000002'::uuid,
    'a0000000-0000-0000-0000-000000000001'::uuid,
    'b0000000-0000-0000-0000-000000000001'::uuid,
    'Implement responsive navigation',
    'Build mobile-friendly navigation component',
    'todo',
    'high',
    :'user_id_var',
    :'user_id_var',
    CURRENT_DATE + INTERVAL '5 days'
),
(
    'c0000000-0000-0000-0000-000000000003'::uuid,
    'a0000000-0000-0000-0000-000000000001'::uuid,
    'b0000000-0000-0000-0000-000000000001'::uuid,
    'Optimize images for web',
    'Compress and optimize all hero images',
    'completed',
    'medium',
    :'user_id_var',
    :'user_id_var',
    CURRENT_DATE - INTERVAL '2 days'
),

-- Mobile App Development tasks
(
    'c0000000-0000-0000-0000-000000000004'::uuid,
    'a0000000-0000-0000-0000-000000000001'::uuid,
    'b0000000-0000-0000-0000-000000000002'::uuid,
    'Set up React Native project',
    'Initialize project with required dependencies',
    'completed',
    'high',
    :'user_id_var',
    :'user_id_var',
    CURRENT_DATE - INTERVAL '10 days'
),
(
    'c0000000-0000-0000-0000-000000000005'::uuid,
    'a0000000-0000-0000-0000-000000000001'::uuid,
    'b0000000-0000-0000-0000-000000000002'::uuid,
    'Implement user authentication',
    'Add OAuth login with biometric support',
    'in_progress',
    'high',
    :'user_id_var',
    :'user_id_var',
    CURRENT_DATE + INTERVAL '7 days'
),
(
    'c0000000-0000-0000-0000-000000000006'::uuid,
    'a0000000-0000-0000-0000-000000000001'::uuid,
    'b0000000-0000-0000-0000-000000000002'::uuid,
    'Design app icon and splash screen',
    'Create branded assets for app stores',
    'todo',
    'medium',
    :'user_id_var',
    :'user_id_var',
    CURRENT_DATE + INTERVAL '14 days'
),

-- Marketing Campaign tasks
(
    'c0000000-0000-0000-0000-000000000007'::uuid,
    'a0000000-0000-0000-0000-000000000001'::uuid,
    'b0000000-0000-0000-0000-000000000003'::uuid,
    'Write product launch email',
    'Draft announcement email for existing customers',
    'todo',
    'high',
    :'user_id_var',
    :'user_id_var',
    CURRENT_DATE + INTERVAL '1 day'
),
(
    'c0000000-0000-0000-0000-000000000008'::uuid,
    'a0000000-0000-0000-0000-000000000001'::uuid,
    'b0000000-0000-0000-0000-000000000003'::uuid,
    'Create social media graphics',
    'Design 10 graphics for Twitter/LinkedIn campaign',
    'in_progress',
    'medium',
    :'user_id_var',
    :'user_id_var',
    CURRENT_DATE + INTERVAL '3 days'
),
(
    'c0000000-0000-0000-0000-000000000009'::uuid,
    'a0000000-0000-0000-0000-000000000001'::uuid,
    'b0000000-0000-0000-0000-000000000003'::uuid,
    'Schedule ad campaigns',
    'Set up Google/Facebook ads for product launch',
    'todo',
    'high',
    :'user_id_var',
    :'user_id_var',
    CURRENT_DATE + INTERVAL '6 days'
),

-- Infrastructure tasks
(
    'c0000000-0000-0000-0000-000000000010'::uuid,
    'a0000000-0000-0000-0000-000000000001'::uuid,
    'b0000000-0000-0000-0000-000000000004'::uuid,
    'Audit current infrastructure',
    'Document all services and dependencies',
    'completed',
    'high',
    :'user_id_var',
    :'user_id_var',
    CURRENT_DATE - INTERVAL '5 days'
),
(
    'c0000000-0000-0000-0000-000000000011'::uuid,
    'a0000000-0000-0000-0000-000000000001'::uuid,
    'b0000000-0000-0000-0000-000000000004'::uuid,
    'Set up Kubernetes cluster',
    'Deploy production-ready k8s cluster on GCP',
    'todo',
    'high',
    :'user_id_var',
    :'user_id_var',
    CURRENT_DATE + INTERVAL '10 days'
),

-- Overdue tasks (for testing)
(
    'c0000000-0000-0000-0000-000000000012'::uuid,
    'a0000000-0000-0000-0000-000000000001'::uuid,
    'b0000000-0000-0000-0000-000000000001'::uuid,
    'Fix critical security vulnerability',
    'Patch XSS vulnerability in contact form',
    'todo',
    'critical',
    :'user_id_var',
    :'user_id_var',
    CURRENT_DATE - INTERVAL '3 days'
),
(
    'c0000000-0000-0000-0000-000000000013'::uuid,
    'a0000000-0000-0000-0000-000000000001'::uuid,
    'b0000000-0000-0000-0000-000000000002'::uuid,
    'Submit app to app stores',
    'Upload builds to Apple/Google for review',
    'todo',
    'high',
    :'user_id_var',
    :'user_id_var',
    CURRENT_DATE - INTERVAL '1 day'
),

-- Tasks for today
(
    'c0000000-0000-0000-0000-000000000014'::uuid,
    'a0000000-0000-0000-0000-000000000001'::uuid,
    'b0000000-0000-0000-0000-000000000003'::uuid,
    'Review campaign analytics',
    'Analyze performance metrics from last week',
    'todo',
    'medium',
    :'user_id_var',
    :'user_id_var',
    CURRENT_DATE
),
(
    'c0000000-0000-0000-0000-000000000015'::uuid,
    'a0000000-0000-0000-0000-000000000001'::uuid,
    'b0000000-0000-0000-0000-000000000001'::uuid,
    'Client feedback meeting',
    'Present website mockups to stakeholders',
    'todo',
    'high',
    :'user_id_var',
    :'user_id_var',
    CURRENT_DATE
)

ON CONFLICT (id) DO NOTHING;

-- ============================================================================
-- CREATE TEST CLIENTS (for client widgets)
-- ============================================================================

INSERT INTO clients (id, workspace_id, name, email, company, status, created_by) VALUES
(
    'd0000000-0000-0000-0000-000000000001'::uuid,
    'a0000000-0000-0000-0000-000000000001'::uuid,
    'Acme Corp',
    'contact@acme.com',
    'Acme Corporation',
    'active',
    :'user_id_var'
),
(
    'd0000000-0000-0000-0000-000000000002'::uuid,
    'a0000000-0000-0000-0000-000000000001'::uuid,
    'TechStart Inc',
    'hello@techstart.io',
    'TechStart Incorporated',
    'active',
    :'user_id_var'
),
(
    'd0000000-0000-0000-0000-000000000003'::uuid,
    'a0000000-0000-0000-0000-000000000001'::uuid,
    'Global Solutions',
    'info@globalsolutions.net',
    'Global Solutions Ltd',
    'lead',
    :'user_id_var'
)
ON CONFLICT (id) DO NOTHING;

-- ============================================================================
-- VERIFICATION
-- ============================================================================

-- Show summary
SELECT 'Test Data Seeded Successfully' AS status;

SELECT
    'Summary' AS report,
    (SELECT COUNT(*) FROM projects WHERE created_by = :'user_id_var') AS projects_created,
    (SELECT COUNT(*) FROM tasks WHERE created_by = :'user_id_var') AS tasks_created,
    (SELECT COUNT(*) FROM tasks WHERE created_by = :'user_id_var' AND status = 'todo') AS tasks_todo,
    (SELECT COUNT(*) FROM tasks WHERE created_by = :'user_id_var' AND status = 'in_progress') AS tasks_in_progress,
    (SELECT COUNT(*) FROM tasks WHERE created_by = :'user_id_var' AND status = 'completed') AS tasks_completed,
    (SELECT COUNT(*) FROM tasks WHERE created_by = :'user_id_var' AND due_date < CURRENT_DATE) AS tasks_overdue,
    (SELECT COUNT(*) FROM tasks WHERE created_by = :'user_id_var' AND due_date = CURRENT_DATE) AS tasks_due_today,
    (SELECT COUNT(*) FROM clients WHERE created_by = :'user_id_var') AS clients_created;

-- Show project breakdown
SELECT
    'Projects' AS info,
    name,
    status,
    priority,
    (SELECT COUNT(*) FROM tasks WHERE project_id = projects.id) AS task_count
FROM projects
WHERE created_by = :'user_id_var'
ORDER BY created_at;

-- Show task breakdown by status
SELECT
    'Task Breakdown' AS info,
    status,
    COUNT(*) AS count,
    COUNT(*) FILTER (WHERE due_date < CURRENT_DATE) AS overdue,
    COUNT(*) FILTER (WHERE due_date = CURRENT_DATE) AS due_today,
    COUNT(*) FILTER (WHERE due_date > CURRENT_DATE) AS upcoming
FROM tasks
WHERE created_by = :'user_id_var'
GROUP BY status
ORDER BY
    CASE status
        WHEN 'todo' THEN 1
        WHEN 'in_progress' THEN 2
        WHEN 'completed' THEN 3
        ELSE 4
    END;
