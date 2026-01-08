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

-- User table (managed by Better Auth, but needed for SQLC joins)
-- This is a reference schema only - the actual table is created by Better Auth
CREATE TABLE "user" (
    id VARCHAR(255) PRIMARY KEY,
    name VARCHAR(255),
    email VARCHAR(255) NOT NULL,
    email_verified BOOLEAN DEFAULT FALSE,
    image VARCHAR(500),
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

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

-- Node to Project links (many-to-many)
CREATE TABLE node_projects (
    node_id UUID NOT NULL REFERENCES nodes(id) ON DELETE CASCADE,
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    linked_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    linked_by VARCHAR(255),
    PRIMARY KEY (node_id, project_id)
);

CREATE INDEX idx_node_projects_node ON node_projects(node_id);
CREATE INDEX idx_node_projects_project ON node_projects(project_id);

-- Node to Context links (many-to-many)
CREATE TABLE node_contexts (
    node_id UUID NOT NULL REFERENCES nodes(id) ON DELETE CASCADE,
    context_id UUID NOT NULL REFERENCES contexts(id) ON DELETE CASCADE,
    linked_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    linked_by VARCHAR(255),
    PRIMARY KEY (node_id, context_id)
);

CREATE INDEX idx_node_contexts_node ON node_contexts(node_id);
CREATE INDEX idx_node_contexts_context ON node_contexts(context_id);

-- Node to Conversation links (many-to-many)
CREATE TABLE node_conversations (
    node_id UUID NOT NULL REFERENCES nodes(id) ON DELETE CASCADE,
    conversation_id UUID NOT NULL REFERENCES conversations(id) ON DELETE CASCADE,
    linked_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    linked_by VARCHAR(255),
    PRIMARY KEY (node_id, conversation_id)
);

CREATE INDEX idx_node_conversations_node ON node_conversations(node_id);
CREATE INDEX idx_node_conversations_conversation ON node_conversations(conversation_id);

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
    start_date TIMESTAMP,
    completed_at TIMESTAMP,
    project_id UUID REFERENCES projects(id) ON DELETE SET NULL,
    assignee_id UUID REFERENCES team_members(id) ON DELETE SET NULL,
    parent_task_id UUID REFERENCES tasks(id) ON DELETE CASCADE,
    custom_status_id UUID,
    position INT DEFAULT 0,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_tasks_user_id ON tasks(user_id);
CREATE INDEX idx_tasks_parent ON tasks(parent_task_id);
CREATE INDEX idx_tasks_position ON tasks(user_id, position);

-- Project custom statuses
CREATE TABLE project_statuses (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    color VARCHAR(7) DEFAULT '#6B7280',
    position INT DEFAULT 0,
    is_done_state BOOLEAN DEFAULT FALSE,
    is_default BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(project_id, name)
);

CREATE INDEX idx_project_statuses_project ON project_statuses(project_id);

-- Add FK from tasks to project_statuses
ALTER TABLE tasks ADD CONSTRAINT fk_tasks_custom_status FOREIGN KEY (custom_status_id) REFERENCES project_statuses(id) ON DELETE SET NULL;

-- Task assignees (many-to-many)
CREATE TABLE task_assignees (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    task_id UUID NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    team_member_id UUID NOT NULL REFERENCES team_members(id) ON DELETE CASCADE,
    role VARCHAR(50) DEFAULT 'assignee',
    assigned_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    assigned_by VARCHAR(255),
    UNIQUE(task_id, team_member_id)
);

CREATE INDEX idx_task_assignees_task ON task_assignees(task_id);
CREATE INDEX idx_task_assignees_member ON task_assignees(team_member_id);

-- Task dependencies
CREATE TYPE dependencytype AS ENUM ('finish_to_start', 'start_to_start', 'finish_to_finish', 'start_to_finish');

CREATE TABLE task_dependencies (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    predecessor_id UUID NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    successor_id UUID NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    dependency_type dependencytype DEFAULT 'finish_to_start',
    lag_days INT DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(predecessor_id, successor_id)
);

CREATE INDEX idx_task_deps_predecessor ON task_dependencies(predecessor_id);
CREATE INDEX idx_task_deps_successor ON task_dependencies(successor_id);

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
    -- Thinking/COT settings
    thinking_enabled BOOLEAN DEFAULT false,
    thinking_show_in_ui BOOLEAN DEFAULT true,
    thinking_save_traces BOOLEAN DEFAULT true,
    thinking_default_template_id UUID,
    thinking_max_tokens INT DEFAULT 4096,
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
    thinking_tokens INTEGER DEFAULT 0,  -- COT reasoning tokens (tracked separately)

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
    ai_thinking_tokens BIGINT DEFAULT 0,  -- COT reasoning tokens
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

-- ===== CUSTOM AGENTS =====

-- User-defined custom agents with custom system prompts and configurations
CREATE TABLE custom_agents (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(255) NOT NULL,

    -- Agent Identity
    name VARCHAR(50) NOT NULL,              -- e.g., "code-reviewer" (internal name, lowercase)
    display_name VARCHAR(100) NOT NULL,     -- e.g., "Code Reviewer" (shown in UI)
    description TEXT,                       -- What the agent does
    avatar VARCHAR(50),                     -- emoji or icon identifier

    -- Agent Configuration
    system_prompt TEXT NOT NULL,            -- Base system prompt for the agent
    model_preference VARCHAR(100),          -- Preferred model (e.g., "claude-3-opus")
    temperature DECIMAL(3,2) DEFAULT 0.7,   -- Default temperature
    max_tokens INTEGER DEFAULT 4096,        -- Default max tokens

    -- Capabilities
    capabilities TEXT[] DEFAULT '{}',       -- e.g., ["code_review", "analysis", "writing"]
    tools_enabled TEXT[] DEFAULT '{}',      -- Which tools the agent can use
    context_sources TEXT[] DEFAULT '{}',    -- What context to auto-load: documents, projects, etc.

    -- Behavior Settings
    thinking_enabled BOOLEAN DEFAULT FALSE,  -- Enable COT for this agent
    streaming_enabled BOOLEAN DEFAULT TRUE,  -- Enable streaming responses

    -- Agent Type/Category
    category VARCHAR(50) DEFAULT 'general', -- general, coding, writing, analysis, business, custom
    is_public BOOLEAN DEFAULT FALSE,        -- Whether to share with team (future)

    -- Usage & Status
    is_active BOOLEAN DEFAULT TRUE,
    times_used INTEGER DEFAULT 0,
    last_used_at TIMESTAMP WITH TIME ZONE,

    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),

    UNIQUE(user_id, name)
);

