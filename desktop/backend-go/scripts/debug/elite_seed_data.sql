-- ============================================================================
-- ELITE DATA SEEDING FOR AGENT SKILL TESTING
-- ============================================================================
-- This script creates comprehensive, realistic test data for testing ALL
-- BusinessOS agent skills including dashboards, tasks, projects, calendar,
-- clients, team members, documents, and analytics.
--
-- Usage:
--   psql $DATABASE_URL -v user_id='your-user-id-here' -f elite_seed_data.sql
--
-- OR via Node.js runner:
--   node run_elite_seed.js <user_id>
-- ============================================================================

\set user_id_var :user_id

BEGIN;

-- ============================================================================
-- 1. WORKSPACE SETUP
-- ============================================================================

INSERT INTO workspaces (id, name, slug, description, settings, owner_id, created_at)
VALUES (
    'e1000000-0000-4000-8000-000000000001'::uuid,
    'BusinessOS Elite Workspace',
    'elite-workspace',
    'Comprehensive workspace for agent skill testing with full data coverage',
    '{"theme": "light", "timezone": "America/New_York", "features": ["dashboards", "crm", "calendar"]}'::jsonb,
    :'user_id_var',
    NOW() - INTERVAL '90 days'
)
ON CONFLICT (id) DO NOTHING;

-- Add user to workspace
INSERT INTO workspace_members (workspace_id, user_id, role, status, invited_by, joined_at)
VALUES (
    'e1000000-0000-4000-8000-000000000001'::uuid,
    :'user_id_var',
    'admin',
    'active',
    :'user_id_var',
    NOW() - INTERVAL '90 days'
)
ON CONFLICT (workspace_id, user_id) DO NOTHING;

-- ============================================================================
-- 2. TEAM MEMBERS (3 members as per spec)
-- ============================================================================

INSERT INTO team_members (id, user_id, name, email, role, status, skills, hourly_rate, share_calendar, created_at) VALUES
(
    'f1000001-0000-4000-8000-000000000001'::uuid,
    :'user_id_var',
    'Javaris Tavel',
    'javaris@businessos.io',
    'Engineering Manager',
    'AVAILABLE',
    '["Leadership", "Go", "Architecture", "Project Management"]'::jsonb,
    150.00,
    true,
    NOW() - INTERVAL '180 days'
),
(
    'f1000002-0000-4000-8000-000000000002'::uuid,
    :'user_id_var',
    'Alex Kim',
    'alex.kim@businessos.io',
    'Senior Developer',
    'BUSY',
    '["TypeScript", "React", "Go", "PostgreSQL", "Docker"]'::jsonb,
    120.00,
    true,
    NOW() - INTERVAL '120 days'
),
(
    'f1000003-0000-4000-8000-000000000003'::uuid,
    :'user_id_var',
    'Morgan Lee',
    'morgan.lee@businessos.io',
    'Data Analyst',
    'AVAILABLE',
    '["SQL", "Python", "Data Visualization", "Analytics", "Reporting"]'::jsonb,
    100.00,
    true,
    NOW() - INTERVAL '90 days'
)
ON CONFLICT (id) DO NOTHING;

-- ============================================================================
-- 3. CLIENTS (Acme Corp, Globex Inc, and more)
-- ============================================================================

INSERT INTO clients (id, user_id, name, type, email, phone, website, industry, status, notes, created_at) VALUES
(
    'c1000001-0000-4000-8000-000000000001'::uuid,
    :'user_id_var',
    'Acme Corporation',
    'company',
    'contact@acme-corp.example.com',
    '+1-555-100-1001',
    'https://acme-corp.example.com',
    'Technology',
    'active',
    'Major enterprise client, 500+ employees. Primary contact is their CTO. Annual contract worth $250k.',
    NOW() - INTERVAL '180 days'
),
(
    'c1000002-0000-4000-8000-000000000002'::uuid,
    :'user_id_var',
    'Globex Inc',
    'company',
    'partnerships@globex.example.com',
    '+1-555-200-2002',
    'https://globex.example.com',
    'Manufacturing',
    'active',
    'Growing manufacturing company. Interested in automation and analytics features.',
    NOW() - INTERVAL '90 days'
),
(
    'c1000003-0000-4000-8000-000000000003'::uuid,
    :'user_id_var',
    'StartupXYZ',
    'company',
    'hello@startupxyz.example.com',
    '+1-555-300-3003',
    'https://startupxyz.example.com',
    'SaaS',
    'lead',
    'Early-stage startup evaluating our platform. Good potential for growth.',
    NOW() - INTERVAL '14 days'
),
(
    'c1000004-0000-4000-8000-000000000004'::uuid,
    :'user_id_var',
    'RetailMax',
    'company',
    'tech@retailmax.example.com',
    '+1-555-400-4004',
    'https://retailmax.example.com',
    'Retail',
    'prospect',
    'Large retail chain looking for inventory management integration.',
    NOW() - INTERVAL '30 days'
)
ON CONFLICT (id) DO NOTHING;

