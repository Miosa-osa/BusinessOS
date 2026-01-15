-- BusinessOS Development Seed Data
-- This script creates demo data for local development and testing
-- Run with: psql $DATABASE_URL < scripts/seed.sql

-- Note: Better Auth handles user/session tables
-- We'll use placeholder user IDs that you'll need to replace with actual user IDs from your auth system

BEGIN;

-- Clean existing seed data (be careful in production!)
-- Uncomment if you want to reset seed data
-- DELETE FROM messages WHERE conversation_id IN (SELECT id FROM conversations WHERE title LIKE 'Demo:%');
-- DELETE FROM conversations WHERE title LIKE 'Demo:%';
-- DELETE FROM tasks WHERE project_id IN (SELECT id FROM projects WHERE name LIKE 'Demo%');
-- DELETE FROM projects WHERE name LIKE 'Demo%';
-- DELETE FROM workspace_members WHERE workspace_id IN (SELECT id FROM workspaces WHERE name LIKE 'Demo%');
-- DELETE FROM workspaces WHERE name LIKE 'Demo%';
-- DELETE FROM clients WHERE name LIKE 'Demo%';

-- ============================================================================
-- WORKSPACES
-- ============================================================================

-- Demo workspace for solo developer
INSERT INTO workspaces (id, name, slug, description, settings, created_at)
VALUES
    ('00000000-0000-0000-0000-000000000001', 'Demo Workspace', 'demo-workspace',
     'Personal workspace for testing BusinessOS features',
     '{"theme": "light", "timezone": "America/New_York"}'::jsonb,
     NOW() - INTERVAL '30 days'),

    ('00000000-0000-0000-0000-000000000002', 'Acme Corp Team', 'acme-corp',
     'Collaborative workspace for Acme Corp product team',
     '{"theme": "dark", "timezone": "America/Los_Angeles"}'::jsonb,
     NOW() - INTERVAL '60 days')
ON CONFLICT (id) DO NOTHING;

-- ============================================================================
-- CLIENTS (CRM)
-- ============================================================================

INSERT INTO clients (id, user_id, name, type, status, email, phone, website, notes, created_at)
VALUES
    ('10000000-0000-0000-0000-000000000001', 'demo-user-1', 'TechStart Inc', 'company', 'active',
     'contact@techstart.example.com', '+1-555-0101', 'https://techstart.example.com',
     'B2B SaaS startup, 50 employees, Series A funded', NOW() - INTERVAL '90 days'),

    ('10000000-0000-0000-0000-000000000002', 'demo-user-1', 'Global Retail Co', 'company', 'prospect',
     'partnerships@globalretail.example.com', '+1-555-0102', 'https://globalretail.example.com',
     'Enterprise retail chain, 10,000+ employees, evaluating our platform', NOW() - INTERVAL '15 days'),

    ('10000000-0000-0000-0000-000000000003', 'demo-user-1', 'Jane Smith Consulting', 'individual', 'active',
     'jane@smithconsulting.example.com', '+1-555-0103', 'https://smithconsulting.example.com',
     'Independent marketing consultant, long-term client', NOW() - INTERVAL '180 days')
ON CONFLICT (id) DO NOTHING;

-- Client Contacts
INSERT INTO client_contacts (id, client_id, name, email, phone, role, is_primary, created_at)
VALUES
    (gen_random_uuid(), '10000000-0000-0000-0000-000000000001', 'Alex Johnson',
     'alex@techstart.example.com', '+1-555-0111', 'CTO', true, NOW()),

    (gen_random_uuid(), '10000000-0000-0000-0000-000000000001', 'Maria Garcia',
     'maria@techstart.example.com', '+1-555-0112', 'Product Manager', false, NOW()),

    (gen_random_uuid(), '10000000-0000-0000-0000-000000000002', 'Robert Chen',
     'robert@globalretail.example.com', '+1-555-0121', 'VP of Technology', true, NOW())
ON CONFLICT DO NOTHING;

-- Client Deals
INSERT INTO client_deals (id, client_id, name, amount, stage, probability, expected_close_date, notes, created_at)
VALUES
    (gen_random_uuid(), '10000000-0000-0000-0000-000000000001', 'Enterprise Platform Upgrade',
     50000.00, 'negotiation', 75, CURRENT_DATE + INTERVAL '30 days',
     'Upgrading from starter to enterprise tier, custom integrations needed', NOW() - INTERVAL '20 days'),

    (gen_random_uuid(), '10000000-0000-0000-0000-000000000002', 'Multi-Store Deployment',
     250000.00, 'proposal', 50, CURRENT_DATE + INTERVAL '60 days',
     'Pilot with 5 stores, potential expansion to 50+ locations', NOW() - INTERVAL '10 days')
