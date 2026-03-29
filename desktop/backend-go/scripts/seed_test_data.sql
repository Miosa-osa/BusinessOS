-- ============================================
-- SEED TEST DATA FOR UI/UX TESTING
-- Run: psql $DATABASE_URL -f seed_test_data.sql
-- Remove: psql $DATABASE_URL -f unseed_test_data.sql
-- ============================================

-- Use a test user ID (change this to your actual user ID if needed)
DO $$
DECLARE
    test_user_id VARCHAR(255) := 'test-user-seed-001';

    -- Team member IDs
    tm_sarah UUID;
    tm_mike UUID;
    tm_lisa UUID;
    tm_john UUID;
    tm_emily UUID;

    -- Client IDs
    cl_techcorp UUID;
    cl_greenleaf UUID;
    cl_urbandev UUID;
    cl_skyline UUID;
    cl_brightpath UUID;
    cl_nexus UUID;
    cl_harmony UUID;
    cl_pinnacle UUID;
    cl_elevate UUID;
    cl_quantum UUID;
    cl_stellar UUID;
    cl_fusion UUID;
    cl_cascade UUID;
    cl_meridian UUID;
    cl_apex UUID;

    -- Project IDs
    pj_website UUID;
    pj_mobile UUID;
    pj_crm UUID;
    pj_marketing UUID;
    pj_analytics UUID;
    pj_rebrand UUID;
    pj_automation UUID;
    pj_expansion UUID;

BEGIN

-- ============================================
-- TEAM MEMBERS (5 members)
-- ============================================

INSERT INTO team_members (id, user_id, name, email, role, status, capacity, skills, hourly_rate)
VALUES
    (gen_random_uuid(), test_user_id, 'Sarah Chen', 'sarah.chen@company.com', 'Lead Developer', 'AVAILABLE', 40, '["React", "TypeScript", "Node.js", "PostgreSQL"]', 125.00),
    (gen_random_uuid(), test_user_id, 'Mike Rodriguez', 'mike.r@company.com', 'UI/UX Designer', 'BUSY', 35, '["Figma", "CSS", "User Research", "Prototyping"]', 95.00),
    (gen_random_uuid(), test_user_id, 'Lisa Park', 'lisa.park@company.com', 'Project Manager', 'AVAILABLE', 45, '["Agile", "Scrum", "Stakeholder Management"]', 110.00),
    (gen_random_uuid(), test_user_id, 'John Smith', 'john.s@company.com', 'Backend Developer', 'OOO', 40, '["Go", "Python", "AWS", "Docker"]', 115.00),
    (gen_random_uuid(), test_user_id, 'Emily Watson', 'emily.w@company.com', 'Marketing Lead', 'AVAILABLE', 38, '["SEO", "Content Strategy", "Analytics", "Social Media"]', 90.00)
ON CONFLICT DO NOTHING;

SELECT id INTO tm_sarah FROM team_members WHERE user_id = test_user_id AND name = 'Sarah Chen';
SELECT id INTO tm_mike FROM team_members WHERE user_id = test_user_id AND name = 'Mike Rodriguez';
SELECT id INTO tm_lisa FROM team_members WHERE user_id = test_user_id AND name = 'Lisa Park';
SELECT id INTO tm_john FROM team_members WHERE user_id = test_user_id AND name = 'John Smith';
SELECT id INTO tm_emily FROM team_members WHERE user_id = test_user_id AND name = 'Emily Watson';

-- ============================================
-- CLIENTS (15 clients - mixed statuses)
-- ============================================

