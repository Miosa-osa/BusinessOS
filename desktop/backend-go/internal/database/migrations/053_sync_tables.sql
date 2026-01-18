-- Migration: 053_sync_tables.sql
-- Description: Sync tables for live bidirectional data synchronization
-- Created: 2026-01-18
-- Phase: Live Sync Implementation

-- =============================================================================
-- SYNCED CALENDAR EVENTS
-- Calendar events from Google Calendar, Microsoft 365, etc.
-- =============================================================================
CREATE TABLE IF NOT EXISTS synced_calendar_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id TEXT NOT NULL REFERENCES "user"(id) ON DELETE CASCADE,
    workspace_id UUID,

    -- Provider info
    provider VARCHAR(50) NOT NULL, -- 'google', 'microsoft'
    external_id VARCHAR(255) NOT NULL,

    -- Event data
    title TEXT,
    description TEXT,
    start_time TIMESTAMPTZ NOT NULL,
    end_time TIMESTAMPTZ NOT NULL,
    all_day BOOLEAN DEFAULT FALSE,
    location TEXT,
    attendees JSONB DEFAULT '[]'::jsonb, -- [{email, name, status}]
    organizer_email VARCHAR(255),
    meeting_link TEXT,
    recurring_event_id VARCHAR(255),

    -- Sync metadata
    raw_data JSONB,
    etag VARCHAR(255), -- For incremental sync
    sync_token VARCHAR(500), -- Provider sync token
    synced_at TIMESTAMPTZ DEFAULT NOW(),

    -- Timestamps
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    deleted_at TIMESTAMPTZ, -- Soft delete for sync

    UNIQUE(user_id, provider, external_id)
);

CREATE INDEX idx_synced_calendar_user_time ON synced_calendar_events(user_id, start_time);
CREATE INDEX idx_synced_calendar_provider ON synced_calendar_events(provider);
CREATE INDEX idx_synced_calendar_workspace ON synced_calendar_events(workspace_id);
CREATE INDEX idx_synced_calendar_synced ON synced_calendar_events(synced_at);

-- =============================================================================
-- SYNCED MESSAGES
-- Messages from Slack, Microsoft Teams, etc.
-- =============================================================================
CREATE TABLE IF NOT EXISTS synced_messages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id TEXT NOT NULL REFERENCES "user"(id) ON DELETE CASCADE,
    workspace_id UUID,

    -- Provider info
    provider VARCHAR(50) NOT NULL, -- 'slack', 'microsoft_teams'
    external_id VARCHAR(255) NOT NULL,

    -- Channel info
    channel_id VARCHAR(255),
    channel_name VARCHAR(255),
    channel_type VARCHAR(50), -- 'channel', 'dm', 'group'

    -- Message data
    sender_id VARCHAR(255),
    sender_name VARCHAR(255),
    sender_avatar_url TEXT,
    content TEXT,
    content_html TEXT,
    thread_id VARCHAR(255),
    is_thread_reply BOOLEAN DEFAULT FALSE,

    -- Attachments
    attachments JSONB DEFAULT '[]'::jsonb, -- [{url, name, type, size}]
    reactions JSONB DEFAULT '[]'::jsonb, -- [{name, count, users}]

    -- Sync metadata
    raw_data JSONB,
    sent_at TIMESTAMPTZ,
    synced_at TIMESTAMPTZ DEFAULT NOW(),

    UNIQUE(user_id, provider, external_id)
);

CREATE INDEX idx_synced_messages_user_channel ON synced_messages(user_id, channel_id);
CREATE INDEX idx_synced_messages_provider ON synced_messages(provider);
CREATE INDEX idx_synced_messages_thread ON synced_messages(thread_id);
CREATE INDEX idx_synced_messages_sent ON synced_messages(sent_at DESC);