-- Client Contacts
INSERT INTO client_contacts (id, client_id, name, email, phone, role, is_primary, created_at) VALUES
(gen_random_uuid(), 'c1000001-0000-4000-8000-000000000001'::uuid, 'John Mitchell', 'john.mitchell@acme-corp.example.com', '+1-555-100-1010', 'CTO', true, NOW()),
(gen_random_uuid(), 'c1000001-0000-4000-8000-000000000001'::uuid, 'Sarah Chen', 'sarah.chen@acme-corp.example.com', '+1-555-100-1011', 'VP of Engineering', false, NOW()),
(gen_random_uuid(), 'c1000002-0000-4000-8000-000000000002'::uuid, 'Mike Thompson', 'mike.thompson@globex.example.com', '+1-555-200-2010', 'Operations Director', true, NOW()),
(gen_random_uuid(), 'c1000003-0000-4000-8000-000000000003'::uuid, 'Emma Davis', 'emma@startupxyz.example.com', '+1-555-300-3010', 'Founder & CEO', true, NOW())
ON CONFLICT DO NOTHING;

-- DROPPED in migration 099: client_deals seed data removed

-- ============================================================================
-- 4. PROJECTS (Website Redesign, Q2 Launch, Marketing Campaign + more)
-- ============================================================================

INSERT INTO projects (id, user_id, workspace_id, name, description, status, priority, client_id, project_type, start_date, due_date, created_at) VALUES
(
    'a1000001-0000-4000-8000-000000000001'::uuid,
    :'user_id_var',
    'e1000000-0000-4000-8000-000000000001'::uuid,
    'Website Redesign',
    'Complete overhaul of the company website with modern UI/UX, improved performance, and mobile-first design.',
    'ACTIVE',
    'HIGH',
    NULL,
    'internal',
    CURRENT_DATE - INTERVAL '30 days',
    CURRENT_DATE + INTERVAL '30 days',
    NOW() - INTERVAL '30 days'
),
(
    'a1000002-0000-4000-8000-000000000002'::uuid,
    :'user_id_var',
    'e1000000-0000-4000-8000-000000000001'::uuid,
    'Q2 Product Launch',
    'Launch of new enterprise features including advanced analytics, team collaboration tools, and API v2.',
    'ACTIVE',
    'CRITICAL',
    NULL,
    'internal',
    CURRENT_DATE - INTERVAL '14 days',
    CURRENT_DATE + INTERVAL '60 days',
    NOW() - INTERVAL '14 days'
),
(
    'a1000003-0000-4000-8000-000000000003'::uuid,
    :'user_id_var',
    'e1000000-0000-4000-8000-000000000001'::uuid,
    'Marketing Campaign',
    'Multi-channel marketing campaign to promote Q2 launch including email, social media, webinars, and content marketing.',
    'ACTIVE',
    'HIGH',
    NULL,
    'internal',
    CURRENT_DATE - INTERVAL '7 days',
    CURRENT_DATE + INTERVAL '45 days',
    NOW() - INTERVAL '7 days'
),
(
    'a1000004-0000-4000-8000-000000000004'::uuid,
    :'user_id_var',
    'e1000000-0000-4000-8000-000000000001'::uuid,
    'Acme Corp Integration',
    'Custom integration project for Acme Corporation - Salesforce sync, custom reports, and API development.',
    'ACTIVE',
    'HIGH',
    'c1000001-0000-4000-8000-000000000001'::uuid,
    'client',
    CURRENT_DATE - INTERVAL '45 days',
    CURRENT_DATE + INTERVAL '15 days',
    NOW() - INTERVAL '45 days'
),
(
    'a1000005-0000-4000-8000-000000000005'::uuid,
    :'user_id_var',
    'e1000000-0000-4000-8000-000000000001'::uuid,
    'Globex Analytics Dashboard',
    'Build custom analytics dashboard for Globex manufacturing metrics and KPIs.',
    'ACTIVE',
    'MEDIUM',
    'c1000002-0000-4000-8000-000000000002'::uuid,
    'client',
    CURRENT_DATE - INTERVAL '21 days',
    CURRENT_DATE + INTERVAL '45 days',
    NOW() - INTERVAL '21 days'
),
(
    'a1000006-0000-4000-8000-000000000006'::uuid,
    :'user_id_var',
    'e1000000-0000-4000-8000-000000000001'::uuid,
    'Infrastructure Upgrade',
    'Migrate to Kubernetes, implement CI/CD improvements, and optimize database performance.',
    'PAUSED',
    'MEDIUM',
    NULL,
    'internal',
    CURRENT_DATE - INTERVAL '60 days',
    NULL,
    NOW() - INTERVAL '60 days'
),
(
    'a1000007-0000-4000-8000-000000000007'::uuid,
    :'user_id_var',
    'e1000000-0000-4000-8000-000000000001'::uuid,
    'Customer Portal MVP',
    'Self-service customer portal for support tickets, documentation, and account management.',
    'COMPLETED',
    'HIGH',
    NULL,
    'internal',
    CURRENT_DATE - INTERVAL '90 days',
    CURRENT_DATE - INTERVAL '14 days',
    NOW() - INTERVAL '90 days'
)
ON CONFLICT (id) DO NOTHING;