CREATE INDEX idx_custom_agents_user_id ON custom_agents(user_id);
CREATE INDEX idx_custom_agents_name ON custom_agents(user_id, name);
CREATE INDEX idx_custom_agents_category ON custom_agents(category);

-- Agent presets (built-in templates users can copy)
CREATE TABLE agent_presets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(50) NOT NULL UNIQUE,
    display_name VARCHAR(100) NOT NULL,
    description TEXT,
    avatar VARCHAR(50),
    system_prompt TEXT NOT NULL,
    model_preference VARCHAR(100),
    temperature DECIMAL(3,2) DEFAULT 0.7,
    max_tokens INTEGER DEFAULT 4096,
    capabilities TEXT[] DEFAULT '{}',
    tools_enabled TEXT[] DEFAULT '{}',
    context_sources TEXT[] DEFAULT '{}',
    thinking_enabled BOOLEAN DEFAULT FALSE,
    category VARCHAR(50) DEFAULT 'general',
    times_copied INTEGER DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

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

-- ===== CHAIN OF THOUGHT (COT) THINKING SYSTEM =====

-- Thinking type enum
CREATE TYPE thinkingtype AS ENUM ('analysis', 'planning', 'reflection', 'tool_use', 'reasoning', 'evaluation');

-- Thinking/reasoning tracking
CREATE TABLE thinking_traces (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(255) NOT NULL,
    conversation_id UUID REFERENCES conversations(id) ON DELETE CASCADE,
    message_id UUID REFERENCES messages(id) ON DELETE CASCADE,

    -- Thinking content
    thinking_content TEXT NOT NULL,
    thinking_type thinkingtype DEFAULT 'reasoning',
    step_number INT DEFAULT 1,

    -- Timing
    started_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    completed_at TIMESTAMP WITH TIME ZONE,
    duration_ms INT,

    -- Token tracking
    thinking_tokens INT DEFAULT 0,

    -- Metadata
    model_used VARCHAR(100),
    reasoning_template_id UUID,
    metadata JSONB DEFAULT '{}',

    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_thinking_traces_user ON thinking_traces(user_id);
CREATE INDEX idx_thinking_traces_conversation ON thinking_traces(conversation_id);
CREATE INDEX idx_thinking_traces_message ON thinking_traces(message_id);
CREATE INDEX idx_thinking_traces_template ON thinking_traces(reasoning_template_id);

-- Custom reasoning templates/systems
CREATE TABLE reasoning_templates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(255) NOT NULL,

    name VARCHAR(255) NOT NULL,
    description TEXT,

    -- Template configuration
    system_prompt TEXT,
    thinking_instruction TEXT,
    output_format VARCHAR(50) DEFAULT 'streaming',

    -- Options
    show_thinking BOOLEAN DEFAULT true,
    save_thinking BOOLEAN DEFAULT true,
    max_thinking_tokens INT DEFAULT 4096,

    -- Usage tracking
    times_used INT DEFAULT 0,

    is_default BOOLEAN DEFAULT false,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_reasoning_templates_user ON reasoning_templates(user_id);
CREATE INDEX idx_reasoning_templates_default ON reasoning_templates(user_id, is_default) WHERE is_default = true;

-- ===== FOCUS CONFIGURATIONS SYSTEM =====

CREATE TABLE focus_mode_templates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(50) NOT NULL UNIQUE,
    display_name VARCHAR(100) NOT NULL,
    description TEXT,
    icon VARCHAR(50),
    default_model VARCHAR(100),
    temperature DECIMAL(3,2) DEFAULT 0.7,
    max_tokens INTEGER DEFAULT 4096,
    output_style VARCHAR(50) DEFAULT 'balanced',
    response_format VARCHAR(50) DEFAULT 'markdown',
    max_response_length INTEGER,
    require_sources BOOLEAN DEFAULT false,
    auto_search BOOLEAN DEFAULT false,
    search_depth VARCHAR(20) DEFAULT 'quick',
    kb_context_limit INTEGER DEFAULT 5,
    include_history_count INTEGER DEFAULT 10,
    thinking_enabled BOOLEAN DEFAULT false,
    thinking_style VARCHAR(50),
    system_prompt_prefix TEXT,
    system_prompt_suffix TEXT,
    sort_order INTEGER DEFAULT 0,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE focus_configurations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(255) NOT NULL,
    template_id UUID REFERENCES focus_mode_templates(id) ON DELETE CASCADE,
    custom_name VARCHAR(100),
    temperature DECIMAL(3,2),
    max_tokens INTEGER,
    output_style VARCHAR(50),
    response_format VARCHAR(50),
    max_response_length INTEGER,
    require_sources BOOLEAN,
    auto_search BOOLEAN,
    search_depth VARCHAR(20),
    kb_context_limit INTEGER,
    include_history_count INTEGER,
    thinking_enabled BOOLEAN,
    thinking_style VARCHAR(50),
    custom_system_prompt TEXT,
    preferred_model VARCHAR(100),
    auto_load_kb_categories TEXT[],
    keyboard_shortcut VARCHAR(20),
    is_favorite BOOLEAN DEFAULT false,
    use_count INTEGER DEFAULT 0,
    last_used_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(user_id, template_id)
);