INSERT INTO clients (id, user_id, name, type, email, phone, website, industry, company_size, status, source, lifetime_value, notes, city, country, last_contacted_at)
VALUES
    -- Active clients (5)
    (gen_random_uuid(), test_user_id, 'TechCorp Industries', 'company', 'contact@techcorp.io', '+1-555-0101', 'https://techcorp.io', 'Technology', '51-200', 'active', 'Referral', 85000.00, 'Key enterprise client. Expanding contract Q2.', 'San Francisco', 'USA', NOW() - INTERVAL '2 days'),
    (gen_random_uuid(), test_user_id, 'GreenLeaf Solutions', 'company', 'hello@greenleaf.co', '+1-555-0102', 'https://greenleaf.co', 'Sustainability', '11-50', 'active', 'Conference', 42000.00, 'Great partnership potential. Interested in mobile app.', 'Portland', 'USA', NOW() - INTERVAL '1 day'),
    (gen_random_uuid(), test_user_id, 'Urban Development Group', 'company', 'projects@urbandev.com', '+1-555-0103', NULL, 'Real Estate', '201-500', 'active', 'Cold Outreach', 120000.00, 'Major construction project starting Q3.', 'Chicago', 'USA', NOW() - INTERVAL '5 days'),
    (gen_random_uuid(), test_user_id, 'Skyline Ventures', 'company', 'invest@skylinevc.com', '+1-555-0104', 'https://skylinevc.com', 'Finance', '11-50', 'active', 'LinkedIn', 65000.00, 'Investment firm. Needs portfolio dashboard.', 'New York', 'USA', NOW() - INTERVAL '3 days'),
    (gen_random_uuid(), test_user_id, 'Marcus Thompson', 'individual', 'marcus.t@email.com', '+1-555-0105', NULL, 'Consulting', '1-10', 'active', 'Website', 18000.00, 'Independent consultant. Repeat client.', 'Austin', 'USA', NOW()),

    -- Leads (4)
    (gen_random_uuid(), test_user_id, 'BrightPath Education', 'company', 'info@brightpath.edu', '+1-555-0106', 'https://brightpath.edu', 'Education', '51-200', 'lead', 'Webinar', NULL, 'Attended our webinar. Interested in LMS integration.', 'Boston', 'USA', NULL),
    (gen_random_uuid(), test_user_id, 'Nexus Health Systems', 'company', 'contact@nexushealth.org', '+1-555-0107', 'https://nexushealth.org', 'Healthcare', '501-1000', 'lead', 'Trade Show', NULL, 'Met at HealthTech 2026. HIPAA compliance needed.', 'Denver', 'USA', NULL),
    (gen_random_uuid(), test_user_id, 'Harmony Wellness', 'company', 'team@harmonywellness.com', '+1-555-0108', NULL, 'Health & Wellness', '11-50', 'lead', 'Referral', NULL, 'Spa chain looking for booking system.', 'Miami', 'USA', NOW() - INTERVAL '10 days'),
    (gen_random_uuid(), test_user_id, 'David Chen', 'individual', 'david.chen@gmail.com', '+1-555-0109', NULL, 'Startup', '1-10', 'lead', 'LinkedIn', NULL, 'Founder of AI startup. Exploring options.', 'Seattle', 'USA', NULL),

    -- Prospects (3)
    (gen_random_uuid(), test_user_id, 'Pinnacle Manufacturing', 'company', 'sales@pinnaclemfg.com', '+1-555-0110', 'https://pinnaclemfg.com', 'Manufacturing', '201-500', 'prospect', 'Cold Outreach', NULL, 'Initial meeting scheduled. ERP integration inquiry.', 'Detroit', 'USA', NOW() - INTERVAL '7 days'),
    (gen_random_uuid(), test_user_id, 'Elevate Marketing Agency', 'company', 'new@elevatemktg.com', '+1-555-0111', 'https://elevatemktg.com', 'Marketing', '11-50', 'prospect', 'Website', NULL, 'Requested demo. Needs client portal.', 'Los Angeles', 'USA', NOW() - INTERVAL '4 days'),
    (gen_random_uuid(), test_user_id, 'Quantum Analytics', 'company', 'hello@quantumanalytics.io', '+1-555-0112', 'https://quantumanalytics.io', 'Data Analytics', '51-200', 'prospect', 'Conference', NULL, 'Strong fit. Proposal sent last week.', 'Atlanta', 'USA', NOW() - INTERVAL '6 days'),

    -- Inactive (2)
    (gen_random_uuid(), test_user_id, 'Stellar Logistics', 'company', 'ops@stellarlogistics.com', '+1-555-0113', NULL, 'Logistics', '51-200', 'inactive', 'Referral', 35000.00, 'Project completed. Potential for phase 2.', 'Dallas', 'USA', NOW() - INTERVAL '60 days'),
    (gen_random_uuid(), test_user_id, 'Fusion Media Group', 'company', 'contact@fusionmedia.tv', '+1-555-0114', 'https://fusionmedia.tv', 'Media', '11-50', 'inactive', 'Website', 22000.00, 'Budget constraints. Re-engage Q3.', 'Nashville', 'USA', NOW() - INTERVAL '45 days'),

    -- Churned (1)
    (gen_random_uuid(), test_user_id, 'Cascade Software', 'company', 'info@cascadesw.com', '+1-555-0115', NULL, 'Software', '11-50', 'churned', 'LinkedIn', 15000.00, 'Went with competitor. Price sensitive.', 'Phoenix', 'USA', NOW() - INTERVAL '90 days');