-- ============================================================================
-- 5. TASKS (Varied priorities, statuses, due dates)
-- ============================================================================

INSERT INTO tasks (id, user_id, project_id, title, description, status, priority, assignee_id, due_date, created_at) VALUES

-- Website Redesign Tasks (assigned to Javaris - team member f1000001)
('b1000000-0000-4000-8000-000000000001'::uuid, :'user_id_var', 'a1000001-0000-4000-8000-000000000001'::uuid,
 'Design homepage mockups', 'Create high-fidelity mockups for the new homepage including hero section, features, and testimonials.',
 'in_progress', 'high', 'f1000001-0000-4000-8000-000000000001'::uuid, CURRENT_DATE + INTERVAL '2 days', NOW() - INTERVAL '5 days'),

('b1000000-0000-4000-8000-000000000002'::uuid, :'user_id_var', 'a1000001-0000-4000-8000-000000000001'::uuid,
 'Implement responsive navigation', 'Build mobile-first navigation component with hamburger menu and smooth animations.',
 'todo', 'high', 'f1000002-0000-4000-8000-000000000002'::uuid, CURRENT_DATE + INTERVAL '5 days', NOW() - INTERVAL '3 days'),

('b1000000-0000-4000-8000-000000000003'::uuid, :'user_id_var', 'a1000001-0000-4000-8000-000000000001'::uuid,
 'Optimize images for web', 'Compress and convert all images to WebP format, implement lazy loading.',
 'done', 'medium', 'f1000002-0000-4000-8000-000000000002'::uuid, CURRENT_DATE - INTERVAL '3 days', NOW() - INTERVAL '10 days'),

('b1000000-0000-4000-8000-000000000004'::uuid, :'user_id_var', 'a1000001-0000-4000-8000-000000000001'::uuid,
 'Set up CI/CD for preview deployments', 'Configure GitHub Actions for automatic preview deployments on PRs.',
 'done', 'medium', 'f1000002-0000-4000-8000-000000000002'::uuid, CURRENT_DATE - INTERVAL '7 days', NOW() - INTERVAL '14 days'),

-- Q2 Product Launch Tasks (assigned to Alex - team member f1000002)
('b1000000-0000-4000-8000-000000000005'::uuid, :'user_id_var', 'a1000002-0000-4000-8000-000000000002'::uuid,
 'Review Q1 analytics report', 'Analyze Q1 performance metrics and identify areas for improvement.',
 'todo', 'high', 'f1000003-0000-4000-8000-000000000003'::uuid, CURRENT_DATE, NOW() - INTERVAL '2 days'),

('b1000000-0000-4000-8000-000000000006'::uuid, :'user_id_var', 'a1000002-0000-4000-8000-000000000002'::uuid,
 'Write API v2 documentation', 'Complete OpenAPI documentation for all new endpoints with examples.',
 'in_progress', 'critical', 'f1000001-0000-4000-8000-000000000001'::uuid, CURRENT_DATE + INTERVAL '7 days', NOW() - INTERVAL '5 days'),

('b1000000-0000-4000-8000-000000000007'::uuid, :'user_id_var', 'a1000002-0000-4000-8000-000000000002'::uuid,
 'Test team collaboration features', 'End-to-end testing of real-time collaboration, permissions, and sharing.',
 'todo', 'high', 'f1000002-0000-4000-8000-000000000002'::uuid, CURRENT_DATE + INTERVAL '10 days', NOW() - INTERVAL '3 days'),

('b1000000-0000-4000-8000-000000000008'::uuid, :'user_id_var', 'a1000002-0000-4000-8000-000000000002'::uuid,
 'Deploy staging environment', 'Set up staging environment with production-like data for QA testing.',
 'done', 'critical', 'f1000002-0000-4000-8000-000000000002'::uuid, CURRENT_DATE - INTERVAL '5 days', NOW() - INTERVAL '10 days'),

-- Marketing Campaign Tasks (assigned to Morgan - team member f1000003)
('b1000000-0000-4000-8000-000000000009'::uuid, :'user_id_var', 'a1000003-0000-4000-8000-000000000003'::uuid,
 'Write product launch email sequence', 'Create 5-email drip campaign: teaser, announcement, features, testimonials, follow-up.',
 'in_progress', 'high', 'f1000001-0000-4000-8000-000000000001'::uuid, CURRENT_DATE + INTERVAL '3 days', NOW() - INTERVAL '4 days'),