CREATE TABLE focus_context_presets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(255) NOT NULL,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    kb_artifact_ids UUID[],
    kb_categories TEXT[],
    project_ids UUID[],
    default_search_queries TEXT[],
    search_domains TEXT[],
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE focus_configuration_presets (
    focus_config_id UUID REFERENCES focus_configurations(id) ON DELETE CASCADE,
    preset_id UUID REFERENCES focus_context_presets(id) ON DELETE CASCADE,
    sort_order INTEGER DEFAULT 0,
    PRIMARY KEY (focus_config_id, preset_id)
);

CREATE INDEX idx_focus_configurations_user ON focus_configurations(user_id);
CREATE INDEX idx_focus_configurations_template ON focus_configurations(template_id);
CREATE INDEX idx_focus_context_presets_user ON focus_context_presets(user_id);

-- ===== WEB SEARCH CACHE =====

CREATE TABLE web_search_results (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    query_hash VARCHAR(64) NOT NULL,
    original_query TEXT NOT NULL,
    optimized_query TEXT,
    user_id VARCHAR(255),
    conversation_id UUID,
    results JSONB NOT NULL DEFAULT '[]',
    result_count INTEGER DEFAULT 0,
    provider VARCHAR(50) DEFAULT 'duckduckgo',
    search_time_ms FLOAT,
    expires_at TIMESTAMPTZ NOT NULL,
    hit_count INTEGER DEFAULT 0,
    last_hit_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_web_search_query_hash ON web_search_results(query_hash);
CREATE INDEX idx_web_search_user ON web_search_results(user_id) WHERE user_id IS NOT NULL;
CREATE INDEX idx_web_search_conversation ON web_search_results(conversation_id) WHERE conversation_id IS NOT NULL;
CREATE INDEX idx_web_search_expires ON web_search_results(expires_at);
CREATE INDEX idx_web_search_lookup ON web_search_results(query_hash, expires_at);

-- ===== NOTIFICATIONS SYSTEM =====

-- Notifications table
CREATE TABLE notifications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(255) NOT NULL,
    workspace_id UUID,

    type VARCHAR(100) NOT NULL,
    title VARCHAR(255) NOT NULL,
    body TEXT,

    entity_type VARCHAR(50),
    entity_id UUID,

    sender_id VARCHAR(255),
    sender_name VARCHAR(255),
    sender_avatar_url TEXT,

    is_read BOOLEAN DEFAULT FALSE,
    read_at TIMESTAMPTZ,

    batch_id UUID,
    batch_count INTEGER DEFAULT 1,

    channels_sent TEXT[] DEFAULT '{}',
    priority VARCHAR(20) DEFAULT 'normal',
    metadata JSONB DEFAULT '{}',

    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_notifications_user_unread ON notifications(user_id, is_read) WHERE is_read = FALSE;
CREATE INDEX idx_notifications_user_created ON notifications(user_id, created_at DESC);
CREATE INDEX idx_notifications_entity ON notifications(entity_type, entity_id);
CREATE INDEX idx_notifications_batch ON notifications(batch_id) WHERE batch_id IS NOT NULL;
CREATE INDEX idx_notifications_type ON notifications(type);

-- Notification preferences table
CREATE TABLE notification_preferences (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(255) NOT NULL,
    workspace_id UUID,

    email_enabled BOOLEAN DEFAULT TRUE,
    push_enabled BOOLEAN DEFAULT TRUE,
    in_app_enabled BOOLEAN DEFAULT TRUE,

    type_settings JSONB DEFAULT '{}',

    quiet_hours_enabled BOOLEAN DEFAULT FALSE,
    quiet_hours_start TIME,
    quiet_hours_end TIME,
    quiet_hours_timezone VARCHAR(50) DEFAULT 'UTC',

    email_digest_enabled BOOLEAN DEFAULT FALSE,
    email_digest_time TIME DEFAULT '09:00',
    email_digest_timezone VARCHAR(50) DEFAULT 'UTC',

    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),

    UNIQUE(user_id, workspace_id)
);