ON CONFLICT DO NOTHING;

-- ============================================================================
-- PROJECTS
-- ============================================================================

INSERT INTO projects (id, user_id, workspace_id, name, description, status, priority, client_id, project_type, start_date, due_date, created_at)
VALUES
    ('20000000-0000-0000-0000-000000000001', 'demo-user-1', '00000000-0000-0000-0000-000000000001',
     'Demo: Product Launch Campaign',
     'Marketing campaign for Q1 2026 product launch with integrated social media, email, and content strategy',
     'ACTIVE', 'HIGH', '10000000-0000-0000-0000-000000000001', 'client',
     CURRENT_DATE - INTERVAL '15 days', CURRENT_DATE + INTERVAL '45 days',
     NOW() - INTERVAL '15 days'),

    ('20000000-0000-0000-0000-000000000002', 'demo-user-1', '00000000-0000-0000-0000-000000000001',
     'Demo: Website Redesign',
     'Complete redesign of company website with modern UI/UX, accessibility improvements, and performance optimization',
     'ACTIVE', 'MEDIUM', NULL, 'internal',
     CURRENT_DATE - INTERVAL '30 days', CURRENT_DATE + INTERVAL '30 days',
     NOW() - INTERVAL '30 days'),

    ('20000000-0000-0000-0000-000000000003', 'demo-user-1', '00000000-0000-0000-0000-000000000001',
     'Demo: API Integration Research',
     'Research and proof-of-concept for integrating with third-party CRM and analytics platforms',
     'PAUSED', 'LOW', NULL, 'research',
     CURRENT_DATE - INTERVAL '45 days', NULL,
     NOW() - INTERVAL '45 days'),

    ('20000000-0000-0000-0000-000000000004', 'demo-user-1', '00000000-0000-0000-0000-000000000002',
     'Demo: Team Onboarding System',
     'Build automated onboarding workflow for new team members including access provisioning and training materials',
     'ACTIVE', 'CRITICAL', NULL, 'internal',
     CURRENT_DATE - INTERVAL '7 days', CURRENT_DATE + INTERVAL '14 days',
     NOW() - INTERVAL '7 days')
ON CONFLICT (id) DO NOTHING;

-- ============================================================================
-- TASKS
-- ============================================================================

INSERT INTO tasks (id, project_id, user_id, title, description, status, priority, due_date, estimated_hours, created_at)
VALUES
    -- Product Launch Campaign tasks
    (gen_random_uuid(), '20000000-0000-0000-0000-000000000001', 'demo-user-1',
     'Draft email campaign copy',
     'Write compelling email sequences for product announcement (3 emails: teaser, launch, follow-up)',
     'done', 'high', CURRENT_DATE - INTERVAL '5 days', 4, NOW() - INTERVAL '10 days'),

    (gen_random_uuid(), '20000000-0000-0000-0000-000000000001', 'demo-user-1',
     'Design social media graphics',
     'Create branded graphics for Twitter, LinkedIn, and Instagram announcements',
     'in_progress', 'high', CURRENT_DATE + INTERVAL '3 days', 6, NOW() - INTERVAL '8 days'),

    (gen_random_uuid(), '20000000-0000-0000-0000-000000000001', 'demo-user-1',
     'Set up analytics tracking',
     'Configure Google Analytics and Mixpanel event tracking for campaign performance',
     'todo', 'medium', CURRENT_DATE + INTERVAL '7 days', 3, NOW() - INTERVAL '5 days'),

    -- Website Redesign tasks
    (gen_random_uuid(), '20000000-0000-0000-0000-000000000002', 'demo-user-1',
     'Create wireframes for homepage',
     'Design lo-fi wireframes for new homepage layout with stakeholder feedback incorporated',
     'done', 'critical', CURRENT_DATE - INTERVAL '10 days', 8, NOW() - INTERVAL '20 days'),

    (gen_random_uuid(), '20000000-0000-0000-0000-000000000002', 'demo-user-1',
     'Develop responsive navigation component',
     'Build mobile-first navigation with accessibility features and smooth animations',
     'in_progress', 'high', CURRENT_DATE + INTERVAL '5 days', 12, NOW() - INTERVAL '5 days'),

    (gen_random_uuid(), '20000000-0000-0000-0000-000000000002', 'demo-user-1',
     'Performance audit and optimization',
     'Run Lighthouse audits, optimize images, implement lazy loading, improve Core Web Vitals',
     'todo', 'medium', CURRENT_DATE + INTERVAL '20 days', 6, NOW() - INTERVAL '2 days'),

    -- Team Onboarding System tasks
    (gen_random_uuid(), '20000000-0000-0000-0000-000000000004', 'demo-user-1',
     'Document access provisioning workflow',
     'Create checklist for granting access to all necessary tools (Slack, GitHub, AWS, etc.)',
     'done', 'critical', CURRENT_DATE - INTERVAL '2 days', 3, NOW() - INTERVAL '5 days'),

    (gen_random_uuid(), '20000000-0000-0000-0000-000000000004', 'demo-user-1',
     'Build automated welcome email sequence',
     'Set up email automation with welcome message, first week schedule, and resource links',
     'in_progress', 'high', CURRENT_DATE + INTERVAL '7 days', 4, NOW() - INTERVAL '3 days'),

    (gen_random_uuid(), '20000000-0000-0000-0000-000000000004', 'demo-user-1',
     'Create interactive training modules',
     'Develop 5 interactive training modules covering company culture, tools, and processes',
     'todo', 'medium', CURRENT_DATE + INTERVAL '14 days', 16, NOW() - INTERVAL '1 day')