-- Get client IDs for relationships
SELECT id INTO cl_techcorp FROM clients WHERE user_id = test_user_id AND name = 'TechCorp Industries';
SELECT id INTO cl_greenleaf FROM clients WHERE user_id = test_user_id AND name = 'GreenLeaf Solutions';
SELECT id INTO cl_urbandev FROM clients WHERE user_id = test_user_id AND name = 'Urban Development Group';
SELECT id INTO cl_skyline FROM clients WHERE user_id = test_user_id AND name = 'Skyline Ventures';
SELECT id INTO cl_brightpath FROM clients WHERE user_id = test_user_id AND name = 'BrightPath Education';
SELECT id INTO cl_pinnacle FROM clients WHERE user_id = test_user_id AND name = 'Pinnacle Manufacturing';
SELECT id INTO cl_elevate FROM clients WHERE user_id = test_user_id AND name = 'Elevate Marketing Agency';
SELECT id INTO cl_quantum FROM clients WHERE user_id = test_user_id AND name = 'Quantum Analytics';

-- ============================================
-- PROJECTS (8 projects)
-- ============================================

INSERT INTO projects (id, user_id, name, description, status, priority, client_id, project_type, start_date, due_date)
VALUES
    (gen_random_uuid(), test_user_id, 'TechCorp Website Redesign', 'Complete overhaul of corporate website with modern design and improved UX', 'ACTIVE', 'HIGH', cl_techcorp, 'client', CURRENT_DATE - INTERVAL '30 days', CURRENT_DATE + INTERVAL '14 days'),
    (gen_random_uuid(), test_user_id, 'GreenLeaf Mobile App', 'iOS and Android app for sustainability tracking', 'ACTIVE', 'HIGH', cl_greenleaf, 'client', CURRENT_DATE - INTERVAL '14 days', CURRENT_DATE + INTERVAL '45 days'),
    (gen_random_uuid(), test_user_id, 'Internal CRM System', 'Build custom CRM to replace spreadsheets', 'ACTIVE', 'MEDIUM', NULL, 'internal', CURRENT_DATE - INTERVAL '60 days', CURRENT_DATE + INTERVAL '30 days'),
    (gen_random_uuid(), test_user_id, 'Q1 Marketing Campaign', 'Multi-channel marketing push for new product launch', 'ACTIVE', 'HIGH', NULL, 'internal', CURRENT_DATE - INTERVAL '7 days', CURRENT_DATE + INTERVAL '21 days'),
    (gen_random_uuid(), test_user_id, 'Analytics Dashboard', 'Real-time business intelligence dashboard', 'PAUSED', 'MEDIUM', cl_skyline, 'client', CURRENT_DATE - INTERVAL '45 days', CURRENT_DATE + INTERVAL '60 days'),
    (gen_random_uuid(), test_user_id, 'Urban Dev Portal', 'Project management portal for construction projects', 'ACTIVE', 'HIGH', cl_urbandev, 'client', CURRENT_DATE + INTERVAL '14 days', CURRENT_DATE + INTERVAL '90 days'),
    (gen_random_uuid(), test_user_id, 'Process Automation', 'Automate repetitive internal workflows', 'ACTIVE', 'LOW', NULL, 'internal', CURRENT_DATE - INTERVAL '21 days', CURRENT_DATE + INTERVAL '7 days'),
    (gen_random_uuid(), test_user_id, 'Market Expansion Research', 'Research and planning for EU market entry', 'COMPLETED', 'MEDIUM', NULL, 'internal', CURRENT_DATE - INTERVAL '90 days', CURRENT_DATE - INTERVAL '30 days');

-- Get project IDs
SELECT id INTO pj_website FROM projects WHERE user_id = test_user_id AND name = 'TechCorp Website Redesign';
SELECT id INTO pj_mobile FROM projects WHERE user_id = test_user_id AND name = 'GreenLeaf Mobile App';
SELECT id INTO pj_crm FROM projects WHERE user_id = test_user_id AND name = 'Internal CRM System';
SELECT id INTO pj_marketing FROM projects WHERE user_id = test_user_id AND name = 'Q1 Marketing Campaign';
SELECT id INTO pj_analytics FROM projects WHERE user_id = test_user_id AND name = 'Analytics Dashboard';
SELECT id INTO pj_rebrand FROM projects WHERE user_id = test_user_id AND name = 'Urban Dev Portal';
SELECT id INTO pj_automation FROM projects WHERE user_id = test_user_id AND name = 'Process Automation';

-- ============================================
-- TASKS (30 tasks - varied statuses/priorities)
-- ============================================