('b1000000-0000-4000-8000-000000000010'::uuid, :'user_id_var', 'a1000003-0000-4000-8000-000000000003'::uuid,
 'Create social media graphics', 'Design graphics for Twitter, LinkedIn, and Instagram campaign.',
 'todo', 'medium', 'f1000003-0000-4000-8000-000000000003'::uuid, CURRENT_DATE + INTERVAL '5 days', NOW() - INTERVAL '2 days'),

('b1000000-0000-4000-8000-000000000011'::uuid, :'user_id_var', 'a1000003-0000-4000-8000-000000000003'::uuid,
 'Schedule webinar series', 'Plan and schedule 3 webinars: Product Demo, Customer Success Stories, Technical Deep Dive.',
 'todo', 'high', 'f1000001-0000-4000-8000-000000000001'::uuid, CURRENT_DATE + INTERVAL '14 days', NOW() - INTERVAL '1 day'),

-- Acme Corp Integration Tasks
('b1000000-0000-4000-8000-000000000012'::uuid, :'user_id_var', 'a1000004-0000-4000-8000-000000000004'::uuid,
 'Build Salesforce connector', 'Develop bidirectional sync with Salesforce CRM for contacts and deals.',
 'in_progress', 'critical', 'f1000002-0000-4000-8000-000000000002'::uuid, CURRENT_DATE + INTERVAL '5 days', NOW() - INTERVAL '14 days'),

('b1000000-0000-4000-8000-000000000013'::uuid, :'user_id_var', 'a1000004-0000-4000-8000-000000000004'::uuid,
 'Create custom reports for Acme', 'Build 5 custom reports: Sales Pipeline, Team Performance, Revenue Forecast, Churn Risk, Usage Analytics.',
 'todo', 'high', 'f1000003-0000-4000-8000-000000000003'::uuid, CURRENT_DATE + INTERVAL '10 days', NOW() - INTERVAL '7 days'),

-- OVERDUE Tasks (Critical for testing)
('b1000000-0000-4000-8000-000000000014'::uuid, :'user_id_var', 'a1000001-0000-4000-8000-000000000001'::uuid,
 'Fix critical security vulnerability', 'Patch XSS vulnerability in user input fields - URGENT.',
 'todo', 'critical', 'f1000002-0000-4000-8000-000000000002'::uuid, CURRENT_DATE - INTERVAL '2 days', NOW() - INTERVAL '5 days'),

('b1000000-0000-4000-8000-000000000015'::uuid, :'user_id_var', 'a1000002-0000-4000-8000-000000000002'::uuid,
 'Submit compliance documentation', 'Complete SOC2 compliance documentation for enterprise clients.',
 'todo', 'high', 'f1000001-0000-4000-8000-000000000001'::uuid, CURRENT_DATE - INTERVAL '1 day', NOW() - INTERVAL '10 days'),

-- Tasks due TODAY (Critical for todays_focus testing)
('b1000000-0000-4000-8000-000000000016'::uuid, :'user_id_var', 'a1000003-0000-4000-8000-000000000003'::uuid,
 'Review campaign analytics dashboard', 'Analyze last weeks campaign performance and prepare report.',
 'todo', 'medium', 'f1000003-0000-4000-8000-000000000003'::uuid, CURRENT_DATE, NOW() - INTERVAL '3 days'),

('b1000000-0000-4000-8000-000000000017'::uuid, :'user_id_var', 'a1000004-0000-4000-8000-000000000004'::uuid,
 'Client feedback meeting prep', 'Prepare presentation materials for Acme Corp quarterly review.',
 'todo', 'high', 'f1000001-0000-4000-8000-000000000001'::uuid, CURRENT_DATE, NOW() - INTERVAL '2 days'),

-- Low priority tasks for variety
('b1000000-0000-4000-8000-000000000018'::uuid, :'user_id_var', 'a1000006-0000-4000-8000-000000000006'::uuid,
 'Research Kubernetes alternatives', 'Evaluate Nomad, Docker Swarm, and ECS for container orchestration.',
 'todo', 'low', 'f1000002-0000-4000-8000-000000000002'::uuid, CURRENT_DATE + INTERVAL '30 days', NOW() - INTERVAL '14 days'),

('b1000000-0000-4000-8000-000000000019'::uuid, :'user_id_var', NULL,
 'Update team wiki documentation', 'Refresh onboarding docs and add new process guidelines.',
 'todo', 'low', NULL, CURRENT_DATE + INTERVAL '21 days', NOW() - INTERVAL '7 days'),

('b1000000-0000-4000-8000-000000000020'::uuid, :'user_id_var', 'a1000005-0000-4000-8000-000000000005'::uuid,
 'Design Globex dashboard wireframes', 'Create low-fidelity wireframes for manufacturing KPI dashboard.',
 'in_progress', 'medium', 'f1000001-0000-4000-8000-000000000001'::uuid, CURRENT_DATE + INTERVAL '7 days', NOW() - INTERVAL '5 days')

