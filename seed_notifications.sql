-- Notification Seed Script for Business OS
-- Run this against your database to seed test notifications
-- 
-- USAGE:
--   1. Replace 'YOUR_USER_ID' with your actual user_id (from the users/session table)
--   2. Run: psql -d your_database -f seed_notifications.sql
--   OR run the individual INSERT statements in your database client
--
-- To get your user_id, run: SELECT id FROM "user" LIMIT 5;

-- First, let's clear any existing seeded notifications (optional)
-- DELETE FROM notifications WHERE metadata->>'seeded' = 'true';

DO $$
DECLARE
    v_user_id VARCHAR(255);
    v_now TIMESTAMPTZ := NOW();
BEGIN
    -- Get the first user ID from the database
    SELECT id INTO v_user_id FROM "user" LIMIT 1;
    
    IF v_user_id IS NULL THEN
        RAISE EXCEPTION 'No users found in database. Please create a user first.';
    END IF;
    
    RAISE NOTICE 'Seeding notifications for user: %', v_user_id;

    -- ============ TODAY - URGENT/HIGH PRIORITY ============
    INSERT INTO notifications (user_id, type, title, body, entity_type, sender_name, priority, is_read, metadata, created_at)
    VALUES (v_user_id, 'task.overdue', 'Task overdue: Q4 Financial Report', 'This task was due yesterday and needs immediate attention.', 'task', 'Sarah Chen', 'urgent', false, '{"seeded": true, "seed_time": "today"}', v_now - INTERVAL '10 minutes');

    INSERT INTO notifications (user_id, type, title, body, entity_type, sender_name, priority, is_read, metadata, created_at)
    VALUES (v_user_id, 'task.due_today', 'Task due today: Review PR #142', 'Code review for the notifications feature is due today.', 'task', 'Mike Johnson', 'high', false, '{"seeded": true, "seed_time": "today"}', v_now - INTERVAL '30 minutes');

    INSERT INTO notifications (user_id, type, title, body, entity_type, sender_name, priority, is_read, metadata, created_at)
    VALUES (v_user_id, 'integration.sync_failed', 'Google Calendar sync failed', 'Unable to sync calendar events. Please reconnect your account.', 'integration', 'System', 'high', false, '{"seeded": true, "integration": "google_calendar"}', v_now - INTERVAL '1 hour');

    -- ============ TODAY - NORMAL PRIORITY ============
    INSERT INTO notifications (user_id, type, title, body, entity_type, sender_name, priority, is_read, metadata, created_at)
    VALUES (v_user_id, 'task.assigned', 'You were assigned: Update API documentation', 'Sarah Chen assigned you to this task in Project Alpha.', 'task', 'Sarah Chen', 'normal', false, '{"seeded": true, "project_name": "Project Alpha"}', v_now - INTERVAL '2 hours');

    INSERT INTO notifications (user_id, type, title, body, entity_type, sender_name, priority, is_read, metadata, created_at)
    VALUES (v_user_id, 'mention.comment', '@you in: Design Review Discussion', 'Mike Johnson mentioned you: "what do you think about this approach?"', 'comment', 'Mike Johnson', 'normal', false, '{"seeded": true}', v_now - INTERVAL '3 hours');

    INSERT INTO notifications (user_id, type, title, body, entity_type, sender_name, priority, is_read, metadata, created_at)
    VALUES (v_user_id, 'project.added', 'Added to project: Mobile App Redesign', 'You''ve been added as a contributor to this project.', 'project', 'Emily Davis', 'normal', false, '{"seeded": true, "role": "contributor"}', v_now - INTERVAL '4 hours');

    INSERT INTO notifications (user_id, type, title, body, entity_type, sender_name, priority, is_read, metadata, created_at)
    VALUES (v_user_id, 'task.comment', 'New comment on: Update API documentation', 'Emily Davis commented: "I''ve added some notes to the shared doc."', 'task', 'Emily Davis', 'normal', true, '{"seeded": true}', v_now - INTERVAL '5 hours');

    INSERT INTO notifications (user_id, type, title, body, entity_type, sender_name, priority, is_read, metadata, created_at)
    VALUES (v_user_id, 'client.meeting_scheduled', 'Meeting scheduled with Acme Corp', 'Tomorrow at 2:00 PM - Quarterly business review', 'client', 'Alex Thompson', 'normal', false, '{"seeded": true, "client_name": "Acme Corp"}', v_now - INTERVAL '6 hours');

    -- ============ YESTERDAY ============
    INSERT INTO notifications (user_id, type, title, body, entity_type, sender_name, priority, is_read, metadata, created_at)
    VALUES (v_user_id, 'task.completed', 'Task completed: Setup CI/CD pipeline', 'Alex Thompson marked this task as complete.', 'task', 'Alex Thompson', 'normal', true, '{"seeded": true}', v_now - INTERVAL '20 hours');

    INSERT INTO notifications (user_id, type, title, body, entity_type, sender_name, priority, is_read, metadata, created_at)
    VALUES (v_user_id, 'team.member_joined', 'New team member: Jordan Park', 'Jordan Park joined the Engineering team.', 'team', 'System', 'low', true, '{"seeded": true, "team_name": "Engineering"}', v_now - INTERVAL '22 hours');

    INSERT INTO notifications (user_id, type, title, body, entity_type, sender_name, priority, is_read, metadata, created_at)
    VALUES (v_user_id, 'mention.task', '@you in task: Database Migration Plan', 'Sarah Chen mentioned you in a task description.', 'task', 'Sarah Chen', 'normal', false, '{"seeded": true}', v_now - INTERVAL '26 hours');

    INSERT INTO notifications (user_id, type, title, body, entity_type, sender_name, priority, is_read, metadata, created_at)
    VALUES (v_user_id, 'project.status_changed', 'Project status: Mobile App Redesign → In Progress', 'Project moved from Planning to In Progress.', 'project', 'Emily Davis', 'normal', true, '{"seeded": true}', v_now - INTERVAL '28 hours');

    -- ============ THIS WEEK ============
    INSERT INTO notifications (user_id, type, title, body, entity_type, sender_name, priority, is_read, metadata, created_at)
    VALUES (v_user_id, 'task.due_soon', 'Task due soon: Prepare presentation slides', 'This task is due in 3 days.', 'task', 'System', 'high', false, '{"seeded": true, "days_until_due": 3}', v_now - INTERVAL '72 hours');

    INSERT INTO notifications (user_id, type, title, body, entity_type, sender_name, priority, is_read, metadata, created_at)
    VALUES (v_user_id, 'client.deal_update', 'Deal update: Acme Corp - Enterprise Plan', 'Deal value updated to $50,000. Stage: Negotiation.', 'client', 'Mike Johnson', 'normal', true, '{"seeded": true, "deal_value": 50000}', v_now - INTERVAL '96 hours');

    INSERT INTO notifications (user_id, type, title, body, entity_type, sender_name, priority, is_read, metadata, created_at)
    VALUES (v_user_id, 'integration.connected', 'Slack connected successfully', 'Your Slack workspace is now connected.', 'integration', 'System', 'low', true, '{"seeded": true, "integration": "slack"}', v_now - INTERVAL '100 hours');

    INSERT INTO notifications (user_id, type, title, body, entity_type, sender_name, priority, is_read, metadata, created_at)
    VALUES (v_user_id, 'project.completed', 'Project completed: Website Refresh', 'Congratulations! The Website Refresh project is now complete.', 'project', 'Emily Davis', 'normal', true, '{"seeded": true}', v_now - INTERVAL '120 hours');

    INSERT INTO notifications (user_id, type, title, body, entity_type, sender_name, priority, is_read, metadata, created_at)
    VALUES (v_user_id, 'chat.artifact_ready', 'AI artifact ready: Market Analysis Report', 'Your requested analysis is ready to view.', 'chat', 'AI Assistant', 'normal', false, '{"seeded": true, "artifact_type": "report"}', v_now - INTERVAL '130 hours');

    -- ============ EARLIER (OLDER) ============
    INSERT INTO notifications (user_id, type, title, body, entity_type, sender_name, priority, is_read, metadata, created_at)
    VALUES (v_user_id, 'system.welcome', 'Welcome to Business OS!', 'Get started by creating your first project or exploring the dashboard.', 'system', 'System', 'low', true, '{"seeded": true}', v_now - INTERVAL '240 hours');

    INSERT INTO notifications (user_id, type, title, body, entity_type, sender_name, priority, is_read, metadata, created_at)
    VALUES (v_user_id, 'system.feature_announcement', 'New feature: AI-powered task suggestions', 'Try our new AI feature that suggests task breakdowns automatically.', 'system', 'System', 'low', true, '{"seeded": true, "feature": "ai_task_suggestions"}', v_now - INTERVAL '336 hours');

    INSERT INTO notifications (user_id, type, title, body, entity_type, sender_name, priority, is_read, metadata, created_at)
    VALUES (v_user_id, 'dailylog.reminder', 'Don''t forget your daily log', 'Take a moment to log your progress for today.', 'dailylog', 'System', 'normal', true, '{"seeded": true}', v_now - INTERVAL '360 hours');

    INSERT INTO notifications (user_id, type, title, body, entity_type, sender_name, priority, is_read, metadata, created_at)
    VALUES (v_user_id, 'team.role_changed', 'Your role updated: Project Lead', 'You''ve been promoted to Project Lead for the Engineering team.', 'team', 'Sarah Chen', 'normal', true, '{"seeded": true}', v_now - INTERVAL '400 hours');

    RAISE NOTICE 'Successfully seeded 22 notifications for user %', v_user_id;
END $$;

-- Show what was created
SELECT type, title, priority, is_read, created_at 
FROM notifications 
WHERE metadata->>'seeded' = 'true'
ORDER BY created_at DESC;
