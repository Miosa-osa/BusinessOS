-- Migration: Notifications System
-- Real-time notification system with batching, preferences, and multi-channel delivery

-- ===== NOTIFICATIONS TABLE =====
CREATE TABLE IF NOT EXISTS notifications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(255) NOT NULL,
    workspace_id UUID,

    -- Notification Content
    type VARCHAR(100) NOT NULL,
    title VARCHAR(255) NOT NULL,
    body TEXT,

    -- Reference to source entity
    entity_type VARCHAR(50),
    entity_id UUID,

    -- Sender info
    sender_id VARCHAR(255),
    sender_name VARCHAR(255),
    sender_avatar_url TEXT,

    -- Status
    is_read BOOLEAN DEFAULT FALSE,
    read_at TIMESTAMPTZ,

    -- Batching
    batch_id UUID,
    batch_count INTEGER DEFAULT 1,

    -- Delivery tracking
    channels_sent TEXT[] DEFAULT '{}',

    -- Priority: low, normal, high, urgent
    priority VARCHAR(20) DEFAULT 'normal',

    -- Flexible metadata
    metadata JSONB DEFAULT '{}',

    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Indexes for notifications
CREATE INDEX IF NOT EXISTS idx_notifications_user_unread
    ON notifications(user_id, is_read) WHERE is_read = FALSE;
CREATE INDEX IF NOT EXISTS idx_notifications_user_created
    ON notifications(user_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_notifications_entity
    ON notifications(entity_type, entity_id);
CREATE INDEX IF NOT EXISTS idx_notifications_batch
    ON notifications(batch_id) WHERE batch_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_notifications_type
    ON notifications(type);
CREATE INDEX IF NOT EXISTS idx_notifications_workspace
    ON notifications(workspace_id) WHERE workspace_id IS NOT NULL;

-- ===== NOTIFICATION PREFERENCES TABLE =====
CREATE TABLE IF NOT EXISTS notification_preferences (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(255) NOT NULL,
    workspace_id UUID,

    -- Global channel toggles
    email_enabled BOOLEAN DEFAULT TRUE,
    push_enabled BOOLEAN DEFAULT TRUE,
    in_app_enabled BOOLEAN DEFAULT TRUE,

    -- Per-type settings
    type_settings JSONB DEFAULT '{}',

    -- Quiet hours
    quiet_hours_enabled BOOLEAN DEFAULT FALSE,
    quiet_hours_start TIME,
    quiet_hours_end TIME,
    quiet_hours_timezone VARCHAR(50) DEFAULT 'UTC',

    -- Email digest preference
    email_digest_enabled BOOLEAN DEFAULT FALSE,
    email_digest_time TIME DEFAULT '09:00',
    email_digest_timezone VARCHAR(50) DEFAULT 'UTC',

    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),

    UNIQUE(user_id, workspace_id)
);

CREATE INDEX IF NOT EXISTS idx_notification_prefs_user
    ON notification_preferences(user_id);

-- ===== NOTIFICATION BATCHES TABLE =====
CREATE TABLE IF NOT EXISTS notification_batches (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(255) NOT NULL,
    batch_key VARCHAR(255) NOT NULL,
    type VARCHAR(100) NOT NULL,
    entity_type VARCHAR(50),
    entity_id UUID,
    pending_ids UUID[] DEFAULT '{}',
    pending_count INTEGER DEFAULT 0,
    first_at TIMESTAMPTZ DEFAULT NOW(),
    dispatch_at TIMESTAMPTZ NOT NULL,
    status VARCHAR(20) DEFAULT 'pending',

    UNIQUE(user_id, batch_key)
);

CREATE INDEX IF NOT EXISTS idx_notification_batches_dispatch
    ON notification_batches(dispatch_at) WHERE status = 'pending';
CREATE INDEX IF NOT EXISTS idx_notification_batches_user
    ON notification_batches(user_id);

-- ===== WEB PUSH SUBSCRIPTIONS TABLE =====
CREATE TABLE IF NOT EXISTS push_subscriptions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id TEXT NOT NULL REFERENCES "user"(id) ON DELETE CASCADE,
    endpoint TEXT NOT NULL UNIQUE,
    p256dh TEXT NOT NULL,
    auth TEXT NOT NULL,
    user_agent TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_push_subscriptions_user ON push_subscriptions(user_id);
CREATE INDEX IF NOT EXISTS idx_push_subscriptions_endpoint ON push_subscriptions(endpoint);

-- Comments
COMMENT ON TABLE notifications IS 'User notifications with multi-channel delivery support';
COMMENT ON TABLE notification_preferences IS 'Per-user notification channel and timing preferences';
COMMENT ON TABLE notification_batches IS 'Pending notification batches for spam reduction';
COMMENT ON TABLE push_subscriptions IS 'Web Push API subscriptions for browser notifications';
