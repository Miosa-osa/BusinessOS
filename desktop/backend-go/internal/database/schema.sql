-- BusinessOS Database Schema for sqlc
-- Note: Better Auth manages the "user" and "session" tables externally

-- Enum types (matching actual database - some use UPPERCASE values)
CREATE TYPE messagerole AS ENUM ('USER', 'ASSISTANT', 'SYSTEM', 'user', 'assistant', 'system');
CREATE TYPE artifacttype AS ENUM ('CODE', 'DOCUMENT', 'MARKDOWN', 'REACT', 'HTML', 'SVG');
CREATE TYPE contexttype AS ENUM ('PERSON', 'BUSINESS', 'PROJECT', 'CUSTOM', 'document', 'DOCUMENT');
CREATE TYPE projectstatus AS ENUM ('ACTIVE', 'PAUSED', 'COMPLETED', 'ARCHIVED');
CREATE TYPE projectpriority AS ENUM ('CRITICAL', 'HIGH', 'MEDIUM', 'LOW');
CREATE TYPE nodetype AS ENUM ('BUSINESS', 'PROJECT', 'LEARNING', 'OPERATIONAL');
CREATE TYPE nodehealth AS ENUM ('HEALTHY', 'NEEDS_ATTENTION', 'CRITICAL', 'NOT_STARTED');
CREATE TYPE taskstatus AS ENUM ('todo', 'in_progress', 'done', 'cancelled');
CREATE TYPE taskpriority AS ENUM ('critical', 'high', 'medium', 'low');
CREATE TYPE memberstatus AS ENUM ('AVAILABLE', 'BUSY', 'OVERLOADED', 'OOO');
CREATE TYPE clienttype AS ENUM ('company', 'individual');
CREATE TYPE clientstatus AS ENUM ('lead', 'prospect', 'active', 'inactive', 'churned');
CREATE TYPE interactiontype AS ENUM ('call', 'email', 'meeting', 'note');
CREATE TYPE dealstage AS ENUM ('qualification', 'proposal', 'negotiation', 'closed_won', 'closed_lost');

-- Contexts table (for documents, profiles)
CREATE TABLE contexts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    type contexttype DEFAULT 'CUSTOM',
    content TEXT,
    structured_data JSONB,
    system_prompt_template TEXT,
    blocks JSONB DEFAULT '[]',
    cover_image VARCHAR(500),
    icon VARCHAR(50),
    parent_id UUID REFERENCES contexts(id) ON DELETE SET NULL,
    is_template BOOLEAN DEFAULT FALSE,
    is_archived BOOLEAN DEFAULT FALSE,
    last_edited_at TIMESTAMP,
    word_count INTEGER DEFAULT 0,
    is_public BOOLEAN DEFAULT FALSE,
    share_id VARCHAR(32) UNIQUE,
    property_schema JSONB DEFAULT '[]',
    properties JSONB DEFAULT '{}',
    client_id UUID,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_contexts_user_id ON contexts(user_id);
CREATE INDEX idx_contexts_parent_id ON contexts(parent_id);
CREATE INDEX idx_contexts_is_archived ON contexts(is_archived);
CREATE INDEX idx_contexts_share_id ON contexts(share_id);

-- Conversations table
CREATE TABLE conversations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(255) NOT NULL,
    title VARCHAR(255) DEFAULT 'New Conversation',
    context_id UUID REFERENCES contexts(id) ON DELETE SET NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_conversations_user_id ON conversations(user_id);

-- Messages table
CREATE TABLE messages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    conversation_id UUID NOT NULL REFERENCES conversations(id) ON DELETE CASCADE,
    role messagerole NOT NULL,
    content TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    message_metadata JSONB
);

CREATE INDEX idx_messages_conversation_id ON messages(conversation_id);