CREATE INDEX idx_notification_prefs_user ON notification_preferences(user_id);

-- Notification batches table
CREATE TABLE notification_batches (
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

CREATE INDEX idx_notification_batches_dispatch ON notification_batches(dispatch_at) WHERE status = 'pending';
CREATE INDEX idx_notification_batches_user ON notification_batches(user_id);

-- Web Push subscriptions for browser notifications
CREATE TABLE push_subscriptions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(255) NOT NULL,
    endpoint TEXT NOT NULL UNIQUE,
    p256dh TEXT NOT NULL,
    auth TEXT NOT NULL,
    user_agent TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_push_subscriptions_user ON push_subscriptions(user_id);

-- ============================================================================
-- DASHBOARD SYSTEM
-- ============================================================================

-- User Dashboards table
CREATE TABLE user_dashboards (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(255) NOT NULL,
    workspace_id UUID REFERENCES workspaces(id) ON DELETE CASCADE,
    
    name VARCHAR(100) NOT NULL,
    description TEXT,
    is_default BOOLEAN DEFAULT FALSE,
    
    layout JSONB NOT NULL DEFAULT '[]',
    
    visibility VARCHAR(50) DEFAULT 'private',
    share_token VARCHAR(100) UNIQUE,
    is_enforced BOOLEAN DEFAULT FALSE,
    enforced_for_roles TEXT[],
    
    created_via VARCHAR(50) DEFAULT 'agent',
    
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_dashboards_user ON user_dashboards(user_id);
CREATE INDEX idx_dashboards_workspace ON user_dashboards(workspace_id);
CREATE INDEX idx_dashboards_share_token ON user_dashboards(share_token) WHERE share_token IS NOT NULL;
CREATE INDEX idx_dashboards_default ON user_dashboards(user_id, is_default) WHERE is_default = TRUE;

-- Dashboard Widgets table (Widget Type Registry)
CREATE TABLE dashboard_widgets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    widget_type VARCHAR(100) NOT NULL UNIQUE,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    category VARCHAR(50),
    
    config_schema JSONB NOT NULL,
    default_config JSONB DEFAULT '{}',
    default_size JSONB DEFAULT '{"w": 4, "h": 3}',
    min_size JSONB DEFAULT '{"w": 2, "h": 2}',
    
    sse_events TEXT[],
    
    is_enabled BOOLEAN DEFAULT TRUE,
    requires_feature VARCHAR(100),
    
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Dashboard Templates table
CREATE TABLE dashboard_templates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    description TEXT,
    category VARCHAR(50),
    
    layout JSONB NOT NULL,
    
    thumbnail_url TEXT,
    is_default BOOLEAN DEFAULT FALSE,
    sort_order INTEGER DEFAULT 0,
    
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Push devices for mobile notifications
CREATE TABLE push_devices (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id TEXT NOT NULL REFERENCES "user"(id) ON DELETE CASCADE,
    device_id TEXT NOT NULL,
    platform TEXT NOT NULL,
    push_token TEXT NOT NULL,
    app_version TEXT,
    os_version TEXT,
    device_model TEXT,
    is_active BOOLEAN DEFAULT true,
    last_used_at TIMESTAMPTZ DEFAULT NOW(),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(user_id, device_id)
);

CREATE INDEX idx_push_devices_user ON push_devices(user_id);
CREATE INDEX idx_push_devices_token ON push_devices(push_token);
-- ============================================================================
-- WORKSPACE & TEAM COLLABORATION SYSTEM
-- ============================================================================

-- Workspaces table (multi-tenant containers)
CREATE TABLE workspaces (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(100) NOT NULL UNIQUE,
    description TEXT,
    logo_url VARCHAR(500),
    plan_type VARCHAR(50) DEFAULT 'free',
    max_members INTEGER DEFAULT 5,
    max_projects INTEGER DEFAULT 10,
    max_storage_gb INTEGER DEFAULT 5,
    settings JSONB DEFAULT '{}',
    owner_id VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_workspaces_owner ON workspaces(owner_id);
CREATE INDEX idx_workspaces_slug ON workspaces(slug);

-- Workspace Roles table
CREATE TABLE workspace_roles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    display_name VARCHAR(100) NOT NULL,
    description TEXT,
    color VARCHAR(20),
    icon VARCHAR(50),
    permissions JSONB NOT NULL DEFAULT '{}',
    is_system BOOLEAN DEFAULT FALSE,
    is_default BOOLEAN DEFAULT FALSE,
    hierarchy_level INTEGER DEFAULT 0,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(workspace_id, name)
);

CREATE INDEX idx_workspace_roles_workspace ON workspace_roles(workspace_id);

-- Workspace Members table
CREATE TABLE workspace_members (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    user_id VARCHAR(255) NOT NULL,
    role_id UUID REFERENCES workspace_roles(id) ON DELETE SET NULL,
    role_name VARCHAR(100),
    status VARCHAR(50) DEFAULT 'active',
    invited_by VARCHAR(255),
    invited_at TIMESTAMPTZ,
    joined_at TIMESTAMPTZ,
    custom_permissions JSONB,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(workspace_id, user_id)
);

CREATE INDEX idx_workspace_members_workspace ON workspace_members(workspace_id);
CREATE INDEX idx_workspace_members_user ON workspace_members(user_id);
CREATE INDEX idx_workspace_members_status ON workspace_members(status);

-- Workspace Invitations table
CREATE TABLE workspace_invitations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    email VARCHAR(255) NOT NULL,
    token VARCHAR(64) NOT NULL UNIQUE,
    role_id UUID REFERENCES workspace_roles(id) ON DELETE SET NULL,
    role_name VARCHAR(100) NOT NULL,
    invited_by_id VARCHAR(255) NOT NULL,
    invited_by_name VARCHAR(255),
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    expires_at TIMESTAMPTZ NOT NULL,
    accepted_at TIMESTAMPTZ,
    accepted_by_user_id VARCHAR(255),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_workspace_invitations_token ON workspace_invitations(token);
CREATE INDEX idx_workspace_invitations_workspace ON workspace_invitations(workspace_id);
CREATE INDEX idx_workspace_invitations_email ON workspace_invitations(email);

-- Workspace Memories table
CREATE TABLE workspace_memories (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    title VARCHAR(255) NOT NULL,
    summary TEXT NOT NULL,
    content TEXT NOT NULL,
    memory_type VARCHAR(50) NOT NULL,
    category VARCHAR(100),
    scope_type VARCHAR(50) DEFAULT 'workspace',
    scope_id UUID,
    visibility VARCHAR(50) DEFAULT 'team',
    created_by VARCHAR(255) NOT NULL,
    importance_score DECIMAL(3,2) DEFAULT 0.5,
    access_count INTEGER DEFAULT 0,
    tags TEXT[] DEFAULT '{}',
    metadata JSONB DEFAULT '{}',
    is_active BOOLEAN DEFAULT TRUE,
    is_pinned BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_workspace_memories_workspace ON workspace_memories(workspace_id);
CREATE INDEX idx_workspace_memories_scope ON workspace_memories(scope_type, scope_id);
CREATE INDEX idx_workspace_memories_type ON workspace_memories(memory_type);

-- User Workspace Profiles table
CREATE TABLE user_workspace_profiles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    user_id VARCHAR(255) NOT NULL,
    display_name VARCHAR(255),
    title VARCHAR(100),
    department VARCHAR(100),
    avatar_url VARCHAR(500),
    work_email VARCHAR(255),
    phone VARCHAR(50),
    timezone VARCHAR(50),
    working_hours JSONB,
    notification_preferences JSONB,
    preferred_output_style VARCHAR(50),
    communication_preferences JSONB,
    expertise_areas TEXT[] DEFAULT '{}',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(workspace_id, user_id)
);

CREATE INDEX idx_user_workspace_profiles_workspace ON user_workspace_profiles(workspace_id);
CREATE INDEX idx_user_workspace_profiles_user ON user_workspace_profiles(user_id);

-- Workspace Project Members table
CREATE TABLE workspace_project_members (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    user_id VARCHAR(255) NOT NULL,
    workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    project_role VARCHAR(50) DEFAULT 'contributor',
    assigned_by VARCHAR(255),
    assigned_at TIMESTAMPTZ DEFAULT NOW(),
    notification_level VARCHAR(50) DEFAULT 'all',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(project_id, user_id)
);

CREATE INDEX idx_workspace_project_members_project ON workspace_project_members(project_id);
CREATE INDEX idx_workspace_project_members_user ON workspace_project_members(user_id);
CREATE INDEX idx_workspace_project_members_workspace ON workspace_project_members(workspace_id);

-- ============================================================================
-- ANALYTICS & DASHBOARD ENHANCEMENTS
-- ============================================================================

-- Analytics Snapshots - Daily metrics snapshots for trend analysis
CREATE TABLE analytics_snapshots (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(255) NOT NULL,
    workspace_id UUID REFERENCES workspaces(id) ON DELETE CASCADE,
    snapshot_date DATE NOT NULL,
    metrics JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(user_id, snapshot_date)
);

CREATE INDEX idx_analytics_snapshots_user_date ON analytics_snapshots(user_id, snapshot_date DESC);
CREATE INDEX idx_analytics_snapshots_workspace ON analytics_snapshots(workspace_id) WHERE workspace_id IS NOT NULL;
CREATE INDEX idx_analytics_snapshots_date ON analytics_snapshots(snapshot_date DESC);

-- Dashboard Views - Tracks dashboard usage for analytics
CREATE TABLE dashboard_views (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    dashboard_id UUID NOT NULL REFERENCES user_dashboards(id) ON DELETE CASCADE,
    user_id VARCHAR(255) NOT NULL,
    viewed_at TIMESTAMPTZ DEFAULT NOW(),
    session_id VARCHAR(100),
    duration_seconds INTEGER,
    widget_interactions JSONB DEFAULT '[]',
    source VARCHAR(50),
    device_type VARCHAR(20)
);

CREATE INDEX idx_dashboard_views_dashboard ON dashboard_views(dashboard_id, viewed_at DESC);
CREATE INDEX idx_dashboard_views_user ON dashboard_views(user_id, viewed_at DESC);
CREATE INDEX idx_dashboard_views_date ON dashboard_views(viewed_at DESC);

-- Dashboard Shares - Granular sharing permissions
CREATE TABLE dashboard_shares (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    dashboard_id UUID NOT NULL REFERENCES user_dashboards(id) ON DELETE CASCADE,
    shared_with_user_id VARCHAR(255),
    shared_with_role VARCHAR(100),
    shared_with_workspace_id UUID REFERENCES workspaces(id) ON DELETE CASCADE,
    permission VARCHAR(20) DEFAULT 'view' CHECK (permission IN ('view', 'edit', 'admin')),
    expires_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    created_by VARCHAR(255) NOT NULL,
    UNIQUE(dashboard_id, shared_with_user_id),
    UNIQUE(dashboard_id, shared_with_role),
    CONSTRAINT chk_share_target CHECK (
        shared_with_user_id IS NOT NULL OR 
        shared_with_role IS NOT NULL OR 
        shared_with_workspace_id IS NOT NULL
    )
);

CREATE INDEX idx_dashboard_shares_dashboard ON dashboard_shares(dashboard_id);
CREATE INDEX idx_dashboard_shares_user ON dashboard_shares(shared_with_user_id) WHERE shared_with_user_id IS NOT NULL;
CREATE INDEX idx_dashboard_shares_role ON dashboard_shares(shared_with_role) WHERE shared_with_role IS NOT NULL;
CREATE INDEX idx_dashboard_shares_workspace ON dashboard_shares(shared_with_workspace_id) WHERE shared_with_workspace_id IS NOT NULL;
CREATE INDEX idx_dashboard_shares_expiry ON dashboard_shares(expires_at) WHERE expires_at IS NOT NULL;

-- Widget Data Cache - Caches expensive widget queries
CREATE TABLE widget_data_cache (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(255) NOT NULL,
    widget_type VARCHAR(100) NOT NULL,
    cache_key VARCHAR(255) NOT NULL,
    data JSONB NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    hit_count INTEGER DEFAULT 0,
    last_hit_at TIMESTAMPTZ,
    UNIQUE(user_id, widget_type, cache_key)
);

CREATE INDEX idx_widget_cache_lookup ON widget_data_cache(user_id, widget_type, cache_key);
CREATE INDEX idx_widget_cache_expiry ON widget_data_cache(expires_at);
CREATE INDEX idx_widget_cache_type ON widget_data_cache(widget_type);

-- ===== COMMENTS SYSTEM =====

-- Generic comments attachable to any entity
CREATE TABLE comments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(255) NOT NULL,
    entity_type VARCHAR(50) NOT NULL,
    entity_id UUID NOT NULL,
    content TEXT NOT NULL,
    parent_id UUID REFERENCES comments(id) ON DELETE CASCADE,
    is_edited BOOLEAN DEFAULT FALSE,
    edited_at TIMESTAMPTZ,
    is_deleted BOOLEAN DEFAULT FALSE,
    deleted_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_comments_entity ON comments(entity_type, entity_id);
CREATE INDEX idx_comments_user ON comments(user_id);
CREATE INDEX idx_comments_parent ON comments(parent_id) WHERE parent_id IS NOT NULL;
CREATE INDEX idx_comments_created ON comments(entity_type, entity_id, created_at DESC);

-- @user mentions tracking
CREATE TABLE entity_mentions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    source_type VARCHAR(50) NOT NULL,
    source_id UUID NOT NULL,
    mentioned_user_id VARCHAR(255) NOT NULL,
    mention_text VARCHAR(100) NOT NULL,
    position_in_text INT,
    entity_type VARCHAR(50),
    entity_id UUID,
    mentioned_by VARCHAR(255) NOT NULL,
    notified BOOLEAN DEFAULT FALSE,
    notified_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_entity_mentions_source ON entity_mentions(source_type, source_id);
CREATE INDEX idx_entity_mentions_user ON entity_mentions(mentioned_user_id);
CREATE INDEX idx_entity_mentions_entity ON entity_mentions(entity_type, entity_id);

-- Comment reactions (emoji)
CREATE TABLE comment_reactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    comment_id UUID NOT NULL REFERENCES comments(id) ON DELETE CASCADE,
    user_id VARCHAR(255) NOT NULL,
    emoji VARCHAR(20) NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(comment_id, user_id, emoji)
);

CREATE INDEX idx_comment_reactions_comment ON comment_reactions(comment_id);