ON CONFLICT (id) DO NOTHING;

-- ============================================================================
-- 6. FOCUS ITEMS (for todays_focus widget)
-- ============================================================================

INSERT INTO focus_items (id, user_id, text, completed, focus_date, created_at) VALUES
(gen_random_uuid(), :'user_id_var', 'Complete API documentation review', false, CURRENT_DATE, NOW()),
(gen_random_uuid(), :'user_id_var', 'Prepare for client meeting at 2pm', false, CURRENT_DATE, NOW()),
(gen_random_uuid(), :'user_id_var', 'Fix critical security issue', false, CURRENT_DATE, NOW()),
(gen_random_uuid(), :'user_id_var', 'Review PR for navigation component', true, CURRENT_DATE, NOW() - INTERVAL '2 hours'),
(gen_random_uuid(), :'user_id_var', 'Send weekly status update', false, CURRENT_DATE, NOW()),
(gen_random_uuid(), :'user_id_var', 'Finalized homepage mockups', true, CURRENT_DATE - INTERVAL '1 day', NOW() - INTERVAL '1 day'),
(gen_random_uuid(), :'user_id_var', 'Deployed staging environment', true, CURRENT_DATE - INTERVAL '2 days', NOW() - INTERVAL '2 days')
ON CONFLICT DO NOTHING;

-- ============================================================================
-- 7. CALENDAR EVENTS (meetings, deadlines, reviews)
-- ============================================================================

INSERT INTO calendar_events (id, user_id, title, description, start_time, end_time, meeting_type, project_id, client_id, location, source, created_at) VALUES
-- TODAY's events
(gen_random_uuid(), :'user_id_var', 'Daily Standup', 'Team sync - what are we working on today?',
 CURRENT_DATE + TIME '09:00', CURRENT_DATE + TIME '09:30', 'standup', NULL, NULL, 'Zoom', 'businessos', NOW()),

(gen_random_uuid(), :'user_id_var', 'Acme Corp Quarterly Review', 'Q1 review and Q2 planning with Acme team',
 CURRENT_DATE + TIME '14:00', CURRENT_DATE + TIME '15:30', 'client', 'a1000004-0000-4000-8000-000000000004'::uuid, 'c1000001-0000-4000-8000-000000000001'::uuid, 'Conference Room A', 'businessos', NOW()),

(gen_random_uuid(), :'user_id_var', '1:1 with Alex Kim', 'Weekly sync on development progress',
 CURRENT_DATE + TIME '16:00', CURRENT_DATE + TIME '16:30', 'one_on_one', NULL, NULL, 'Zoom', 'businessos', NOW()),

-- THIS WEEK events
(gen_random_uuid(), :'user_id_var', 'Sprint Planning', 'Plan Q2 sprint 3 - feature prioritization',
 CURRENT_DATE + INTERVAL '1 day' + TIME '10:00', CURRENT_DATE + INTERVAL '1 day' + TIME '11:30', 'planning', 'a1000002-0000-4000-8000-000000000002'::uuid, NULL, 'Main Office', 'businessos', NOW()),

(gen_random_uuid(), :'user_id_var', 'Website Redesign Review', 'Review homepage mockups with stakeholders',
 CURRENT_DATE + INTERVAL '2 days' + TIME '11:00', CURRENT_DATE + INTERVAL '2 days' + TIME '12:00', 'review', 'a1000001-0000-4000-8000-000000000001'::uuid, NULL, 'Design Room', 'businessos', NOW()),

(gen_random_uuid(), :'user_id_var', 'Globex Discovery Call', 'Initial technical requirements gathering',
 CURRENT_DATE + INTERVAL '2 days' + TIME '14:00', CURRENT_DATE + INTERVAL '2 days' + TIME '15:00', 'sales', 'a1000005-0000-4000-8000-000000000005'::uuid, 'c1000002-0000-4000-8000-000000000002'::uuid, 'Zoom', 'businessos', NOW()),

(gen_random_uuid(), :'user_id_var', 'Marketing Strategy Sync', 'Align on Q2 launch campaign messaging',
 CURRENT_DATE + INTERVAL '3 days' + TIME '13:00', CURRENT_DATE + INTERVAL '3 days' + TIME '14:00', 'team', 'a1000003-0000-4000-8000-000000000003'::uuid, NULL, 'Marketing Room', 'businessos', NOW()),

(gen_random_uuid(), :'user_id_var', 'Team Retrospective', 'Sprint 2 retrospective - what went well, what to improve',
 CURRENT_DATE + INTERVAL '4 days' + TIME '15:00', CURRENT_DATE + INTERVAL '4 days' + TIME '16:30', 'retrospective', NULL, NULL, 'All Hands Room', 'businessos', NOW()),