-- Conversation tags
CREATE TABLE conversation_tags (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    conversation_id UUID NOT NULL REFERENCES conversations(id) ON DELETE CASCADE,
    tag VARCHAR(100) NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Projects table
CREATE TABLE projects (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    status projectstatus DEFAULT 'ACTIVE',
    priority projectpriority DEFAULT 'MEDIUM',
    client_name VARCHAR(255),
    client_id UUID REFERENCES clients(id) ON DELETE SET NULL,
    project_type VARCHAR(100) DEFAULT 'internal',
    project_metadata JSONB,
    -- Date tracking
    start_date DATE,
    due_date DATE,
    completed_at TIMESTAMP WITH TIME ZONE,
    -- Visibility/sharing
    visibility VARCHAR(20) DEFAULT 'private',
    owner_id VARCHAR(255),
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_projects_user_id ON projects(user_id);
CREATE INDEX idx_projects_client ON projects(client_id);
CREATE INDEX idx_projects_status ON projects(status);
CREATE INDEX idx_projects_due_date ON projects(due_date);

-- Project notes
CREATE TABLE project_notes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    content TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Project conversations (many-to-many)
CREATE TABLE project_conversations (
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    conversation_id UUID NOT NULL REFERENCES conversations(id) ON DELETE CASCADE,
    linked_at TIMESTAMP DEFAULT NOW(),
    PRIMARY KEY (project_id, conversation_id)
);

-- Artifacts table
CREATE TABLE artifacts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(255) NOT NULL,
    conversation_id UUID REFERENCES conversations(id) ON DELETE SET NULL,
    message_id UUID REFERENCES messages(id) ON DELETE SET NULL,
    project_id UUID REFERENCES projects(id) ON DELETE SET NULL,
    context_id UUID REFERENCES contexts(id) ON DELETE SET NULL,
    title VARCHAR(255) NOT NULL,
    type artifacttype NOT NULL,
    language VARCHAR(50),
    content TEXT NOT NULL,
    summary VARCHAR(500),
    version INTEGER DEFAULT 1,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_artifacts_user_id ON artifacts(user_id);
CREATE INDEX idx_artifacts_conversation_id ON artifacts(conversation_id);

-- Artifact versions
CREATE TABLE artifact_versions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    artifact_id UUID NOT NULL REFERENCES artifacts(id) ON DELETE CASCADE,
    version INTEGER NOT NULL,
    content TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Nodes table
CREATE TABLE nodes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(255) NOT NULL,
    parent_id UUID REFERENCES nodes(id) ON DELETE SET NULL,
    context_id UUID REFERENCES contexts(id) ON DELETE SET NULL,
    name VARCHAR(255) NOT NULL,
    type nodetype NOT NULL,
    health nodehealth DEFAULT 'NOT_STARTED',
    purpose TEXT,
    current_status TEXT,
    this_week_focus JSONB,
    decision_queue JSONB,
    delegation_ready JSONB,
    is_active BOOLEAN DEFAULT FALSE,
    is_archived BOOLEAN DEFAULT FALSE,
    sort_order INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_nodes_user_id ON nodes(user_id);

-- Node metrics
CREATE TABLE node_metrics (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    node_id UUID NOT NULL REFERENCES nodes(id) ON DELETE CASCADE,
    metric_name VARCHAR(255) NOT NULL,
    metric_value VARCHAR(500) NOT NULL,
    recorded_at TIMESTAMP DEFAULT NOW()
);

-- Team members table
CREATE TABLE team_members (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL,
    role VARCHAR(255) NOT NULL,
    avatar_url TEXT,
    status memberstatus DEFAULT 'AVAILABLE',
    capacity INTEGER DEFAULT 0,
    manager_id UUID REFERENCES team_members(id) ON DELETE SET NULL,
    skills JSONB,
    hourly_rate NUMERIC(10, 2),
    share_calendar BOOLEAN DEFAULT FALSE,
    calendar_user_id VARCHAR(255),
    joined_at TIMESTAMP DEFAULT NOW(),
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_team_members_user_id ON team_members(user_id);

-- Team member activities
CREATE TABLE team_member_activities (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    member_id UUID NOT NULL REFERENCES team_members(id) ON DELETE CASCADE,
    activity_type VARCHAR(100) NOT NULL,
    description TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Tasks table
CREATE TABLE tasks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(255) NOT NULL,
    title VARCHAR(500) NOT NULL,
    description TEXT,
    status taskstatus DEFAULT 'todo',
    priority taskpriority DEFAULT 'medium',
    due_date TIMESTAMP,
    completed_at TIMESTAMP,
    project_id UUID REFERENCES projects(id) ON DELETE SET NULL,
    assignee_id UUID REFERENCES team_members(id) ON DELETE SET NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_tasks_user_id ON tasks(user_id);

-- Focus items table
CREATE TABLE focus_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(255) NOT NULL,
    text VARCHAR(500) NOT NULL,
    completed BOOLEAN DEFAULT FALSE,
    focus_date TIMESTAMP DEFAULT NOW(),
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_focus_items_user_id ON focus_items(user_id);

-- Daily logs table
CREATE TABLE daily_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(255) NOT NULL,
    date DATE NOT NULL,
    content TEXT NOT NULL,
    transcription_source VARCHAR(50),
    extracted_actions JSONB,
    extracted_patterns JSONB,
    energy_level INTEGER,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(user_id, date)
);

CREATE INDEX idx_daily_logs_user_id ON daily_logs(user_id);
CREATE INDEX idx_daily_logs_date ON daily_logs(date);

-- User settings table
CREATE TABLE user_settings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(255) UNIQUE NOT NULL,
    default_model VARCHAR(100),
    email_notifications BOOLEAN DEFAULT TRUE,
    daily_summary BOOLEAN DEFAULT FALSE,
    theme VARCHAR(20) DEFAULT 'light',
    sidebar_collapsed BOOLEAN DEFAULT FALSE,
    share_analytics BOOLEAN DEFAULT TRUE,
    custom_settings JSONB,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_user_settings_user_id ON user_settings(user_id);

-- Clients table
CREATE TABLE clients (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    type clienttype DEFAULT 'company',
    email VARCHAR(255),
    phone VARCHAR(50),
    website VARCHAR(255),
    industry VARCHAR(100),
    company_size VARCHAR(50),
    address VARCHAR(255),
    city VARCHAR(100),
    state VARCHAR(100),
    zip_code VARCHAR(20),
    country VARCHAR(100),
    status clientstatus DEFAULT 'lead',
    source VARCHAR(100),
    assigned_to VARCHAR(255),
    lifetime_value NUMERIC(12, 2),
    tags JSONB DEFAULT '[]',
    custom_fields JSONB DEFAULT '{}',
    notes TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    last_contacted_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_clients_user_id ON clients(user_id);
CREATE INDEX ix_clients_user_status ON clients(user_id, status);
CREATE INDEX ix_clients_user_type ON clients(user_id, type);

-- Add FK from contexts to clients
ALTER TABLE contexts ADD CONSTRAINT fk_contexts_client_id FOREIGN KEY (client_id) REFERENCES clients(id) ON DELETE SET NULL;
CREATE INDEX idx_contexts_client_id ON contexts(client_id);

-- Client contacts
CREATE TABLE client_contacts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    client_id UUID NOT NULL REFERENCES clients(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255),
    phone VARCHAR(50),
    role VARCHAR(100),
    is_primary BOOLEAN DEFAULT FALSE,
    notes TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX ix_client_contacts_client ON client_contacts(client_id);

-- Client interactions
CREATE TABLE client_interactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    client_id UUID NOT NULL REFERENCES clients(id) ON DELETE CASCADE,
    contact_id UUID REFERENCES client_contacts(id) ON DELETE SET NULL,
    type interactiontype NOT NULL,
    subject VARCHAR(255) NOT NULL,
    description TEXT,
    outcome VARCHAR(255),
    occurred_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX ix_client_interactions_client ON client_interactions(client_id);
CREATE INDEX ix_client_interactions_occurred ON client_interactions(occurred_at);

-- Client deals
CREATE TABLE client_deals (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    client_id UUID NOT NULL REFERENCES clients(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    value NUMERIC(12, 2) DEFAULT 0,
    stage dealstage DEFAULT 'qualification',
    probability INTEGER DEFAULT 0,
    expected_close_date DATE,
    notes TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    closed_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX ix_client_deals_client ON client_deals(client_id);
CREATE INDEX ix_client_deals_stage ON client_deals(stage);

-- Google OAuth tokens for calendar integration
CREATE TABLE google_oauth_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(255) UNIQUE NOT NULL,
    access_token TEXT NOT NULL,
    refresh_token TEXT NOT NULL,
    token_type VARCHAR(50) DEFAULT 'Bearer',
    expiry TIMESTAMP WITH TIME ZONE NOT NULL,
    scopes TEXT[],
    google_email VARCHAR(255),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_google_oauth_user_id ON google_oauth_tokens(user_id);

-- Slack OAuth tokens for workspace integration
CREATE TABLE slack_oauth_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(255) UNIQUE NOT NULL,
    -- Workspace info
    workspace_id VARCHAR(255) NOT NULL,
    workspace_name VARCHAR(255),
    -- Tokens - Slack provides both bot and user tokens
    bot_token TEXT NOT NULL,
    user_token TEXT,
    -- Token metadata
    bot_user_id VARCHAR(255),
    authed_user_id VARCHAR(255),
    -- Scopes granted
    bot_scopes TEXT[],
    user_scopes TEXT[],
    -- Webhook URL (if configured)
    incoming_webhook_url TEXT,
    incoming_webhook_channel VARCHAR(255),
    -- Metadata
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_slack_oauth_user_id ON slack_oauth_tokens(user_id);
CREATE INDEX idx_slack_oauth_workspace ON slack_oauth_tokens(workspace_id);

-- Notion OAuth tokens for workspace integration
CREATE TABLE notion_oauth_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(255) UNIQUE NOT NULL,
    -- Workspace info
    workspace_id VARCHAR(255) NOT NULL,
    workspace_name VARCHAR(255),
    workspace_icon TEXT,
    -- Token - Notion provides a single access token (no refresh needed)
    access_token TEXT NOT NULL,
    bot_id VARCHAR(255),
    -- Owner info
    owner_type VARCHAR(50), -- 'user' or 'workspace'
    owner_user_id VARCHAR(255),
    owner_user_name VARCHAR(255),
    owner_user_email VARCHAR(255),
    -- Metadata
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_notion_oauth_user_id ON notion_oauth_tokens(user_id);
CREATE INDEX idx_notion_oauth_workspace ON notion_oauth_tokens(workspace_id);

-- Meeting types enum
CREATE TYPE meetingtype AS ENUM (
    'team', 'sales', 'onboarding', 'kickoff', 'implementation',
    'standup', 'retrospective', 'planning', 'review', 'one_on_one',
    'client', 'internal', 'external', 'other'
);

-- Calendar events cache with meeting management
CREATE TABLE calendar_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(255) NOT NULL,
    google_event_id VARCHAR(255),
    calendar_id VARCHAR(255) DEFAULT 'primary',
    title VARCHAR(500),
    description TEXT,
    start_time TIMESTAMP WITH TIME ZONE NOT NULL,
    end_time TIMESTAMP WITH TIME ZONE NOT NULL,
    all_day BOOLEAN DEFAULT FALSE,
    location VARCHAR(500),
    attendees JSONB DEFAULT '[]',
    status VARCHAR(50) DEFAULT 'confirmed',
    visibility VARCHAR(50) DEFAULT 'default',
    html_link TEXT,
    source VARCHAR(50) DEFAULT 'google',

    -- Meeting management fields
    meeting_type meetingtype DEFAULT 'other',
    context_id UUID REFERENCES contexts(id) ON DELETE SET NULL,
    project_id UUID REFERENCES projects(id) ON DELETE SET NULL,
    client_id UUID REFERENCES clients(id) ON DELETE SET NULL,

    -- Recording and external links
    recording_url TEXT,
    meeting_link TEXT,
    external_links JSONB DEFAULT '[]',

    -- Meeting notes and follow-ups
    meeting_notes TEXT,
    action_items JSONB DEFAULT '[]',

    -- Metadata
    synced_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(user_id, google_event_id)
);

CREATE INDEX idx_calendar_events_user_id ON calendar_events(user_id);
CREATE INDEX idx_calendar_events_time ON calendar_events(user_id, start_time, end_time);
CREATE INDEX idx_calendar_events_source ON calendar_events(source);
CREATE INDEX idx_calendar_events_type ON calendar_events(meeting_type);
CREATE INDEX idx_calendar_events_context ON calendar_events(context_id);
CREATE INDEX idx_calendar_events_project ON calendar_events(project_id);
CREATE INDEX idx_calendar_events_client ON calendar_events(client_id);

-- ===== USAGE ANALYTICS TABLES =====

-- AI usage tracking (per request)
CREATE TABLE ai_usage_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(255) NOT NULL,
    conversation_id UUID REFERENCES conversations(id) ON DELETE SET NULL,

    -- Provider and model info
    provider VARCHAR(50) NOT NULL,
    model VARCHAR(100) NOT NULL,

    -- Token usage
    input_tokens INTEGER DEFAULT 0,
    output_tokens INTEGER DEFAULT 0,
    total_tokens INTEGER DEFAULT 0,

    -- Agent tracking
    agent_name VARCHAR(100),
    delegated_to VARCHAR(100),
    parent_request_id UUID REFERENCES ai_usage_logs(id) ON DELETE SET NULL,

    -- Request context
    request_type VARCHAR(50),  -- 'chat', 'completion', 'extract', 'analyze'
    context_ids UUID[],
    node_id UUID REFERENCES nodes(id) ON DELETE SET NULL,
    project_id UUID REFERENCES projects(id) ON DELETE SET NULL,

    -- Timing
    duration_ms INTEGER DEFAULT 0,
    started_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    completed_at TIMESTAMP WITH TIME ZONE,

    -- Cost tracking (optional)
    estimated_cost NUMERIC(10, 6),

    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_ai_usage_user_id ON ai_usage_logs(user_id);
CREATE INDEX idx_ai_usage_conversation ON ai_usage_logs(conversation_id);
CREATE INDEX idx_ai_usage_provider ON ai_usage_logs(provider);
CREATE INDEX idx_ai_usage_model ON ai_usage_logs(model);
CREATE INDEX idx_ai_usage_agent ON ai_usage_logs(agent_name);
CREATE INDEX idx_ai_usage_date ON ai_usage_logs(started_at);

-- MCP tool usage tracking
CREATE TABLE mcp_usage_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(255) NOT NULL,

    -- Tool info
    tool_name VARCHAR(255) NOT NULL,
    server_name VARCHAR(255),

    -- Request details
    input_params JSONB,
    output_result JSONB,
    success BOOLEAN DEFAULT TRUE,
    error_message TEXT,

    -- Timing
    duration_ms INTEGER DEFAULT 0,

    -- Context
    conversation_id UUID REFERENCES conversations(id) ON DELETE SET NULL,
    ai_request_id UUID REFERENCES ai_usage_logs(id) ON DELETE SET NULL,

    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_mcp_usage_user_id ON mcp_usage_logs(user_id);
CREATE INDEX idx_mcp_usage_tool ON mcp_usage_logs(tool_name);
CREATE INDEX idx_mcp_usage_date ON mcp_usage_logs(created_at);

-- Daily usage summary (aggregated)
CREATE TABLE usage_daily_summary (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(255) NOT NULL,
    date DATE NOT NULL,

    -- AI usage totals
    ai_requests INTEGER DEFAULT 0,
    ai_input_tokens INTEGER DEFAULT 0,
    ai_output_tokens INTEGER DEFAULT 0,
    ai_total_tokens INTEGER DEFAULT 0,
    ai_estimated_cost NUMERIC(10, 4) DEFAULT 0,

    -- Breakdown by provider
    provider_breakdown JSONB DEFAULT '{}',

    -- Breakdown by model
    model_breakdown JSONB DEFAULT '{}',

    -- Breakdown by agent
    agent_breakdown JSONB DEFAULT '{}',

    -- MCP usage totals
    mcp_requests INTEGER DEFAULT 0,
    mcp_tool_breakdown JSONB DEFAULT '{}',

    -- System usage
    conversations_created INTEGER DEFAULT 0,
    messages_sent INTEGER DEFAULT 0,
    artifacts_created INTEGER DEFAULT 0,
    documents_created INTEGER DEFAULT 0,

    -- Context usage
    contexts_accessed UUID[],
    nodes_accessed UUID[],
    projects_accessed UUID[],

    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(user_id, date)
);

CREATE INDEX idx_usage_summary_user_id ON usage_daily_summary(user_id);
CREATE INDEX idx_usage_summary_date ON usage_daily_summary(date);

-- System event logs (general activity tracking)
CREATE TABLE system_event_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(255) NOT NULL,

    -- Event details
    event_type VARCHAR(100) NOT NULL,  -- 'page_view', 'action', 'api_call'
    event_name VARCHAR(255) NOT NULL,
    event_data JSONB,

    -- Context
    module VARCHAR(100),  -- 'chat', 'calendar', 'clients', 'nodes', etc.
    resource_type VARCHAR(100),
    resource_id UUID,

    -- Session info
    session_id VARCHAR(255),

    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_system_events_user_id ON system_event_logs(user_id);
CREATE INDEX idx_system_events_type ON system_event_logs(event_type);
CREATE INDEX idx_system_events_module ON system_event_logs(module);
CREATE INDEX idx_system_events_date ON system_event_logs(created_at);

-- ===== CUSTOM SLASH COMMANDS =====

-- User custom commands for AI chat
CREATE TABLE user_commands (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(255) NOT NULL,
    name VARCHAR(50) NOT NULL,           -- e.g., "weekly-report" (the slash command name)
    display_name VARCHAR(100) NOT NULL,  -- e.g., "Weekly Report" (shown in UI)
    description TEXT,                    -- Description of what the command does
    icon VARCHAR(10),                    -- emoji icon
    system_prompt TEXT NOT NULL,         -- Custom prompt template
    context_sources TEXT[] DEFAULT '{}', -- What context to load: documents, projects, clients, tasks, artifacts
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(user_id, name)
);

CREATE INDEX idx_user_commands_user_id ON user_commands(user_id);
CREATE INDEX idx_user_commands_name ON user_commands(user_id, name);

-- ===== VOICE NOTES =====

-- Voice transcription history with stats
CREATE TABLE voice_notes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(255) NOT NULL,
    transcript TEXT NOT NULL,
    duration_seconds INTEGER NOT NULL,
    word_count INTEGER NOT NULL,
    words_per_minute NUMERIC(10, 2),
    language VARCHAR(10) DEFAULT 'en',
    audio_file_path VARCHAR(500),
    context_id UUID REFERENCES contexts(id) ON DELETE SET NULL,
    project_id UUID REFERENCES projects(id) ON DELETE SET NULL,
    conversation_id UUID REFERENCES conversations(id) ON DELETE SET NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_voice_notes_user_id ON voice_notes(user_id);
CREATE INDEX idx_voice_notes_date ON voice_notes(created_at);
CREATE INDEX idx_voice_notes_context ON voice_notes(context_id);
CREATE INDEX idx_voice_notes_project ON voice_notes(project_id);

-- ===== PROJECT MANAGEMENT ENHANCEMENTS =====

-- Project role type for team assignment
CREATE TYPE projectrole AS ENUM ('owner', 'admin', 'member', 'viewer');

-- Project members (team assignment)
CREATE TABLE project_members (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    user_id VARCHAR(255) NOT NULL,
    team_member_id UUID REFERENCES team_members(id) ON DELETE CASCADE,
    role projectrole DEFAULT 'member',
    assigned_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    assigned_by VARCHAR(255),
    UNIQUE(project_id, user_id),
    UNIQUE(project_id, team_member_id)
);

CREATE INDEX idx_project_members_project ON project_members(project_id);
CREATE INDEX idx_project_members_user ON project_members(user_id);
CREATE INDEX idx_project_members_team_member ON project_members(team_member_id);

-- Project tags (user-defined labels)
CREATE TABLE project_tags (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(255) NOT NULL,
    name VARCHAR(50) NOT NULL,
    color VARCHAR(7) DEFAULT '#6366f1',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(user_id, name)
);

CREATE INDEX idx_project_tags_user ON project_tags(user_id);

-- Project tag assignments (many-to-many)
CREATE TABLE project_tag_assignments (
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    tag_id UUID NOT NULL REFERENCES project_tags(id) ON DELETE CASCADE,
    assigned_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    PRIMARY KEY (project_id, tag_id)
);

CREATE INDEX idx_tag_assignments_project ON project_tag_assignments(project_id);
CREATE INDEX idx_tag_assignments_tag ON project_tag_assignments(tag_id);

-- Project documents (linking projects to contexts/documents)
CREATE TABLE project_documents (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    document_id UUID NOT NULL REFERENCES contexts(id) ON DELETE CASCADE,
    linked_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    linked_by VARCHAR(255),
    UNIQUE(project_id, document_id)
);

CREATE INDEX idx_project_docs_project ON project_documents(project_id);
CREATE INDEX idx_project_docs_document ON project_documents(document_id);

-- Project templates
CREATE TABLE project_templates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    default_status projectstatus DEFAULT 'ACTIVE',
    default_priority projectpriority DEFAULT 'MEDIUM',
    template_data JSONB DEFAULT '{}',
    is_public BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_templates_user ON project_templates(user_id);
CREATE INDEX idx_templates_public ON project_templates(is_public) WHERE is_public = TRUE;