-- =============================================================================
-- SYNCED TASKS
-- Tasks from Linear, ClickUp, Asana, etc.
-- =============================================================================
CREATE TABLE IF NOT EXISTS synced_tasks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id TEXT NOT NULL REFERENCES "user"(id) ON DELETE CASCADE,
    workspace_id UUID,

    -- Provider info
    provider VARCHAR(50) NOT NULL, -- 'linear', 'clickup', 'asana'
    external_id VARCHAR(255) NOT NULL,

    -- Task identifiers
    identifier VARCHAR(50), -- e.g., 'LIN-123', 'CLICK-456'
    url TEXT,

    -- Task data
    title TEXT NOT NULL,
    description TEXT,
    description_html TEXT,
    status VARCHAR(100),
    status_type VARCHAR(50), -- 'backlog', 'todo', 'in_progress', 'done', 'canceled'
    priority VARCHAR(50),
    priority_order INT,

    -- Assignment
    assignee_id VARCHAR(255),
    assignee_name VARCHAR(255),
    assignee_email VARCHAR(255),
    assignee_avatar_url TEXT,

    -- Project/Team
    project_id VARCHAR(255),
    project_name VARCHAR(255),
    team_id VARCHAR(255),
    team_name VARCHAR(255),

    -- Dates
    due_date TIMESTAMPTZ,
    started_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,

    -- Labels/Tags
    labels JSONB DEFAULT '[]'::jsonb, -- [{id, name, color}]

    -- Estimates
    estimate INT, -- Story points or similar
    time_estimate_minutes INT,
    time_spent_minutes INT,

    -- Parent/Sub-tasks
    parent_id VARCHAR(255),

    -- Sync metadata
    raw_data JSONB,
    synced_at TIMESTAMPTZ DEFAULT NOW(),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),

    UNIQUE(user_id, provider, external_id)
);

CREATE INDEX idx_synced_tasks_user_status ON synced_tasks(user_id, status_type);
CREATE INDEX idx_synced_tasks_provider ON synced_tasks(provider);
CREATE INDEX idx_synced_tasks_project ON synced_tasks(project_id);
CREATE INDEX idx_synced_tasks_assignee ON synced_tasks(assignee_id);
CREATE INDEX idx_synced_tasks_due ON synced_tasks(due_date);
CREATE INDEX idx_synced_tasks_identifier ON synced_tasks(identifier);

-- =============================================================================
-- SYNCED CONTACTS
-- Contacts/Leads from HubSpot, Salesforce, etc.
-- =============================================================================
CREATE TABLE IF NOT EXISTS synced_contacts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id TEXT NOT NULL REFERENCES "user"(id) ON DELETE CASCADE,
    workspace_id UUID,

    -- Provider info
    provider VARCHAR(50) NOT NULL, -- 'hubspot', 'salesforce'
    external_id VARCHAR(255) NOT NULL,

    -- Contact data
    email VARCHAR(255),
    first_name VARCHAR(255),
    last_name VARCHAR(255),
    full_name VARCHAR(500),
    phone VARCHAR(100),
    mobile_phone VARCHAR(100),

    -- Company info
    company VARCHAR(255),
    company_id VARCHAR(255),
    job_title VARCHAR(255),

    -- Lead/Contact status
    contact_type VARCHAR(50), -- 'lead', 'contact', 'customer'
    lifecycle_stage VARCHAR(100),
    lead_status VARCHAR(100),

    -- Social
    linkedin_url TEXT,
    twitter_handle VARCHAR(100),

    -- Address
    address_line1 TEXT,
    address_line2 TEXT,
    city VARCHAR(255),
    state VARCHAR(100),
    postal_code VARCHAR(50),
    country VARCHAR(100),

    -- Source
    lead_source VARCHAR(255),

    -- Custom properties
    properties JSONB DEFAULT '{}'::jsonb,

    -- Sync metadata
    raw_data JSONB,
    synced_at TIMESTAMPTZ DEFAULT NOW(),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),

    UNIQUE(user_id, provider, external_id)
);

CREATE INDEX idx_synced_contacts_user_email ON synced_contacts(user_id, email);
CREATE INDEX idx_synced_contacts_provider ON synced_contacts(provider);
CREATE INDEX idx_synced_contacts_company ON synced_contacts(company_id);
CREATE INDEX idx_synced_contacts_stage ON synced_contacts(lifecycle_stage);