ON CONFLICT DO NOTHING;

-- ============================================================================
-- CONVERSATIONS & MESSAGES
-- ============================================================================

-- Conversation 1: Project Planning
INSERT INTO conversations (id, user_id, title, created_at)
VALUES
    ('30000000-0000-0000-0000-000000000001', 'demo-user-1',
     'Demo: Planning Product Launch Strategy',
     NOW() - INTERVAL '10 days')
ON CONFLICT (id) DO NOTHING;

INSERT INTO messages (conversation_id, role, content, created_at)
VALUES
    ('30000000-0000-0000-0000-000000000001', 'USER',
     'I need help planning our Q1 product launch campaign. We''re launching a new enterprise feature and need to reach decision-makers at mid-to-large companies.',
     NOW() - INTERVAL '10 days'),

    ('30000000-0000-0000-0000-000000000001', 'ASSISTANT',
     'I''ll help you plan a comprehensive product launch campaign targeting enterprise decision-makers. Let me break this down into key components:

**Target Audience Analysis:**
- Decision-makers: CTOs, VPs of Engineering, Product Leaders
- Company size: 100-5000 employees
- Industries: SaaS, FinTech, E-commerce

**Launch Campaign Strategy:**

1. **Pre-Launch (2 weeks before):**
   - Teaser email to existing customers
   - LinkedIn thought leadership posts
   - Beta program for early adopters

2. **Launch Week:**
   - Official announcement blog post
   - Product Hunt launch
   - Press release distribution
   - Live demo webinar

3. **Post-Launch (2 weeks after):**
   - Case studies from beta users
   - Follow-up email sequences
   - Sales enablement materials

**Key Metrics to Track:**
- Email open rates and click-through rates
- Webinar registrations and attendance
- Demo requests from qualified leads
- Product Hunt upvotes and comments

Would you like me to create detailed task breakdowns for any of these phases?',
     NOW() - INTERVAL '10 days'),

    ('30000000-0000-0000-0000-000000000001', 'USER',
     'Yes, let''s start with the email campaign. What should be in each email?',
     NOW() - INTERVAL '10 days'),

    ('30000000-0000-0000-0000-000000000001', 'ASSISTANT',
     'Here''s a detailed email sequence strategy:

**Email 1: Teaser (T-7 days)**
- Subject: "Something big is coming... [sneak peek inside]"
- Focus: Build curiosity without revealing everything
- Include: Vague hint at the problem being solved, countdown timer
- CTA: "Join the early access list"

**Email 2: Launch Announcement (Launch day)**
- Subject: "Introducing [Feature Name]: [Key Benefit]"
- Focus: Clear value proposition and social proof
- Include: Feature overview, customer testimonial quote, pricing
- CTA: "Schedule a demo" or "Start free trial"

**Email 3: Follow-up (T+3 days)**
- Subject: "See [Feature] in action + exclusive Q&A"
- Focus: Live demo webinar invitation
- Include: Recording of product demo, FAQs, integration guides
- CTA: "Register for live Q&A session"

For each email, I recommend:
- A/B test subject lines with 10% of your list first
- Personalize with {{company_name}} and {{industry}} variables
- Keep email length under 200 words
- Mobile-first design (60%+ will read on mobile)

Want me to draft the actual copy for Email 1?',
     NOW() - INTERVAL '10 days')
ON CONFLICT DO NOTHING;

-- Conversation 2: Technical Discussion
INSERT INTO conversations (id, user_id, title, created_at)
VALUES
    ('30000000-0000-0000-0000-000000000002', 'demo-user-1',
     'Demo: Database Architecture Discussion',
     NOW() - INTERVAL '5 days')
ON CONFLICT (id) DO NOTHING;