INSERT INTO tasks (user_id, title, description, status, priority, due_date, project_id, assignee_id, position)
VALUES
    -- TechCorp Website tasks
    (test_user_id, 'Finalize homepage wireframes', 'Complete wireframe designs for homepage and get stakeholder approval', 'done', 'high', CURRENT_DATE - INTERVAL '5 days', pj_website, tm_mike, 1),
    (test_user_id, 'Implement responsive navigation', 'Build mobile-first navigation component with dropdown menus', 'in_progress', 'high', CURRENT_DATE + INTERVAL '2 days', pj_website, tm_sarah, 2),
    (test_user_id, 'Create product showcase section', 'Design and develop interactive product gallery', 'todo', 'medium', CURRENT_DATE + INTERVAL '5 days', pj_website, tm_sarah, 3),
    (test_user_id, 'SEO optimization', 'Implement meta tags, schema markup, and optimize page speed', 'todo', 'medium', CURRENT_DATE + INTERVAL '10 days', pj_website, tm_emily, 4),
    (test_user_id, 'Client review meeting prep', 'Prepare demo and talking points for stakeholder review', 'todo', 'high', CURRENT_DATE + INTERVAL '1 day', pj_website, tm_lisa, 5),

    -- GreenLeaf Mobile tasks
    (test_user_id, 'Design app onboarding flow', 'Create intuitive first-time user experience', 'done', 'high', CURRENT_DATE - INTERVAL '7 days', pj_mobile, tm_mike, 1),
    (test_user_id, 'Build user authentication', 'Implement OAuth and email/password auth', 'done', 'high', CURRENT_DATE - INTERVAL '3 days', pj_mobile, tm_john, 2),
    (test_user_id, 'Develop carbon footprint calculator', 'Core feature: Calculate and track carbon emissions', 'in_progress', 'high', CURRENT_DATE + INTERVAL '7 days', pj_mobile, tm_sarah, 3),
    (test_user_id, 'Push notification system', 'Set up Firebase for push notifications', 'todo', 'medium', CURRENT_DATE + INTERVAL '14 days', pj_mobile, tm_john, 4),
    (test_user_id, 'Beta testing coordination', 'Recruit testers and set up TestFlight/Play Console', 'todo', 'low', CURRENT_DATE + INTERVAL '21 days', pj_mobile, tm_lisa, 5),

    -- CRM System tasks
    (test_user_id, 'Database schema design', 'Design normalized schema for contacts, deals, activities', 'done', 'high', CURRENT_DATE - INTERVAL '45 days', pj_crm, tm_john, 1),
    (test_user_id, 'Contact management module', 'CRUD operations for contacts with search and filtering', 'done', 'high', CURRENT_DATE - INTERVAL '30 days', pj_crm, tm_sarah, 2),
    (test_user_id, 'Deal pipeline visualization', 'Kanban board for tracking deals through stages', 'in_progress', 'high', CURRENT_DATE + INTERVAL '5 days', pj_crm, tm_sarah, 3),
    (test_user_id, 'Email integration', 'Sync emails from Gmail/Outlook to contact records', 'todo', 'medium', CURRENT_DATE + INTERVAL '14 days', pj_crm, tm_john, 4),
    (test_user_id, 'Reporting dashboard', 'Sales metrics and conversion analytics', 'todo', 'medium', CURRENT_DATE + INTERVAL '21 days', pj_crm, NULL, 5),

    -- Marketing Campaign tasks
    (test_user_id, 'Draft launch email sequence', 'Write 5-email nurture sequence for product launch', 'in_progress', 'high', CURRENT_DATE + INTERVAL '3 days', pj_marketing, tm_emily, 1),
    (test_user_id, 'Create social media calendar', 'Plan 30 days of social content across platforms', 'todo', 'high', CURRENT_DATE + INTERVAL '4 days', pj_marketing, tm_emily, 2),
    (test_user_id, 'Design promotional graphics', 'Banner ads, social images, email headers', 'todo', 'medium', CURRENT_DATE + INTERVAL '6 days', pj_marketing, tm_mike, 3),
    (test_user_id, 'Set up tracking pixels', 'Configure analytics for campaign attribution', 'todo', 'low', CURRENT_DATE + INTERVAL '5 days', pj_marketing, tm_sarah, 4),
    (test_user_id, 'Influencer outreach', 'Contact and negotiate with 10 industry influencers - BLOCKED: waiting on budget approval', 'todo', 'medium', CURRENT_DATE + INTERVAL '7 days', pj_marketing, tm_emily, 5),

    -- Process Automation tasks
    (test_user_id, 'Map current workflows', 'Document existing manual processes', 'done', 'medium', CURRENT_DATE - INTERVAL '14 days', pj_automation, tm_lisa, 1),
    (test_user_id, 'Invoice generation automation', 'Auto-generate invoices from project completion', 'in_progress', 'medium', CURRENT_DATE + INTERVAL '3 days', pj_automation, tm_john, 2),
    (test_user_id, 'Client onboarding checklist', 'Automated task creation for new clients', 'todo', 'low', CURRENT_DATE + INTERVAL '5 days', pj_automation, tm_lisa, 3),

    -- Unassigned/standalone tasks
    (test_user_id, 'Review Q4 financials', 'Analyze Q4 revenue and expenses for board meeting', 'todo', 'high', CURRENT_DATE + INTERVAL '2 days', NULL, NULL, 1),
    (test_user_id, 'Update team handbook', 'Add new PTO policy and remote work guidelines', 'todo', 'low', CURRENT_DATE + INTERVAL '14 days', NULL, tm_lisa, 2),
    (test_user_id, 'Server security audit', 'Quarterly security review and penetration testing', 'todo', 'high', CURRENT_DATE + INTERVAL '7 days', NULL, tm_john, 3),
    (test_user_id, 'Prepare Quantum Analytics proposal', 'Draft SOW and pricing for analytics platform', 'in_progress', 'high', CURRENT_DATE + INTERVAL '1 day', NULL, tm_lisa, 4),
    (test_user_id, 'Schedule Pinnacle demo', 'Coordinate calendars for product demonstration', 'todo', 'medium', CURRENT_DATE + INTERVAL '3 days', NULL, tm_lisa, 5),

    -- Overdue tasks (for testing UI handling)
    (test_user_id, 'Follow up with BrightPath', 'Send follow-up email after webinar registration', 'todo', 'medium', CURRENT_DATE - INTERVAL '3 days', NULL, tm_emily, 6),
    (test_user_id, 'Fix login bug on staging', 'Users reporting intermittent login failures', 'in_progress', 'critical', CURRENT_DATE - INTERVAL '1 day', pj_crm, tm_sarah, 7);