-- =============================================================================
-- SYNCED FILES
-- Files from Google Drive, OneDrive, Dropbox, etc.
-- =============================================================================
CREATE TABLE IF NOT EXISTS synced_files (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id TEXT NOT NULL REFERENCES "user"(id) ON DELETE CASCADE,
    workspace_id UUID,

    -- Provider info
    provider VARCHAR(50) NOT NULL, -- 'google_drive', 'onedrive', 'dropbox'
    external_id VARCHAR(255) NOT NULL,

    -- File data
    name VARCHAR(500),
    mime_type VARCHAR(255),
    file_extension VARCHAR(50),
    size_bytes BIGINT,

    -- Folder structure
    parent_folder_id VARCHAR(255),
    parent_folder_name VARCHAR(500),
    path TEXT, -- Full path in provider

    -- Access
    web_url TEXT,
    download_url TEXT,
    thumbnail_url TEXT,

    -- Ownership
    owner_email VARCHAR(255),
    owner_name VARCHAR(255),
    shared_with JSONB DEFAULT '[]'::jsonb, -- [{email, role}]

    -- Dates
    provider_created_at TIMESTAMPTZ,
    provider_modified_at TIMESTAMPTZ,

    -- Sync metadata
    version VARCHAR(255),
    etag VARCHAR(255),
    raw_data JSONB,
    synced_at TIMESTAMPTZ DEFAULT NOW(),

    UNIQUE(user_id, provider, external_id)
);

CREATE INDEX idx_synced_files_user_folder ON synced_files(user_id, parent_folder_id);
CREATE INDEX idx_synced_files_provider ON synced_files(provider);
CREATE INDEX idx_synced_files_mime ON synced_files(mime_type);
CREATE INDEX idx_synced_files_name ON synced_files(name);

-- =============================================================================
-- SYNCED MEETINGS
-- Meeting recordings from Fathom, Fireflies, Otter, etc.
-- =============================================================================
CREATE TABLE IF NOT EXISTS synced_meetings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id TEXT NOT NULL REFERENCES "user"(id) ON DELETE CASCADE,
    workspace_id UUID,

    -- Provider info
    provider VARCHAR(50) NOT NULL, -- 'fathom', 'fireflies', 'otter'
    external_id VARCHAR(255) NOT NULL,

    -- Meeting data
    title VARCHAR(500),
    meeting_type VARCHAR(100), -- 'sales_call', 'team_sync', 'interview', etc.

    -- Timing
    start_time TIMESTAMPTZ,
    end_time TIMESTAMPTZ,
    duration_seconds INT,

    -- Participants
    participants JSONB DEFAULT '[]'::jsonb, -- [{name, email, role}]
    participant_count INT,
    organizer_name VARCHAR(255),
    organizer_email VARCHAR(255),

    -- Content
    transcript TEXT,
    transcript_segments JSONB, -- [{speaker, start_ms, end_ms, text}]
    summary TEXT,
    key_points JSONB DEFAULT '[]'::jsonb,
    action_items JSONB DEFAULT '[]'::jsonb, -- [{task, assignee, due_date}]
    questions JSONB DEFAULT '[]'::jsonb,

    -- Media
    recording_url TEXT,
    video_url TEXT,
    audio_url TEXT,

    -- Calendar link
    calendar_event_id VARCHAR(255),

    -- AI analysis
    sentiment VARCHAR(50),
    topics JSONB DEFAULT '[]'::jsonb,

    -- Sync metadata
    raw_data JSONB,
    synced_at TIMESTAMPTZ DEFAULT NOW(),
    created_at TIMESTAMPTZ DEFAULT NOW(),

    UNIQUE(user_id, provider, external_id)
);

CREATE INDEX idx_synced_meetings_user_time ON synced_meetings(user_id, start_time);
CREATE INDEX idx_synced_meetings_provider ON synced_meetings(provider);
CREATE INDEX idx_synced_meetings_type ON synced_meetings(meeting_type);
CREATE INDEX idx_synced_meetings_calendar ON synced_meetings(calendar_event_id);