INSERT INTO messages (conversation_id, role, content, created_at)
VALUES
    ('30000000-0000-0000-0000-000000000002', 'USER',
     'Should we use a single database or split into multiple databases as we scale?',
     NOW() - INTERVAL '5 days'),

    ('30000000-0000-0000-0000-000000000002', 'ASSISTANT',
     'Great question! The answer depends on your current scale and growth trajectory.

**Stick with Single Database if:**
- Team size < 10 developers
- Data size < 100GB
- Query performance is acceptable
- You value transactional consistency across all entities
- You''re pre-product-market fit and iterating quickly

**Consider Multiple Databases if:**
- Team size > 20 developers (microservices benefits)
- Data size > 500GB (sharding needed)
- Clear service boundaries with minimal cross-service queries
- Different scaling requirements (e.g., read-heavy analytics vs write-heavy transactions)
- Compliance requirements for data isolation

**Your Current Architecture Analysis:**
Based on typical SaaS patterns, I recommend staying with a single PostgreSQL database until you hit clear scaling limits. Here''s why:

1. **Simplicity**: No distributed transactions, easier debugging
2. **ACID guarantees**: Full consistency across all operations
3. **Query flexibility**: Can join across all tables without service calls
4. **Lower operational overhead**: One database to backup, monitor, optimize

**When to Revisit:**
- Query latency > 500ms for common operations
- Database size > 250GB
- Team splits into 3+ autonomous service teams

Want me to analyze your specific database schema to provide more tailored recommendations?',
     NOW() - INTERVAL '5 days')
ON CONFLICT DO NOTHING;

-- Conversation 3: Quick Q&A
INSERT INTO conversations (id, user_id, title, created_at)
VALUES
    ('30000000-0000-0000-0000-000000000003', 'demo-user-1',
     'Demo: Quick Design Feedback',
     NOW() - INTERVAL '2 days')
ON CONFLICT (id) DO NOTHING;

INSERT INTO messages (conversation_id, role, content, created_at)
VALUES
    ('30000000-0000-0000-0000-000000000003', 'USER',
     'What color palette would work well for a B2B SaaS dashboard?',
     NOW() - INTERVAL '2 days'),

    ('30000000-0000-0000-0000-000000000003', 'ASSISTANT',
     'For B2B SaaS dashboards, I recommend a professional, calming palette:

**Primary Colors:**
- Blue (#2563EB): Trust, stability, professional
- Neutral Gray (#64748B): Data visualization backgrounds
- White (#FFFFFF): Clean canvas for content

**Accent Colors:**
- Green (#10B981): Success states, positive metrics
- Red (#EF4444): Errors, critical alerts
- Amber (#F59E0B): Warnings, pending actions
- Purple (#8B5CF6): Premium features, highlights

**Design Principles:**
- Use 60-30-10 rule: 60% neutral, 30% primary, 10% accent
- Maintain 4.5:1 contrast ratio (WCAG AA compliance)
- Limit to 5-6 colors total to avoid visual clutter
- Use color consistently (green always = positive)

Popular B2B examples:
- Linear uses purple/blue with dark mode
- Stripe uses blue/purple gradient
- Notion uses warm grays with colored accents

Want me to generate a full Tailwind CSS color palette based on these principles?',
     NOW() - INTERVAL '2 days')
ON CONFLICT DO NOTHING;

-- ============================================================================
-- CONTEXTS (Documents/Profiles)
-- ============================================================================

INSERT INTO contexts (id, user_id, name, type, content, created_at)
VALUES
    (gen_random_uuid(), 'demo-user-1', 'Company Brand Guidelines', 'DOCUMENT',
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
     NOW() - INTERVAL '60 days'),

    (gen_random_uuid(), 'demo-user-1', 'Target Customer Profile', 'BUSINESS',
     '# Ideal Customer Profile (ICP)

## Demographics
- Company Size: 100-1000 employees
- Revenue: $10M-$100M ARR
- Industry: B2B SaaS, FinTech
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
     NOW() - INTERVAL '90 days')
ON CONFLICT DO NOTHING;

COMMIT;

-- ============================================================================
-- Seed Data Summary
-- ============================================================================
-- Created:
-- - 2 workspaces (Demo Workspace, Acme Corp Team)
-- - 3 clients (TechStart Inc, Global Retail Co, Jane Smith Consulting)
-- - 4 projects (Product Launch, Website Redesign, API Research, Onboarding)
-- - 9 tasks (across all projects with varied status)
-- - 3 conversations (with realistic AI chat messages)
-- - 2 context documents (brand guidelines, customer profile)
-- ============================================================================

SELECT 'Seed data loaded successfully! 🌱' AS status;
SELECT 'Created 2 workspaces, 3 clients, 4 projects, 9 tasks, 3 conversations' AS summary;
