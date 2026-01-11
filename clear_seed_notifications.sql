-- Clear all seeded notifications
-- Run this to remove test data when you're done testing

DELETE FROM notifications WHERE metadata->>'seeded' = 'true';

-- Show remaining count
SELECT COUNT(*) as remaining_notifications FROM notifications;