-- =============================================================================
-- SYNCED NOTION PAGES
-- Pages and databases from Notion
-- =============================================================================
CREATE TABLE IF NOT EXISTS synced_notion_pages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id TEXT NOT NULL REFERENCES "user"(id) ON DELETE CASCADE,
    workspace_id UUID,

    -- Provider info
    external_id VARCHAR(255) NOT NULL,

    -- Page data
    title TEXT,
    icon VARCHAR(100), -- emoji or icon url
    cover_url TEXT,

    -- Hierarchy
    parent_type VARCHAR(50), -- 'database', 'page', 'workspace', 'block'
    parent_id VARCHAR(255),
    parent_name TEXT,

    -- Type
    object_type VARCHAR(50), -- 'page', 'database'
    is_database BOOLEAN DEFAULT FALSE,

    -- Content
    content_preview TEXT, -- First ~500 chars of content
    properties JSONB DEFAULT '{}'::jsonb, -- Database properties

    -- Access
    url TEXT,
    public_url TEXT,

    -- Dates
    provider_created_at TIMESTAMPTZ,
    provider_last_edited_at TIMESTAMPTZ,
    last_edited_by VARCHAR(255),
    created_by VARCHAR(255),

    -- Sync metadata
    raw_data JSONB,
    synced_at TIMESTAMPTZ DEFAULT NOW(),

    UNIQUE(user_id, external_id)
);

CREATE INDEX idx_synced_notion_user_parent ON synced_notion_pages(user_id, parent_id);
CREATE INDEX idx_synced_notion_type ON synced_notion_pages(object_type);
CREATE INDEX idx_synced_notion_database ON synced_notion_pages(is_database);

-- =============================================================================
-- WEBHOOK SUBSCRIPTIONS
-- Track active webhook registrations per user/provider
-- =============================================================================
CREATE TABLE IF NOT EXISTS webhook_subscriptions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id TEXT NOT NULL REFERENCES "user"(id) ON DELETE CASCADE,
    workspace_id UUID,

    -- Provider info
    provider VARCHAR(50) NOT NULL,
    resource_type VARCHAR(100), -- 'calendar', 'messages', 'issues', 'contacts'

    -- External subscription
    external_subscription_id VARCHAR(255),
    webhook_url TEXT,
    webhook_secret_encrypted TEXT, -- AES-256-GCM encrypted

    -- Events
    events JSONB DEFAULT '[]'::jsonb, -- Array of subscribed event types

    -- Expiration (Microsoft/Google require renewal)
    expires_at TIMESTAMPTZ,

    -- Status
    status VARCHAR(50) DEFAULT 'active', -- 'active', 'expired', 'failed', 'paused'
    last_event_at TIMESTAMPTZ,
    event_count INT DEFAULT 0,

    -- Error tracking
    consecutive_failures INT DEFAULT 0,
    last_error TEXT,
    last_error_at TIMESTAMPTZ,

    -- Timestamps
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),

    UNIQUE(user_id, provider, resource_type)
);

CREATE INDEX idx_webhook_subs_user ON webhook_subscriptions(user_id);
CREATE INDEX idx_webhook_subs_provider ON webhook_subscriptions(provider);
CREATE INDEX idx_webhook_subs_status ON webhook_subscriptions(status);
CREATE INDEX idx_webhook_subs_expires ON webhook_subscriptions(expires_at)
    WHERE expires_at IS NOT NULL AND status = 'active';

-- =============================================================================
-- SYNC TOKENS
-- Store incremental sync tokens per user/provider/resource
-- =============================================================================
CREATE TABLE IF NOT EXISTS sync_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id TEXT NOT NULL REFERENCES "user"(id) ON DELETE CASCADE,

    -- Provider/Resource
    provider VARCHAR(50) NOT NULL,
    resource_type VARCHAR(100) NOT NULL, -- 'calendar', 'drive', 'contacts'
    resource_id VARCHAR(255), -- Optional: specific resource ID

    -- Token
    sync_token TEXT NOT NULL, -- Provider's sync/delta token
    page_token TEXT, -- For paginated syncs

    -- Sync state
    last_sync_at TIMESTAMPTZ DEFAULT NOW(),
    full_sync_required BOOLEAN DEFAULT FALSE,

    -- Timestamps
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),

    UNIQUE(user_id, provider, resource_type, COALESCE(resource_id, ''))
);