-- NEXT WEEK events
(gen_random_uuid(), :'user_id_var', 'Product Demo Webinar', 'Public webinar showcasing new features',
 CURRENT_DATE + INTERVAL '7 days' + TIME '11:00', CURRENT_DATE + INTERVAL '7 days' + TIME '12:00', 'external', 'a1000002-0000-4000-8000-000000000002'::uuid, NULL, 'Webinar Platform', 'businessos', NOW()),

(gen_random_uuid(), :'user_id_var', 'StartupXYZ Demo', 'Product demo for potential new client',
 CURRENT_DATE + INTERVAL '8 days' + TIME '10:00', CURRENT_DATE + INTERVAL '8 days' + TIME '11:00', 'sales', NULL, 'c1000003-0000-4000-8000-000000000003'::uuid, 'Zoom', 'businessos', NOW()),

-- PAST events (for history)
(gen_random_uuid(), :'user_id_var', 'Q1 All Hands Meeting', 'Company-wide Q1 review and Q2 outlook',
 CURRENT_DATE - INTERVAL '7 days' + TIME '10:00', CURRENT_DATE - INTERVAL '7 days' + TIME '12:00', 'team', NULL, NULL, 'Main Auditorium', 'businessos', NOW() - INTERVAL '7 days'),

(gen_random_uuid(), :'user_id_var', 'Security Audit Review', 'Review findings from external security audit',
 CURRENT_DATE - INTERVAL '3 days' + TIME '14:00', CURRENT_DATE - INTERVAL '3 days' + TIME '15:00', 'review', NULL, NULL, 'Secure Room', 'businessos', NOW() - INTERVAL '3 days')

ON CONFLICT DO NOTHING;

-- ============================================================================
-- 8. KNOWLEDGE BASE / DOCUMENTS
-- ============================================================================

INSERT INTO uploaded_documents (id, user_id, filename, original_filename, display_name, description, file_type, mime_type, file_size_bytes, storage_path, storage_provider, extracted_text, document_type, category, processing_status, created_at) VALUES
(
    'd0c10000-0000-4000-8000-000000000001'::uuid,
    :'user_id_var',
    'meeting-notes-q1-review.md',
    'Q1 Review Meeting Notes.md',
    'Q1 Review Meeting Notes',
    'Notes from Q1 review meeting including key decisions, action items, and next steps.',
    'markdown',
    'text/markdown',
    4500,
    '/documents/meeting-notes-q1-review.md',
    'local',
    '# Q1 Review Meeting Notes

## Date: March 15, 2026

### Attendees
- Javaris Tavel (Engineering Manager)
- Alex Kim (Senior Developer)
- Morgan Lee (Data Analyst)
- Sarah Chen (Product Manager)

### Key Accomplishments
1. Launched Customer Portal MVP - 2 weeks ahead of schedule
2. Achieved 99.9% uptime for production systems
3. Reduced API latency by 40% through optimization
4. Onboarded 15 new enterprise customers

### Challenges Faced
- Security vulnerability required emergency patch
- Scaling issues with real-time collaboration feature
- Integration delays with Salesforce connector

### Action Items
1. [Javaris] Complete SOC2 documentation by end of month
2. [Alex] Implement rate limiting for API v2
3. [Morgan] Create Q1 analytics dashboard for board meeting

### Next Quarter Focus
- Q2 Product Launch with enterprise features
- Website redesign completion
- Marketing campaign execution',
    'reference',
    'meetings',
    'completed',
    NOW() - INTERVAL '30 days'
),
(
    'd0c10000-0000-4000-8000-000000000002'::uuid,
    :'user_id_var',
    'marketing-strategy-q2.md',
    'Marketing Strategy Q2 2026.md',
    'Q2 Marketing Strategy',
    'Comprehensive marketing strategy document for Q2 2026 product launch campaign.',
    'markdown',
    'text/markdown',
    8200,
    '/documents/marketing-strategy-q2.md',
    'local',
    '# Q2 2026 Marketing Strategy

## Campaign Overview
Launch campaign for new enterprise features targeting mid-to-large B2B companies.

## Target Audience
- Decision-makers: CTOs, VPs of Engineering, Product Leaders
- Company size: 100-5000 employees
- Industries: SaaS, FinTech, E-commerce, Manufacturing

## Campaign Channels

### Email Marketing
- 5-email drip sequence
- Segmented by company size and industry
- A/B testing on subject lines

### Social Media
- LinkedIn thought leadership posts (3x/week)
- Twitter product announcements
- YouTube demo videos

### Content Marketing
- 4 blog posts on new features
- 2 customer case studies
- Technical whitepaper on architecture

### Webinars
1. Product Demo Webinar - Week 1
2. Customer Success Stories - Week 2
3. Technical Deep Dive - Week 3

## Budget Allocation
- Paid Ads: $15,000
- Content Creation: $8,000
- Events/Webinars: $5,000
- Tools/Software: $2,000
Total: $30,000

## Success Metrics
- 500 webinar registrations
- 50 demo requests
- 10 new enterprise trials
- 25% email open rate',
    'sop',
    'marketing',
    'completed',
    NOW() - INTERVAL '14 days'
),
(
    'd0c10000-0000-4000-8000-000000000003'::uuid,
    :'user_id_var',
    'client-onboarding-guide.md',
    'Client Onboarding Guide.md',
    'Client Onboarding Guide',
    'Step-by-step guide for onboarding new enterprise clients including setup, training, and success milestones.',
    'markdown',
    'text/markdown',
    6800,
    '/documents/client-onboarding-guide.md',
    'local',
    '# Client Onboarding Guide

## Overview
This guide outlines the process for onboarding new enterprise clients to ensure a smooth transition and successful adoption.

## Pre-Onboarding (Week -1)
1. Send welcome email with account credentials
2. Schedule kickoff meeting
3. Prepare customized onboarding plan
4. Set up dedicated Slack channel

## Week 1: Foundation
- [ ] Kickoff meeting with stakeholders
- [ ] Account configuration and SSO setup
- [ ] Import existing data (contacts, projects)
- [ ] Set up initial integrations

## Week 2: Training
- [ ] Admin training session (2 hours)
- [ ] End-user training webinar (1 hour)
- [ ] Provide training documentation
- [ ] Q&A office hours

## Week 3: Optimization
- [ ] Review usage analytics
- [ ] Custom workflow configuration
- [ ] Advanced feature training
- [ ] Integration fine-tuning

## Week 4: Success Review
- [ ] 30-day check-in meeting
- [ ] Gather feedback and testimonial
- [ ] Identify expansion opportunities
- [ ] Transition to Customer Success

## Success Milestones
1. 80% of team active in first week
2. Core workflows configured by week 2
3. At least one integration connected
4. NPS score of 8+ at 30 days',
    'template',
    'operations',
    'completed',
    NOW() - INTERVAL '60 days'
)
ON CONFLICT (id) DO NOTHING;