-- ============================================
-- CALENDAR EVENTS (15 events - next 14 days)
-- ============================================

INSERT INTO calendar_events (user_id, title, description, start_time, end_time, location, meeting_type, project_id, client_id, status, attendees)
VALUES
    -- Today
    (test_user_id, 'Daily Standup', 'Team sync - what we did, what we''re doing, blockers', CURRENT_DATE + TIME '09:00', CURRENT_DATE + TIME '09:30', 'Zoom', 'team', NULL, NULL, 'confirmed', '[{"email": "sarah.chen@company.com"}, {"email": "mike.r@company.com"}, {"email": "lisa.park@company.com"}]'),
    (test_user_id, 'TechCorp Website Review', 'Review progress with client stakeholders', CURRENT_DATE + TIME '14:00', CURRENT_DATE + TIME '15:00', 'Google Meet', 'client', pj_website, cl_techcorp, 'confirmed', '[{"email": "contact@techcorp.io"}, {"email": "sarah.chen@company.com"}]'),

    -- Tomorrow
    (test_user_id, 'GreenLeaf Sprint Planning', 'Plan next sprint for mobile app development', CURRENT_DATE + INTERVAL '1 day' + TIME '10:00', CURRENT_DATE + INTERVAL '1 day' + TIME '11:30', 'Conference Room A', 'team', pj_mobile, cl_greenleaf, 'confirmed', '[{"email": "hello@greenleaf.co"}, {"email": "sarah.chen@company.com"}, {"email": "john.s@company.com"}]'),
    (test_user_id, 'Lunch with Marcus Thompson', 'Quarterly check-in and discuss expansion', CURRENT_DATE + INTERVAL '1 day' + TIME '12:30', CURRENT_DATE + INTERVAL '1 day' + TIME '13:30', 'The Capital Grille', 'one_on_one', NULL, NULL, 'confirmed', '[{"email": "marcus.t@email.com"}]'),

    -- This week
    (test_user_id, 'Marketing Campaign Kickoff', 'Align on Q1 campaign strategy and timelines', CURRENT_DATE + INTERVAL '2 days' + TIME '11:00', CURRENT_DATE + INTERVAL '2 days' + TIME '12:00', 'Zoom', 'team', pj_marketing, NULL, 'confirmed', '[{"email": "emily.w@company.com"}, {"email": "mike.r@company.com"}]'),
    (test_user_id, 'Pinnacle Manufacturing Demo', 'Product demonstration for ERP integration', CURRENT_DATE + INTERVAL '3 days' + TIME '15:00', CURRENT_DATE + INTERVAL '3 days' + TIME '16:00', 'Google Meet', 'sales', NULL, cl_pinnacle, 'tentative', '[{"email": "sales@pinnaclemfg.com"}, {"email": "lisa.park@company.com"}]'),
    (test_user_id, 'Team Happy Hour', 'Celebrate project milestones', CURRENT_DATE + INTERVAL '4 days' + TIME '17:00', CURRENT_DATE + INTERVAL '4 days' + TIME '19:00', 'Rooftop Bar Downtown', 'other', NULL, NULL, 'confirmed', '[]'),
    (test_user_id, 'Quantum Analytics Proposal Review', 'Internal review before sending proposal', CURRENT_DATE + INTERVAL '3 days' + TIME '09:00', CURRENT_DATE + INTERVAL '3 days' + TIME '10:00', 'Conference Room B', 'internal', NULL, cl_quantum, 'confirmed', '[{"email": "lisa.park@company.com"}, {"email": "sarah.chen@company.com"}]'),

    -- Next week
    (test_user_id, 'Board Meeting', 'Q4 Review and 2026 Planning', CURRENT_DATE + INTERVAL '7 days' + TIME '09:00', CURRENT_DATE + INTERVAL '7 days' + TIME '12:00', 'Executive Boardroom', 'other', NULL, NULL, 'confirmed', '[]'),
    (test_user_id, 'Skyline Ventures Check-in', 'Monthly project status update', CURRENT_DATE + INTERVAL '8 days' + TIME '14:00', CURRENT_DATE + INTERVAL '8 days' + TIME '14:30', 'Phone', 'client', pj_analytics, cl_skyline, 'confirmed', '[{"email": "invest@skylinevc.com"}]'),
    (test_user_id, 'Design Review - Mobile App', 'Review UI designs before development', CURRENT_DATE + INTERVAL '9 days' + TIME '11:00', CURRENT_DATE + INTERVAL '9 days' + TIME '12:00', 'Figma', 'team', pj_mobile, NULL, 'confirmed', '[{"email": "mike.r@company.com"}, {"email": "sarah.chen@company.com"}]'),
    (test_user_id, 'Urban Dev Kickoff Meeting', 'Project kickoff with stakeholders', CURRENT_DATE + INTERVAL '10 days' + TIME '10:00', CURRENT_DATE + INTERVAL '10 days' + TIME '11:30', 'Client Office - Chicago', 'client', pj_rebrand, cl_urbandev, 'tentative', '[{"email": "projects@urbandev.com"}, {"email": "lisa.park@company.com"}]'),
    (test_user_id, 'Security Audit Review', 'Review findings from quarterly audit', CURRENT_DATE + INTERVAL '11 days' + TIME '15:00', CURRENT_DATE + INTERVAL '11 days' + TIME '16:00', 'Zoom', 'internal', NULL, NULL, 'confirmed', '[{"email": "john.s@company.com"}]'),

    -- All-day events
    (test_user_id, 'Company Retreat', 'Annual team building and strategy session', CURRENT_DATE + INTERVAL '12 days', CURRENT_DATE + INTERVAL '13 days', 'Lake Tahoe Resort', 'other', NULL, NULL, 'confirmed', '[]'),

    -- Cancelled event
    (test_user_id, 'Cancelled: Elevate Agency Call', 'Rescheduled to next month', CURRENT_DATE + INTERVAL '5 days' + TIME '13:00', CURRENT_DATE + INTERVAL '5 days' + TIME '14:00', 'Zoom', 'client', NULL, cl_elevate, 'cancelled', '[{"email": "new@elevatemktg.com"}]');