CREATE INDEX idx_sync_tokens_user_provider ON sync_tokens(user_id, provider);
CREATE INDEX idx_sync_tokens_last_sync ON sync_tokens(last_sync_at);

-- =============================================================================
-- SYNC MAPPINGS
-- Map BusinessOS entities to external provider entities (for bidirectional sync)
-- =============================================================================
CREATE TABLE IF NOT EXISTS sync_mappings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id TEXT NOT NULL REFERENCES "user"(id) ON DELETE CASCADE,
    workspace_id UUID,

    -- Local entity
    local_entity_type VARCHAR(100) NOT NULL, -- 'task', 'project', 'contact'
    local_entity_id UUID NOT NULL,

    -- Remote entity
    provider VARCHAR(50) NOT NULL,
    external_entity_type VARCHAR(100) NOT NULL,
    external_entity_id VARCHAR(255) NOT NULL,

    -- Sync direction
    sync_direction VARCHAR(50) DEFAULT 'bidirectional', -- 'inbound', 'outbound', 'bidirectional'

    -- Version tracking
    local_version INT DEFAULT 0,
    remote_version INT DEFAULT 0,
    last_synced_at TIMESTAMPTZ DEFAULT NOW(),

    -- Timestamps
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),

    UNIQUE(local_entity_type, local_entity_id, provider)
);

CREATE INDEX idx_sync_mappings_local ON sync_mappings(local_entity_type, local_entity_id);
CREATE INDEX idx_sync_mappings_remote ON sync_mappings(provider, external_entity_id);
CREATE INDEX idx_sync_mappings_user ON sync_mappings(user_id);

-- =============================================================================
-- UPDATE TRIGGERS
-- =============================================================================

-- Generic updated_at trigger function (reuse if exists)
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Apply to all sync tables
CREATE TRIGGER trigger_synced_calendar_events_updated
    BEFORE UPDATE ON synced_calendar_events
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER trigger_synced_messages_updated
    BEFORE UPDATE ON synced_messages
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER trigger_synced_tasks_updated
    BEFORE UPDATE ON synced_tasks
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER trigger_synced_contacts_updated
    BEFORE UPDATE ON synced_contacts
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER trigger_synced_files_updated
    BEFORE UPDATE ON synced_files
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER trigger_webhook_subscriptions_updated
    BEFORE UPDATE ON webhook_subscriptions
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER trigger_sync_tokens_updated
    BEFORE UPDATE ON sync_tokens
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER trigger_sync_mappings_updated
    BEFORE UPDATE ON sync_mappings
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- =============================================================================
-- COMMENTS
-- =============================================================================
COMMENT ON TABLE synced_calendar_events IS 'Calendar events synced from Google Calendar, Microsoft 365';
COMMENT ON TABLE synced_messages IS 'Messages synced from Slack, Microsoft Teams';
COMMENT ON TABLE synced_tasks IS 'Tasks synced from Linear, ClickUp, Asana';
COMMENT ON TABLE synced_contacts IS 'Contacts/leads synced from HubSpot, Salesforce';
COMMENT ON TABLE synced_files IS 'Files synced from Google Drive, OneDrive, Dropbox';
COMMENT ON TABLE synced_meetings IS 'Meeting recordings synced from Fathom, Fireflies';
COMMENT ON TABLE synced_notion_pages IS 'Pages and databases synced from Notion';
COMMENT ON TABLE webhook_subscriptions IS 'Active webhook registrations for real-time sync';
COMMENT ON TABLE sync_tokens IS 'Incremental sync tokens for delta sync operations';
COMMENT ON TABLE sync_mappings IS 'Bidirectional mapping between local and external entities';