-- ============================================================================
-- 9. DASHBOARDS (Sales Overview, Productivity Dashboard)
-- ============================================================================

INSERT INTO user_dashboards (id, user_id, workspace_id, name, description, is_default, layout, visibility, created_via, created_at) VALUES
(
    'da510000-0000-4000-8000-000000000001'::uuid,
    :'user_id_var',
    'e1000000-0000-4000-8000-000000000001'::uuid,
    'My Day',
    'Daily productivity dashboard with tasks, focus items, and calendar',
    true,
    '[
        {"widget_id": "w1", "widget_type": "todays_focus", "position": {"x": 0, "y": 0, "w": 4, "h": 3}, "config": {"max_items": 5, "show_completed": true}},
        {"widget_id": "w2", "widget_type": "task_summary", "position": {"x": 4, "y": 0, "w": 4, "h": 3}, "config": {"group_by": "status", "show_completed": false}},
        {"widget_id": "w3", "widget_type": "upcoming_deadlines", "position": {"x": 8, "y": 0, "w": 4, "h": 3}, "config": {"days_ahead": 7, "show_overdue": true}},
        {"widget_id": "w4", "widget_type": "recent_activity", "position": {"x": 0, "y": 3, "w": 6, "h": 3}, "config": {"limit": 10}},
        {"widget_id": "w5", "widget_type": "quick_actions", "position": {"x": 6, "y": 3, "w": 3, "h": 2}, "config": {}},
        {"widget_id": "w6", "widget_type": "metric_card", "position": {"x": 9, "y": 3, "w": 3, "h": 2}, "config": {"metric": "tasks_due_today"}}
    ]'::jsonb,
    'private',
    'agent',
    NOW() - INTERVAL '7 days'
),
(
    'da510000-0000-4000-8000-000000000002'::uuid,
    :'user_id_var',
    'e1000000-0000-4000-8000-000000000001'::uuid,
    'Sales Overview',
    'CRM dashboard with client overview, pipeline, and revenue metrics',
    false,
    '[
        {"widget_id": "w1", "widget_type": "metric_card", "position": {"x": 0, "y": 0, "w": 3, "h": 2}, "config": {"metric": "active_projects"}},
        {"widget_id": "w2", "widget_type": "metric_card", "position": {"x": 3, "y": 0, "w": 3, "h": 2}, "config": {"metric": "tasks_completed_week"}},
        {"widget_id": "w3", "widget_type": "metric_card", "position": {"x": 6, "y": 0, "w": 3, "h": 2}, "config": {"metric": "tasks_overdue"}},
        {"widget_id": "w4", "widget_type": "metric_card", "position": {"x": 9, "y": 0, "w": 3, "h": 2}, "config": {"metric": "tasks_due_today"}},
        {"widget_id": "w5", "widget_type": "client_overview", "position": {"x": 0, "y": 2, "w": 6, "h": 4}, "config": {"limit": 5, "sort_by": "recent_activity"}},
        {"widget_id": "w6", "widget_type": "project_progress", "position": {"x": 6, "y": 2, "w": 6, "h": 4}, "config": {"show_all": true}}
    ]'::jsonb,
    'private',
    'agent',
    NOW() - INTERVAL '5 days'
),
(
    'da510000-0000-4000-8000-000000000003'::uuid,
    :'user_id_var',
    'e1000000-0000-4000-8000-000000000001'::uuid,
    'Productivity Dashboard',
    'Analytics dashboard with burndown charts, heatmaps, and task metrics',
    false,
    '[
        {"widget_id": "w1", "widget_type": "task_burndown", "position": {"x": 0, "y": 0, "w": 6, "h": 4}, "config": {"days": 30}},
        {"widget_id": "w2", "widget_type": "workload_heatmap", "position": {"x": 6, "y": 0, "w": 6, "h": 3}, "config": {"date_range": "month"}},
        {"widget_id": "w3", "widget_type": "task_list", "position": {"x": 0, "y": 4, "w": 6, "h": 4}, "config": {"filter": {"assigned_to_me": true}, "sort": "priority", "limit": 15}},
        {"widget_id": "w4", "widget_type": "notes_pinned", "position": {"x": 6, "y": 3, "w": 6, "h": 3}, "config": {"limit": 5}},
        {"widget_id": "w5", "widget_type": "agent_shortcuts", "position": {"x": 6, "y": 6, "w": 6, "h": 2}, "config": {}}
    ]'::jsonb,
    'private',
    'manual',
    NOW() - INTERVAL '3 days'
)
ON CONFLICT (id) DO NOTHING;