-- ============================================
-- EMAILS (15 emails - inbox simulation)
-- ============================================

INSERT INTO emails (user_id, provider, external_id, thread_id, subject, snippet, from_email, from_name, to_emails, body_text, is_read, is_starred, is_important, date)
VALUES
    -- Unread important emails
    (test_user_id, 'gmail', 'msg-001', 'thread-001', 'RE: TechCorp Website - Urgent Feedback', 'Hi team, we''ve reviewed the latest designs and have some critical feedback...', 'ceo@techcorp.io', 'Jennifer Walsh', '["you@company.com"]', 'Hi team,\n\nWe''ve reviewed the latest designs and have some critical feedback that needs to be addressed before we can move forward:\n\n1. The hero section needs more emphasis on our enterprise solutions\n2. Mobile navigation is confusing - can we simplify?\n3. Load time seems slow - any optimizations possible?\n\nLet''s discuss tomorrow.\n\nBest,\nJennifer', false, true, true, NOW() - INTERVAL '2 hours'),

    (test_user_id, 'gmail', 'msg-002', 'thread-002', 'Proposal Request - Enterprise Solution', 'Hello, I came across your company at the recent tech conference and I''m interested...', 'david.chen@innovate.ai', 'David Chen', '["you@company.com"]', 'Hello,\n\nI came across your company at the recent tech conference and I''m interested in discussing a potential partnership. We''re a Series A startup focused on AI solutions and need help building our customer-facing platform.\n\nWould you be available for a call this week?\n\nBest regards,\nDavid Chen\nFounder, Innovate AI', false, false, true, NOW() - INTERVAL '5 hours'),

    (test_user_id, 'gmail', 'msg-003', 'thread-003', 'Invoice #2024-089 - Payment Received', 'Thank you for your payment of $12,500.00 for Invoice #2024-089...', 'accounting@greenleaf.co', 'GreenLeaf Accounting', '["billing@company.com"]', 'Thank you for your payment of $12,500.00 for Invoice #2024-089.\n\nPayment Details:\n- Amount: $12,500.00\n- Method: ACH Transfer\n- Reference: GRN-2024-089\n\nThank you for your business!\n\nGreenLeaf Accounting Team', false, false, false, NOW() - INTERVAL '1 day'),

    -- Read emails
    (test_user_id, 'gmail', 'msg-004', 'thread-004', 'Weekly Team Update - Jan 20-24', 'Here''s our weekly summary of accomplishments and upcoming priorities...', 'lisa.park@company.com', 'Lisa Park', '["team@company.com"]', 'Team,\n\nHere''s our weekly summary:\n\nCompleted:\n- TechCorp wireframes approved\n- GreenLeaf auth system deployed\n- CRM pipeline feature 80% complete\n\nUpcoming:\n- Marketing campaign launch prep\n- Quantum Analytics proposal due Friday\n\nBlockers:\n- Need design assets for mobile app\n\nLet me know if you have questions!\n\nLisa', true, false, false, NOW() - INTERVAL '1 day'),

    (test_user_id, 'gmail', 'msg-005', 'thread-005', 'RE: Meeting Reschedule Request', 'That works for me. See you Thursday at 3pm instead.', 'invest@skylinevc.com', 'Robert Kim', '["you@company.com"]', 'That works for me. See you Thursday at 3pm instead.\n\nI''ll send an updated calendar invite.\n\nRobert', true, false, false, NOW() - INTERVAL '2 days'),

    (test_user_id, 'gmail', 'msg-006', 'thread-006', 'New Lead: BrightPath Education', 'A new lead has been captured from your webinar registration page...', 'notifications@crm.company.com', 'CRM System', '["sales@company.com"]', 'A new lead has been captured:\n\nCompany: BrightPath Education\nContact: Amanda Foster\nEmail: amanda@brightpath.edu\nSource: Webinar Registration\nInterest: LMS Integration\n\nFollow up within 24 hours for best results.', true, true, false, NOW() - INTERVAL '3 days'),

    (test_user_id, 'gmail', 'msg-007', 'thread-007', 'GitHub: [businessos] Pull Request #234 merged', 'Pull request #234 has been merged into main by sarah-chen...', 'noreply@github.com', 'GitHub', '["dev@company.com"]', 'Pull request #234 has been merged into main.\n\nTitle: feat: Add deal pipeline kanban view\nAuthor: sarah-chen\nReviewers: john-smith\n\nChanges: +847 -123\n\nView on GitHub: https://github.com/company/businessos/pull/234', true, false, false, NOW() - INTERVAL '4 hours'),

    (test_user_id, 'gmail', 'msg-008', 'thread-008', 'Your AWS Bill for January 2026', 'Your AWS charges for this billing period total $2,847.23...', 'no-reply@aws.amazon.com', 'Amazon Web Services', '["billing@company.com"]', 'Your AWS Bill\n\nBilling Period: January 1-31, 2026\nTotal: $2,847.23\n\nTop Services:\n- EC2: $1,234.56\n- RDS: $892.34\n- S3: $234.12\n- Lambda: $156.78\n\nView detailed bill in AWS Console.', true, false, false, NOW() - INTERVAL '5 days'),

    -- Starred emails
    (test_user_id, 'gmail', 'msg-009', 'thread-009', 'Contract Draft - Urban Development Group', 'Please find attached the draft contract for the portal development project...', 'legal@urbandev.com', 'Sarah Martinez', '["you@company.com"]', 'Hi,\n\nPlease find attached the draft contract for the portal development project. Key terms:\n\n- Project value: $120,000\n- Timeline: 6 months\n- Payment: 30% upfront, 70% on milestones\n\nPlease review and send any redlines by Friday.\n\nBest,\nSarah Martinez\nGeneral Counsel', true, true, true, NOW() - INTERVAL '6 days'),

    (test_user_id, 'gmail', 'msg-010', 'thread-010', 'Referral: Nexus Health Systems', 'Hi! My colleague at Nexus Health mentioned they''re looking for a development partner...', 'marcus.t@email.com', 'Marcus Thompson', '["you@company.com"]', 'Hi!\n\nMy colleague at Nexus Health mentioned they''re looking for a development partner for their patient portal. I immediately thought of you.\n\nTheir contact is James Wilson (james.w@nexushealth.org). They''re a larger org with serious budget.\n\nHappy to make an intro if helpful!\n\nMarcus', true, true, false, NOW() - INTERVAL '8 days'),

    -- Draft
    (test_user_id, 'gmail', 'msg-011', 'thread-011', 'DRAFT: Quarterly Newsletter', 'Q1 2026 Newsletter Draft - Company Updates and Achievements...', 'you@company.com', 'You', '[]', 'Subject: Q1 2026 Updates\n\nDear Valued Clients,\n\nWe''re excited to share our latest updates:\n\n[TODO: Add project highlights]\n[TODO: Include team news]\n[TODO: Mention upcoming webinar]\n\nThank you for your continued partnership.\n\nBest regards,', true, false, false, NOW() - INTERVAL '2 days'),

    -- Sent emails
    (test_user_id, 'gmail', 'msg-012', 'thread-012', 'RE: Proposal Request - Enterprise Solution', 'Hi David, Thank you for reaching out! I''d be happy to schedule a call...', 'you@company.com', 'You', '["david.chen@innovate.ai"]', 'Hi David,\n\nThank you for reaching out! I''d be happy to schedule a call to discuss your platform needs.\n\nHow does Thursday at 2pm PST work for you?\n\nLooking forward to learning more about Innovate AI.\n\nBest regards', true, false, false, NOW() - INTERVAL '4 hours'),

    (test_user_id, 'gmail', 'msg-013', 'thread-002', 'Follow-up: Quantum Analytics Proposal', 'Hi team, Following up on our discussion last week regarding the analytics platform...', 'you@company.com', 'You', '["hello@quantumanalytics.io"]', 'Hi team,\n\nFollowing up on our discussion last week regarding the analytics platform. I''ve attached our formal proposal with detailed scope and pricing.\n\nKey highlights:\n- 3-month implementation\n- Custom dashboards\n- API integrations\n- 24/7 support\n\nPlease let me know if you have any questions.\n\nBest regards', true, false, false, NOW() - INTERVAL '6 days'),

    -- Newsletter/promotional (read)
    (test_user_id, 'gmail', 'msg-014', 'thread-014', 'This Week in Tech: AI Development Trends', 'The latest trends in AI-powered development tools and what they mean for your team...', 'newsletter@techweekly.io', 'Tech Weekly', '["you@company.com"]', 'THIS WEEK IN TECH\n\nTop Stories:\n\n1. AI Code Assistants Hit 50% Developer Adoption\n2. New Framework Promises 10x Faster Builds\n3. Security Concerns Rise with AI-Generated Code\n\nRead more at techweekly.io', true, false, false, NOW() - INTERVAL '3 days'),

    (test_user_id, 'gmail', 'msg-015', 'thread-015', 'Reminder: Company Retreat Registration', 'Don''t forget to complete your registration for the annual retreat...', 'hr@company.com', 'HR Team', '["all@company.com"]', 'Hi everyone,\n\nFriendly reminder to complete your registration for our annual company retreat at Lake Tahoe (Feb 3-4).\n\nPlease fill out the form by this Friday:\n- Dietary restrictions\n- Room preferences\n- Activity choices\n\nLooking forward to seeing everyone there!\n\nHR Team', true, false, false, NOW() - INTERVAL '4 days');

-- Update is_sent and is_draft flags
UPDATE emails SET is_sent = true WHERE user_id = test_user_id AND from_email = 'you@company.com' AND subject NOT LIKE 'DRAFT:%';
UPDATE emails SET is_draft = true WHERE user_id = test_user_id AND subject LIKE 'DRAFT:%';

RAISE NOTICE 'Seed data inserted successfully for user: %', test_user_id;
RAISE NOTICE 'Created: 5 team members, 15 clients, 8 projects, 30 tasks, 15 calendar events, 15 emails';

END $$;