-- ============================================================================
-- 10. CONTEXTS (for document search and AI context)
-- ============================================================================

INSERT INTO contexts (id, user_id, name, type, content, created_at) VALUES
(gen_random_uuid(), :'user_id_var', 'Company Brand Guidelines', 'DOCUMENT',
'# Brand Guidelines

## Voice & Tone
- Professional but approachable
- Confident without being arrogant
- Educational and helpful
- Clear and concise

## Visual Identity
- Primary: Blue (#2563EB)
- Secondary: Purple (#8B5CF6)
- Fonts: Inter (UI), Newsreader (Marketing)

## Messaging Pillars
1. Simplicity: Complex problems, simple solutions
2. Speed: Fast implementation, quick results
3. Scale: Grows with your business',
NOW() - INTERVAL '90 days'),

(gen_random_uuid(), :'user_id_var', 'Target Customer Profile', 'BUSINESS',
'# Ideal Customer Profile (ICP)

## Demographics
- Company Size: 100-1000 employees
- Revenue: $10M-$100M ARR
- Industry: B2B SaaS, FinTech, Manufacturing
- Team: Product/Engineering focused

## Pain Points
- Manual workflows slowing down teams
- Data scattered across multiple tools
- Lack of visibility into project status
- Difficulty scaling operations

## Buying Criteria
- ROI within 6 months
- Easy integration with existing stack
- Strong customer support
- Compliance and security features',
NOW() - INTERVAL '60 days')
ON CONFLICT DO NOTHING;

COMMIT;

-- ============================================================================
-- VERIFICATION QUERIES
-- ============================================================================

SELECT '✅ Elite Data Seeding Complete!' AS status;

SELECT
    'Summary' AS report,
    (SELECT COUNT(*) FROM projects WHERE user_id = :'user_id_var') AS projects,
    (SELECT COUNT(*) FROM tasks WHERE user_id = :'user_id_var') AS tasks,
    (SELECT COUNT(*) FROM clients WHERE user_id = :'user_id_var') AS clients,
    (SELECT COUNT(*) FROM team_members WHERE user_id = :'user_id_var') AS team_members,
    (SELECT COUNT(*) FROM calendar_events WHERE user_id = :'user_id_var') AS calendar_events,
    (SELECT COUNT(*) FROM uploaded_documents WHERE user_id = :'user_id_var') AS documents,
    (SELECT COUNT(*) FROM user_dashboards WHERE user_id = :'user_id_var') AS dashboards,
    (SELECT COUNT(*) FROM focus_items WHERE user_id = :'user_id_var') AS focus_items;

-- Task breakdown by status
SELECT
    'Task Status Breakdown' AS info,
    status,
    COUNT(*) AS count,
    COUNT(*) FILTER (WHERE due_date < CURRENT_DATE AND status != 'done') AS overdue,
    COUNT(*) FILTER (WHERE due_date = CURRENT_DATE) AS due_today
FROM tasks
WHERE user_id = :'user_id_var'
GROUP BY status;

-- Project summary
SELECT
    'Project Summary' AS info,
    name,
    status,
    priority,
    (SELECT COUNT(*) FROM tasks WHERE project_id = projects.id) AS task_count
FROM projects
WHERE user_id = :'user_id_var'
ORDER BY created_at;

SELECT '🎯 Ready for Agent Skill Testing!' AS message;
