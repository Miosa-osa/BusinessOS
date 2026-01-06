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

-- ===== INTEGRATIONS MODULE (Migration 025) =====

-- Integration Providers (system-defined catalog)
CREATE TABLE integration_providers (
    id VARCHAR(50) PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    category VARCHAR(50) NOT NULL,
    icon_url TEXT,
    oauth_config JSONB NOT NULL DEFAULT '{}',
    modules TEXT[] NOT NULL DEFAULT '{}',
    skills TEXT[] NOT NULL DEFAULT '{}',
    status VARCHAR(20) DEFAULT 'available',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- User Integration Connections
CREATE TABLE user_integrations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(255) NOT NULL REFERENCES "user"(id) ON DELETE CASCADE,
    provider_id VARCHAR(50) NOT NULL REFERENCES integration_providers(id),
    status VARCHAR(20) DEFAULT 'connected',
    connected_at TIMESTAMPTZ DEFAULT NOW(),
    last_used_at TIMESTAMPTZ,
    access_token_encrypted BYTEA,
    refresh_token_encrypted BYTEA,
    token_expires_at TIMESTAMPTZ,
    scopes TEXT[],
    external_account_id VARCHAR(255),
    external_account_name VARCHAR(255),
    external_workspace_id VARCHAR(255),
    external_workspace_name VARCHAR(255),
    metadata JSONB DEFAULT '{}',
    settings JSONB DEFAULT '{"enabledSkills": [], "notifications": true}',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(user_id, provider_id)
);

CREATE INDEX idx_user_integrations_user ON user_integrations(user_id);
CREATE INDEX idx_user_integrations_provider ON user_integrations(provider_id);
CREATE INDEX idx_user_integrations_status ON user_integrations(status);

-- Module Integration Settings
CREATE TABLE module_integration_settings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(255) NOT NULL REFERENCES "user"(id) ON DELETE CASCADE,
    module_id VARCHAR(50) NOT NULL,
    provider_id VARCHAR(50) NOT NULL REFERENCES integration_providers(id),
    enabled BOOLEAN DEFAULT true,
    sync_direction VARCHAR(20) DEFAULT 'bidirectional',
    sync_frequency VARCHAR(20) DEFAULT 'realtime',
    custom_settings JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(user_id, module_id, provider_id)
);

CREATE INDEX idx_module_integration_settings_user ON module_integration_settings(user_id);
CREATE INDEX idx_module_integration_settings_module ON module_integration_settings(module_id);

-- User Model Preferences
CREATE TABLE user_model_preferences (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(255) NOT NULL REFERENCES "user"(id) ON DELETE CASCADE UNIQUE,
    tier_2_model JSONB DEFAULT '{"model_id": "claude-3-5-haiku", "provider": "anthropic"}',
    tier_3_model JSONB DEFAULT '{"model_id": "claude-sonnet-4", "provider": "anthropic"}',
    tier_4_model JSONB DEFAULT '{"model_id": "claude-opus-4", "provider": "anthropic"}',
    tier_2_fallbacks JSONB DEFAULT '[]',
    tier_3_fallbacks JSONB DEFAULT '[]',
    tier_4_fallbacks JSONB DEFAULT '[]',
    skill_overrides JSONB DEFAULT '{}',
    allow_model_upgrade_on_failure BOOLEAN DEFAULT true,
    max_latency_ms INTEGER DEFAULT 30000,
    prefer_local BOOLEAN DEFAULT false,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Pending Decisions (human-in-the-loop)
CREATE TABLE pending_decisions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    execution_id VARCHAR(255) NOT NULL,
    skill_id VARCHAR(255) NOT NULL,
    step_id VARCHAR(255) NOT NULL,
    user_id VARCHAR(255) NOT NULL REFERENCES "user"(id) ON DELETE CASCADE,
    question TEXT NOT NULL,
    description TEXT,
    options TEXT[],
    input_fields JSONB,
    context JSONB,
    priority VARCHAR(20) DEFAULT 'medium',
    status VARCHAR(20) DEFAULT 'pending',
    decision TEXT,
    decision_inputs JSONB,
    decided_by VARCHAR(255) REFERENCES "user"(id),
    decided_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    expires_at TIMESTAMPTZ
);

CREATE INDEX idx_pending_decisions_user_status ON pending_decisions(user_id, status);
CREATE INDEX idx_pending_decisions_execution ON pending_decisions(execution_id);
CREATE INDEX idx_pending_decisions_expires ON pending_decisions(expires_at) WHERE status = 'pending';

-- Integration Sync Log
CREATE TABLE integration_sync_log (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_integration_id UUID NOT NULL REFERENCES user_integrations(id) ON DELETE CASCADE,
    module_id VARCHAR(50),
    sync_type VARCHAR(50) NOT NULL,
    direction VARCHAR(20) NOT NULL,
    status VARCHAR(20) NOT NULL,
    records_processed INT DEFAULT 0,
    records_created INT DEFAULT 0,
    records_updated INT DEFAULT 0,
    records_failed INT DEFAULT 0,
    error_message TEXT,
    error_details JSONB,
    started_at TIMESTAMPTZ DEFAULT NOW(),
    completed_at TIMESTAMPTZ
);

CREATE INDEX idx_integration_sync_log_integration ON integration_sync_log(user_integration_id);
CREATE INDEX idx_integration_sync_log_started ON integration_sync_log(started_at DESC);

-- Skill Executions
CREATE TABLE skill_executions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    skill_id VARCHAR(255) NOT NULL,
    user_id VARCHAR(255) NOT NULL REFERENCES "user"(id) ON DELETE CASCADE,
    status VARCHAR(20) NOT NULL,
    current_step INT DEFAULT 0,
    params JSONB DEFAULT '{}',
    result JSONB,
    error TEXT,
    context JSONB DEFAULT '{}',
    step_results JSONB DEFAULT '{}',
    metrics JSONB DEFAULT '{}',
    started_at TIMESTAMPTZ DEFAULT NOW(),
    completed_at TIMESTAMPTZ
);

CREATE INDEX idx_skill_executions_user ON skill_executions(user_id);
CREATE INDEX idx_skill_executions_status ON skill_executions(status);

-- ===== CREDENTIAL VAULT (Migration 027) =====

-- Unified credential storage with encryption
CREATE TABLE credential_vault (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(255) NOT NULL REFERENCES "user"(id) ON DELETE CASCADE,
    provider_id VARCHAR(50) NOT NULL,
    credential_type VARCHAR(20) NOT NULL DEFAULT 'oauth',
    encrypted_data BYTEA NOT NULL,
    encryption_version INT DEFAULT 1,
    expires_at TIMESTAMPTZ,
    external_account_id VARCHAR(255),
    external_account_email VARCHAR(255),
    external_workspace_id VARCHAR(255),
    external_workspace_name VARCHAR(255),
    scopes TEXT[] DEFAULT '{}',
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    last_used_at TIMESTAMPTZ,
    last_rotated_at TIMESTAMPTZ,
    UNIQUE(user_id, provider_id)
);

CREATE INDEX idx_credential_vault_user ON credential_vault(user_id);
CREATE INDEX idx_credential_vault_provider ON credential_vault(provider_id);
CREATE INDEX idx_credential_vault_type ON credential_vault(credential_type);
CREATE INDEX idx_credential_vault_expires ON credential_vault(expires_at) WHERE expires_at IS NOT NULL;

-- Webhook registrations
CREATE TABLE integration_webhooks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(255) NOT NULL REFERENCES "user"(id) ON DELETE CASCADE,
    provider_id VARCHAR(50) NOT NULL,
    webhook_url TEXT NOT NULL,
    webhook_secret_encrypted BYTEA,
    events TEXT[] NOT NULL DEFAULT '{}',
    status VARCHAR(20) DEFAULT 'active',
    last_triggered_at TIMESTAMPTZ,
    failure_count INT DEFAULT 0,
    last_error TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(user_id, provider_id, webhook_url)
);

CREATE INDEX idx_webhooks_user ON integration_webhooks(user_id);
CREATE INDEX idx_webhooks_provider ON integration_webhooks(provider_id);
CREATE INDEX idx_webhooks_status ON integration_webhooks(status);

-- Data sync mappings
CREATE TABLE data_sync_mappings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(255) NOT NULL REFERENCES "user"(id) ON DELETE CASCADE,
    source_provider VARCHAR(50) NOT NULL,
    source_entity VARCHAR(100) NOT NULL,
    target_module VARCHAR(50) NOT NULL,
    target_entity VARCHAR(100),
    field_mappings JSONB NOT NULL DEFAULT '{}',
    transform_rules JSONB DEFAULT '{}',
    enabled BOOLEAN DEFAULT true,
    sync_direction VARCHAR(20) DEFAULT 'import',
    sync_frequency VARCHAR(20) DEFAULT 'manual',
    last_synced_at TIMESTAMPTZ,
    records_synced INT DEFAULT 0,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(user_id, source_provider, source_entity, target_module)
);

CREATE INDEX idx_sync_mappings_user ON data_sync_mappings(user_id);
CREATE INDEX idx_sync_mappings_provider ON data_sync_mappings(source_provider);
CREATE INDEX idx_sync_mappings_enabled ON data_sync_mappings(enabled) WHERE enabled = true;

-- ===== DATA IMPORTS (Migration 028) =====

-- Import status enum
CREATE TYPE import_status AS ENUM (
    'pending', 'validating', 'mapping', 'processing', 'completed', 'failed', 'cancelled'
);

-- Import source type enum
CREATE TYPE import_source_type AS ENUM (
    'chatgpt_export', 'claude_export', 'custom_chat_export',
    'hubspot_contacts', 'hubspot_deals', 'hubspot_companies',
    'salesforce_contacts', 'salesforce_accounts',
    'linear_issues', 'notion_database', 'asana_tasks', 'jira_issues',
    'google_calendar', 'outlook_calendar',
    'fathom_analytics', 'plausible_analytics',
    'google_drive', 'dropbox', 'notion_pages',
    'csv_generic', 'json_generic', 'custom'
);

-- Import jobs table
CREATE TABLE import_jobs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(255) NOT NULL REFERENCES "user"(id) ON DELETE CASCADE,
    source_type import_source_type NOT NULL,
    source_provider VARCHAR(50),
    original_filename VARCHAR(500),
    file_size_bytes BIGINT,
    content_type VARCHAR(100),
    status import_status DEFAULT 'pending',
    progress_percent INT DEFAULT 0,
    total_records INT DEFAULT 0,
    processed_records INT DEFAULT 0,
    imported_records INT DEFAULT 0,
    skipped_records INT DEFAULT 0,
    failed_records INT DEFAULT 0,
    field_mapping JSONB DEFAULT '{}',
    transform_rules JSONB DEFAULT '{}',
    import_options JSONB DEFAULT '{}',
    target_module VARCHAR(50) NOT NULL,
    target_entity VARCHAR(100),
    result_summary JSONB DEFAULT '{}',
    error_log JSONB DEFAULT '[]',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    started_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    error_message TEXT,
    error_details JSONB
);

CREATE INDEX idx_import_jobs_user ON import_jobs(user_id);
CREATE INDEX idx_import_jobs_status ON import_jobs(status);
CREATE INDEX idx_import_jobs_source ON import_jobs(source_type);
CREATE INDEX idx_import_jobs_created ON import_jobs(created_at DESC);

-- Imported records tracking (deduplication)
CREATE TABLE imported_records (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(255) NOT NULL REFERENCES "user"(id) ON DELETE CASCADE,
    import_job_id UUID REFERENCES import_jobs(id) ON DELETE SET NULL,
    source_type import_source_type NOT NULL,
    source_provider VARCHAR(50),
    external_id VARCHAR(500) NOT NULL,
    target_module VARCHAR(50) NOT NULL,
    target_entity VARCHAR(100),
    target_record_id UUID NOT NULL,
    external_data_hash VARCHAR(64),
    last_synced_at TIMESTAMPTZ DEFAULT NOW(),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(user_id, source_type, external_id)
);

CREATE INDEX idx_imported_records_user ON imported_records(user_id);
CREATE INDEX idx_imported_records_source ON imported_records(source_type, external_id);
CREATE INDEX idx_imported_records_target ON imported_records(target_module, target_record_id);

-- Import mapping templates
CREATE TABLE import_mapping_templates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    source_type import_source_type NOT NULL,
    target_module VARCHAR(50) NOT NULL,
    template_name VARCHAR(100) NOT NULL,
    field_mappings JSONB NOT NULL DEFAULT '{}',
    transform_rules JSONB DEFAULT '{}',
    default_values JSONB DEFAULT '{}',
    description TEXT,
    is_system_template BOOLEAN DEFAULT FALSE,
    created_by VARCHAR(255),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(source_type, target_module, template_name)
);

-- Imported conversations (ChatGPT, Claude, etc.)
CREATE TABLE imported_conversations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(255) NOT NULL REFERENCES "user"(id) ON DELETE CASCADE,
    import_job_id UUID REFERENCES import_jobs(id) ON DELETE SET NULL,
    source_type import_source_type NOT NULL,
    external_conversation_id VARCHAR(255),
    title VARCHAR(500),
    model VARCHAR(100),
    messages JSONB NOT NULL DEFAULT '[]',
    message_count INT DEFAULT 0,
    original_created_at TIMESTAMPTZ,
    original_updated_at TIMESTAMPTZ,
    metadata JSONB DEFAULT '{}',
    linked_context_id UUID,
    linked_project_id UUID,
    tags TEXT[] DEFAULT '{}',
    search_content TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_imported_conversations_user ON imported_conversations(user_id);
CREATE INDEX idx_imported_conversations_source ON imported_conversations(source_type);
CREATE INDEX idx_imported_conversations_job ON imported_conversations(import_job_id);
CREATE INDEX idx_imported_conversations_search ON imported_conversations USING GIN(to_tsvector('english', search_content));

-- ============================================================================
-- FATHOM ANALYTICS TABLES
-- ============================================================================

-- Fathom sites (website properties)
CREATE TABLE fathom_sites (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(255) NOT NULL,
    site_id VARCHAR(100) NOT NULL,
    name VARCHAR(255),
    sharing_url TEXT,
    share_config VARCHAR(50),
    synced_at TIMESTAMPTZ DEFAULT NOW(),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(user_id, site_id)
);

CREATE INDEX idx_fathom_sites_user ON fathom_sites(user_id);

-- Fathom aggregations (daily analytics data)
CREATE TABLE fathom_aggregations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(255) NOT NULL,
    site_id VARCHAR(100) NOT NULL,
    date DATE NOT NULL,
    visits INT DEFAULT 0,
    uniques INT DEFAULT 0,
    pageviews INT DEFAULT 0,
    avg_duration DECIMAL(10,2) DEFAULT 0,
    bounce_rate DECIMAL(5,2) DEFAULT 0,
    synced_at TIMESTAMPTZ DEFAULT NOW(),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(user_id, site_id, date)
);

CREATE INDEX idx_fathom_agg_user_site ON fathom_aggregations(user_id, site_id);
CREATE INDEX idx_fathom_agg_date ON fathom_aggregations(user_id, site_id, date DESC);

-- Fathom page-level analytics (grouped by pathname)
CREATE TABLE fathom_pages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(255) NOT NULL,
    site_id VARCHAR(100) NOT NULL,
    pathname VARCHAR(500) NOT NULL,
    hostname VARCHAR(255),
    visits INT DEFAULT 0,
    uniques INT DEFAULT 0,
    pageviews INT DEFAULT 0,
    avg_duration DECIMAL(10,2) DEFAULT 0,
    bounce_rate DECIMAL(5,2) DEFAULT 0,
    period_start DATE NOT NULL,
    period_end DATE NOT NULL,
    synced_at TIMESTAMPTZ DEFAULT NOW(),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(user_id, site_id, pathname, period_start, period_end)
);

CREATE INDEX idx_fathom_pages_user_site ON fathom_pages(user_id, site_id);
CREATE INDEX idx_fathom_pages_pathname ON fathom_pages(user_id, pathname);

-- Fathom referrers analytics
CREATE TABLE fathom_referrers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(255) NOT NULL,
    site_id VARCHAR(100) NOT NULL,
    referrer VARCHAR(500) NOT NULL,
    visits INT DEFAULT 0,
    uniques INT DEFAULT 0,
    period_start DATE NOT NULL,
    period_end DATE NOT NULL,
    synced_at TIMESTAMPTZ DEFAULT NOW(),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(user_id, site_id, referrer, period_start, period_end)
);

CREATE INDEX idx_fathom_referrers_user_site ON fathom_referrers(user_id, site_id);

-- Fathom custom events
CREATE TABLE fathom_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(255) NOT NULL,
    site_id VARCHAR(100) NOT NULL,
    event_id VARCHAR(100) NOT NULL,
    event_name VARCHAR(255) NOT NULL,
    count INT DEFAULT 0,
    period_start DATE NOT NULL,
    period_end DATE NOT NULL,
    synced_at TIMESTAMPTZ DEFAULT NOW(),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(user_id, site_id, event_id, period_start, period_end)
);

CREATE INDEX idx_fathom_events_user_site ON fathom_events(user_id, site_id);

-- ============================================================================
-- GOOGLE DRIVE/DOCS TABLES
-- ============================================================================

-- Google Drive files
CREATE TABLE google_drive_files (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(255) NOT NULL,
    file_id VARCHAR(255) NOT NULL,
    name VARCHAR(500) NOT NULL,
    mime_type VARCHAR(255),
    file_extension VARCHAR(50),
    size_bytes BIGINT,
    parent_folder_id VARCHAR(255),
    parent_folder_name VARCHAR(500),
    path TEXT,
    shared BOOLEAN DEFAULT FALSE,
    sharing_user VARCHAR(255),
    permissions JSONB DEFAULT '[]',
    web_view_link TEXT,
    web_content_link TEXT,
    thumbnail_link TEXT,
    icon_link TEXT,
    created_time TIMESTAMPTZ,
    modified_time TIMESTAMPTZ,
    viewed_by_me_time TIMESTAMPTZ,
    owners JSONB DEFAULT '[]',
    last_modifying_user JSONB,
    synced_at TIMESTAMPTZ DEFAULT NOW(),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(user_id, file_id)
);

CREATE INDEX idx_drive_files_user ON google_drive_files(user_id);
CREATE INDEX idx_drive_files_parent ON google_drive_files(user_id, parent_folder_id);
CREATE INDEX idx_drive_files_mime ON google_drive_files(user_id, mime_type);
CREATE INDEX idx_drive_files_modified ON google_drive_files(user_id, modified_time DESC);

-- Google Docs content
CREATE TABLE google_docs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(255) NOT NULL,
    document_id VARCHAR(255) NOT NULL,
    drive_file_id UUID REFERENCES google_drive_files(id) ON DELETE SET NULL,
    title VARCHAR(500) NOT NULL,
    body_text TEXT,
    word_count INT DEFAULT 0,
    headers JSONB DEFAULT '[]',
    locale VARCHAR(20),
    created_time TIMESTAMPTZ,
    modified_time TIMESTAMPTZ,
    synced_at TIMESTAMPTZ DEFAULT NOW(),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(user_id, document_id)
);

CREATE INDEX idx_docs_user ON google_docs(user_id);
CREATE INDEX idx_docs_title ON google_docs(user_id, title);
CREATE INDEX idx_docs_modified ON google_docs(user_id, modified_time DESC);
CREATE INDEX idx_docs_search ON google_docs USING GIN(to_tsvector('english', body_text));

-- Google Sheets
CREATE TABLE google_sheets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(255) NOT NULL,
    spreadsheet_id VARCHAR(255) NOT NULL,
    drive_file_id UUID REFERENCES google_drive_files(id) ON DELETE SET NULL,
    title VARCHAR(500) NOT NULL,
    locale VARCHAR(20),
    time_zone VARCHAR(100),
    sheet_count INT DEFAULT 0,
    sheets JSONB DEFAULT '[]',
    named_ranges JSONB DEFAULT '[]',
    created_time TIMESTAMPTZ,
    modified_time TIMESTAMPTZ,
    synced_at TIMESTAMPTZ DEFAULT NOW(),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(user_id, spreadsheet_id)
);

CREATE INDEX idx_sheets_user ON google_sheets(user_id);
CREATE INDEX idx_sheets_title ON google_sheets(user_id, title);

-- Google Slides presentations
CREATE TABLE google_slides (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(255) NOT NULL,
    presentation_id VARCHAR(255) NOT NULL,
    drive_file_id UUID REFERENCES google_drive_files(id) ON DELETE SET NULL,
    title VARCHAR(500) NOT NULL,
    locale VARCHAR(20),
    slide_count INT DEFAULT 0,
    slides JSONB DEFAULT '[]',
    page_width DECIMAL(10,2),
    page_height DECIMAL(10,2),
    created_time TIMESTAMPTZ,
    modified_time TIMESTAMPTZ,
    synced_at TIMESTAMPTZ DEFAULT NOW(),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(user_id, presentation_id)
);

CREATE INDEX idx_slides_user ON google_slides(user_id);
CREATE INDEX idx_slides_title ON google_slides(user_id, title);

-- ============================================================================
-- GOOGLE CONTACTS TABLES
-- ============================================================================

CREATE TABLE google_contacts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(255) NOT NULL,
    resource_name VARCHAR(255) NOT NULL,
    display_name VARCHAR(255),
    given_name VARCHAR(255),
    family_name VARCHAR(255),
    middle_name VARCHAR(255),
    emails JSONB DEFAULT '[]',
    phone_numbers JSONB DEFAULT '[]',
    addresses JSONB DEFAULT '[]',
    organization VARCHAR(255),
    job_title VARCHAR(255),
    department VARCHAR(255),
    photo_url TEXT,
    contact_groups JSONB DEFAULT '[]',
    metadata JSONB DEFAULT '{}',
    created_time TIMESTAMPTZ,
    modified_time TIMESTAMPTZ,
    synced_at TIMESTAMPTZ DEFAULT NOW(),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(user_id, resource_name)
);

CREATE INDEX idx_contacts_user ON google_contacts(user_id);
CREATE INDEX idx_contacts_name ON google_contacts(user_id, display_name);
CREATE INDEX idx_contacts_org ON google_contacts(user_id, organization);

-- ============================================================================
-- GOOGLE TASKS TABLES
-- ============================================================================

CREATE TABLE google_task_lists (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(255) NOT NULL,
    task_list_id VARCHAR(255) NOT NULL,
    title VARCHAR(255) NOT NULL,
    kind VARCHAR(100),
    updated TIMESTAMPTZ,
    synced_at TIMESTAMPTZ DEFAULT NOW(),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(user_id, task_list_id)
);

CREATE INDEX idx_task_lists_user ON google_task_lists(user_id);

CREATE TABLE google_tasks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(255) NOT NULL,
    task_id VARCHAR(255) NOT NULL,
    task_list_id VARCHAR(255) NOT NULL,
    title VARCHAR(500) NOT NULL,
    notes TEXT,
    status VARCHAR(50) DEFAULT 'needsAction',
    due TIMESTAMPTZ,
    completed TIMESTAMPTZ,
    deleted BOOLEAN DEFAULT FALSE,
    hidden BOOLEAN DEFAULT FALSE,
    parent_task_id VARCHAR(255),
    position VARCHAR(100),
    links JSONB DEFAULT '[]',
    updated TIMESTAMPTZ,
    synced_at TIMESTAMPTZ DEFAULT NOW(),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(user_id, task_id)
);

CREATE INDEX idx_google_tasks_user ON google_tasks(user_id);
CREATE INDEX idx_google_tasks_list ON google_tasks(user_id, task_list_id);
CREATE INDEX idx_google_tasks_status ON google_tasks(user_id, status);
CREATE INDEX idx_google_tasks_due ON google_tasks(user_id, due);

-- ============================================================================
-- HUBSPOT CRM TABLES
-- ============================================================================

CREATE TABLE hubspot_contacts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(255) NOT NULL,
    hubspot_id VARCHAR(100) NOT NULL,
    email VARCHAR(255),
    first_name VARCHAR(255),
    last_name VARCHAR(255),
    phone VARCHAR(100),
    company VARCHAR(255),
    job_title VARCHAR(255),
    lifecycle_stage VARCHAR(100),
    lead_status VARCHAR(100),
    owner_id VARCHAR(100),
    properties JSONB DEFAULT '{}',
    created_at_hubspot TIMESTAMPTZ,
    updated_at_hubspot TIMESTAMPTZ,
    synced_at TIMESTAMPTZ DEFAULT NOW(),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(user_id, hubspot_id)
);

CREATE INDEX idx_hubspot_contacts_user ON hubspot_contacts(user_id);
CREATE INDEX idx_hubspot_contacts_email ON hubspot_contacts(user_id, email);

CREATE TABLE hubspot_companies (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(255) NOT NULL,
    hubspot_id VARCHAR(100) NOT NULL,
    name VARCHAR(500),
    domain VARCHAR(255),
    industry VARCHAR(255),
    number_of_employees INT,
    annual_revenue DECIMAL(15,2),
    city VARCHAR(255),
    state VARCHAR(255),
    country VARCHAR(255),
    owner_id VARCHAR(100),
    properties JSONB DEFAULT '{}',
    created_at_hubspot TIMESTAMPTZ,
    updated_at_hubspot TIMESTAMPTZ,
    synced_at TIMESTAMPTZ DEFAULT NOW(),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(user_id, hubspot_id)
);

CREATE INDEX idx_hubspot_companies_user ON hubspot_companies(user_id);
CREATE INDEX idx_hubspot_companies_name ON hubspot_companies(user_id, name);

CREATE TABLE hubspot_deals (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(255) NOT NULL,
    hubspot_id VARCHAR(100) NOT NULL,
    deal_name VARCHAR(500),
    amount DECIMAL(15,2),
    pipeline VARCHAR(255),
    deal_stage VARCHAR(255),
    close_date DATE,
    owner_id VARCHAR(100),
    associated_company_ids JSONB DEFAULT '[]',
    associated_contact_ids JSONB DEFAULT '[]',
    properties JSONB DEFAULT '{}',
    created_at_hubspot TIMESTAMPTZ,
    updated_at_hubspot TIMESTAMPTZ,
    synced_at TIMESTAMPTZ DEFAULT NOW(),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(user_id, hubspot_id)
);

CREATE INDEX idx_hubspot_deals_user ON hubspot_deals(user_id);
CREATE INDEX idx_hubspot_deals_stage ON hubspot_deals(user_id, deal_stage);
CREATE INDEX idx_hubspot_deals_close ON hubspot_deals(user_id, close_date);

-- ============================================================================
-- CLICKUP TABLES
-- ============================================================================

CREATE TABLE clickup_workspaces (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(255) NOT NULL,
    workspace_id VARCHAR(100) NOT NULL,
    name VARCHAR(255) NOT NULL,
    color VARCHAR(50),
    avatar TEXT,
    member_count INT DEFAULT 0,
    synced_at TIMESTAMPTZ DEFAULT NOW(),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(user_id, workspace_id)
);

CREATE INDEX idx_clickup_workspaces_user ON clickup_workspaces(user_id);

CREATE TABLE clickup_spaces (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(255) NOT NULL,
    space_id VARCHAR(100) NOT NULL,
    workspace_id VARCHAR(100) NOT NULL,
    name VARCHAR(255) NOT NULL,
    color VARCHAR(50),
    private BOOLEAN DEFAULT FALSE,
    archived BOOLEAN DEFAULT FALSE,
    synced_at TIMESTAMPTZ DEFAULT NOW(),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(user_id, space_id)
);

CREATE INDEX idx_clickup_spaces_user ON clickup_spaces(user_id);
CREATE INDEX idx_clickup_spaces_workspace ON clickup_spaces(user_id, workspace_id);

CREATE TABLE clickup_folders (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(255) NOT NULL,
    folder_id VARCHAR(100) NOT NULL,
    space_id VARCHAR(100) NOT NULL,
    name VARCHAR(255) NOT NULL,
    hidden BOOLEAN DEFAULT FALSE,
    archived BOOLEAN DEFAULT FALSE,
    synced_at TIMESTAMPTZ DEFAULT NOW(),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(user_id, folder_id)
);

CREATE INDEX idx_clickup_folders_user ON clickup_folders(user_id);
CREATE INDEX idx_clickup_folders_space ON clickup_folders(user_id, space_id);

CREATE TABLE clickup_lists (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(255) NOT NULL,
    list_id VARCHAR(100) NOT NULL,
    folder_id VARCHAR(100),
    space_id VARCHAR(100) NOT NULL,
    name VARCHAR(255) NOT NULL,
    archived BOOLEAN DEFAULT FALSE,
    task_count INT DEFAULT 0,
    synced_at TIMESTAMPTZ DEFAULT NOW(),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(user_id, list_id)
);

CREATE INDEX idx_clickup_lists_user ON clickup_lists(user_id);
CREATE INDEX idx_clickup_lists_folder ON clickup_lists(user_id, folder_id);
CREATE INDEX idx_clickup_lists_space ON clickup_lists(user_id, space_id);

CREATE TABLE clickup_tasks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(255) NOT NULL,
    task_id VARCHAR(100) NOT NULL,
    custom_id VARCHAR(100),
    list_id VARCHAR(100) NOT NULL,
    folder_id VARCHAR(100),
    space_id VARCHAR(100) NOT NULL,
    name VARCHAR(500) NOT NULL,
    description TEXT,
    status VARCHAR(100),
    status_color VARCHAR(50),
    priority VARCHAR(50),
    priority_color VARCHAR(50),
    due_date TIMESTAMPTZ,
    start_date TIMESTAMPTZ,
    date_created TIMESTAMPTZ,
    date_updated TIMESTAMPTZ,
    date_closed TIMESTAMPTZ,
    time_spent BIGINT DEFAULT 0,
    time_estimate BIGINT,
    parent_task_id VARCHAR(100),
    assignees JSONB DEFAULT '[]',
    creator JSONB,
    tags JSONB DEFAULT '[]',
    url TEXT,
    synced_at TIMESTAMPTZ DEFAULT NOW(),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(user_id, task_id)
);

CREATE INDEX idx_clickup_tasks_user ON clickup_tasks(user_id);
CREATE INDEX idx_clickup_tasks_list ON clickup_tasks(user_id, list_id);
CREATE INDEX idx_clickup_tasks_space ON clickup_tasks(user_id, space_id);
CREATE INDEX idx_clickup_tasks_status ON clickup_tasks(user_id, status);
CREATE INDEX idx_clickup_tasks_due ON clickup_tasks(user_id, due_date);

-- ============================================================================
-- AIRTABLE TABLES
-- ============================================================================

CREATE TABLE airtable_bases (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(255) NOT NULL,
    base_id VARCHAR(100) NOT NULL,
    name VARCHAR(255) NOT NULL,
    permission_level VARCHAR(50),
    synced_at TIMESTAMPTZ DEFAULT NOW(),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(user_id, base_id)
);

CREATE INDEX idx_airtable_bases_user ON airtable_bases(user_id);

CREATE TABLE airtable_tables (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(255) NOT NULL,
    table_id VARCHAR(100) NOT NULL,
    base_id VARCHAR(100) NOT NULL,
    name VARCHAR(255) NOT NULL,
    primary_field_id VARCHAR(100),
    fields JSONB DEFAULT '[]',
    views JSONB DEFAULT '[]',
    synced_at TIMESTAMPTZ DEFAULT NOW(),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(user_id, table_id)
);

CREATE INDEX idx_airtable_tables_user ON airtable_tables(user_id);
CREATE INDEX idx_airtable_tables_base ON airtable_tables(user_id, base_id);

CREATE TABLE airtable_records (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(255) NOT NULL,
    record_id VARCHAR(100) NOT NULL,
    table_id VARCHAR(100) NOT NULL,
    base_id VARCHAR(100) NOT NULL,
    fields JSONB DEFAULT '{}',
    created_time_airtable TIMESTAMPTZ,
    synced_at TIMESTAMPTZ DEFAULT NOW(),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(user_id, record_id)
);

CREATE INDEX idx_airtable_records_user ON airtable_records(user_id);
CREATE INDEX idx_airtable_records_table ON airtable_records(user_id, table_id);
CREATE INDEX idx_airtable_records_base ON airtable_records(user_id, base_id);
-- Migration 035: Microsoft 365 Integration Tables
-- This adds storage for Microsoft 365 data (Outlook, OneDrive, To Do, etc.)

-- ============================================================================
-- MICROSOFT OAUTH TOKENS
-- ============================================================================

CREATE TABLE IF NOT EXISTS microsoft_oauth_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(255) NOT NULL UNIQUE,
    access_token TEXT NOT NULL,
    refresh_token TEXT,
    expiry TIMESTAMPTZ,
    scopes TEXT[],
    microsoft_id VARCHAR(255),
    microsoft_email VARCHAR(255),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_microsoft_tokens_user ON microsoft_oauth_tokens(user_id);

-- ============================================================================
-- MICROSOFT OUTLOOK MAIL
-- ============================================================================

CREATE TABLE IF NOT EXISTS microsoft_mail_messages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(255) NOT NULL,
    message_id VARCHAR(255) NOT NULL,

    -- Conversation threading
    conversation_id VARCHAR(255),

    -- Message details
    subject TEXT,
    body_preview TEXT,
    body_content TEXT,
    body_content_type VARCHAR(50), -- text, html
    importance VARCHAR(50), -- low, normal, high

    -- Sender
    from_email VARCHAR(255),
    from_name VARCHAR(255),

    -- Recipients
    to_recipients JSONB DEFAULT '[]',
    cc_recipients JSONB DEFAULT '[]',
    bcc_recipients JSONB DEFAULT '[]',
    reply_to JSONB DEFAULT '[]',

    -- Flags
    is_read BOOLEAN DEFAULT FALSE,
    is_draft BOOLEAN DEFAULT FALSE,
    has_attachments BOOLEAN DEFAULT FALSE,

    -- Folder
    folder_id VARCHAR(255),
    folder_name VARCHAR(255),

    -- Categories/Labels
    categories JSONB DEFAULT '[]',
    flag_status VARCHAR(50), -- notFlagged, flagged, complete

    -- Attachments metadata
    attachments JSONB DEFAULT '[]',

    -- Timestamps
    received_datetime TIMESTAMPTZ,
    sent_datetime TIMESTAMPTZ,
    created_datetime TIMESTAMPTZ,
    last_modified_datetime TIMESTAMPTZ,

    -- Sync tracking
    synced_at TIMESTAMPTZ DEFAULT NOW(),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),

    UNIQUE(user_id, message_id)
);

CREATE INDEX IF NOT EXISTS idx_ms_mail_user ON microsoft_mail_messages(user_id);
CREATE INDEX IF NOT EXISTS idx_ms_mail_conversation ON microsoft_mail_messages(user_id, conversation_id);
CREATE INDEX IF NOT EXISTS idx_ms_mail_folder ON microsoft_mail_messages(user_id, folder_id);
CREATE INDEX IF NOT EXISTS idx_ms_mail_received ON microsoft_mail_messages(user_id, received_datetime DESC);
CREATE INDEX IF NOT EXISTS idx_ms_mail_from ON microsoft_mail_messages(user_id, from_email);

-- ============================================================================
-- MICROSOFT OUTLOOK CALENDAR
-- ============================================================================

CREATE TABLE IF NOT EXISTS microsoft_calendar_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(255) NOT NULL,
    event_id VARCHAR(255) NOT NULL,

    -- Calendar
    calendar_id VARCHAR(255),
    calendar_name VARCHAR(255),

    -- Event details
    subject VARCHAR(500),
    body_preview TEXT,
    body_content TEXT,
    body_content_type VARCHAR(50),

    -- Location
    location_display_name VARCHAR(500),
    location_address JSONB,
    location_coordinates JSONB, -- lat, long

    -- Time
    start_datetime TIMESTAMPTZ,
    start_timezone VARCHAR(100),
    end_datetime TIMESTAMPTZ,
    end_timezone VARCHAR(100),
    is_all_day BOOLEAN DEFAULT FALSE,

    -- Recurrence
    recurrence JSONB, -- pattern, range
    series_master_id VARCHAR(255),
    type VARCHAR(50), -- singleInstance, occurrence, exception, seriesMaster

    -- Attendees
    attendees JSONB DEFAULT '[]',
    organizer_email VARCHAR(255),
    organizer_name VARCHAR(255),

    -- Online meeting
    is_online_meeting BOOLEAN DEFAULT FALSE,
    online_meeting_provider VARCHAR(100), -- teamsForBusiness, skypeForBusiness, etc.
    online_meeting_url TEXT,
    online_meeting_join_url TEXT,

    -- Response
    response_status VARCHAR(50), -- none, organizer, accepted, tentative, declined
    response_time TIMESTAMPTZ,

    -- Flags
    importance VARCHAR(50),
    sensitivity VARCHAR(50), -- normal, personal, private, confidential
    show_as VARCHAR(50), -- free, tentative, busy, oof, workingElsewhere
    is_cancelled BOOLEAN DEFAULT FALSE,
    is_reminder_on BOOLEAN DEFAULT TRUE,
    reminder_minutes_before_start INT DEFAULT 15,

    -- Categories
    categories JSONB DEFAULT '[]',

    -- Timestamps
    created_datetime TIMESTAMPTZ,
    last_modified_datetime TIMESTAMPTZ,

    -- Sync tracking
    synced_at TIMESTAMPTZ DEFAULT NOW(),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),

    UNIQUE(user_id, event_id)
);

CREATE INDEX IF NOT EXISTS idx_ms_calendar_user ON microsoft_calendar_events(user_id);
CREATE INDEX IF NOT EXISTS idx_ms_calendar_calendar ON microsoft_calendar_events(user_id, calendar_id);
CREATE INDEX IF NOT EXISTS idx_ms_calendar_start ON microsoft_calendar_events(user_id, start_datetime);
CREATE INDEX IF NOT EXISTS idx_ms_calendar_series ON microsoft_calendar_events(user_id, series_master_id);

-- ============================================================================
-- MICROSOFT CONTACTS
-- ============================================================================

CREATE TABLE IF NOT EXISTS microsoft_contacts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(255) NOT NULL,
    contact_id VARCHAR(255) NOT NULL,

    -- Name
    display_name VARCHAR(255),
    given_name VARCHAR(255),
    surname VARCHAR(255),
    middle_name VARCHAR(255),
    nickname VARCHAR(255),
    title VARCHAR(50), -- Mr., Ms., Dr., etc.

    -- Contact info
    email_addresses JSONB DEFAULT '[]',
    phone_numbers JSONB DEFAULT '[]',
    addresses JSONB DEFAULT '[]',
    im_addresses JSONB DEFAULT '[]',
    websites JSONB DEFAULT '[]',

    -- Organization
    company_name VARCHAR(255),
    department VARCHAR(255),
    job_title VARCHAR(255),
    office_location VARCHAR(255),
    profession VARCHAR(255),
    manager VARCHAR(255),
    assistant_name VARCHAR(255),

    -- Personal
    birthday DATE,
    spouse_name VARCHAR(255),
    personal_notes TEXT,

    -- Photo
    photo_url TEXT,

    -- Category
    categories JSONB DEFAULT '[]',

    -- Parent folder
    parent_folder_id VARCHAR(255),

    -- Timestamps
    created_datetime TIMESTAMPTZ,
    last_modified_datetime TIMESTAMPTZ,

    -- Sync tracking
    synced_at TIMESTAMPTZ DEFAULT NOW(),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),

    UNIQUE(user_id, contact_id)
);

CREATE INDEX IF NOT EXISTS idx_ms_contacts_user ON microsoft_contacts(user_id);
CREATE INDEX IF NOT EXISTS idx_ms_contacts_name ON microsoft_contacts(user_id, display_name);
CREATE INDEX IF NOT EXISTS idx_ms_contacts_company ON microsoft_contacts(user_id, company_name);

-- ============================================================================
-- MICROSOFT ONEDRIVE
-- ============================================================================

CREATE TABLE IF NOT EXISTS microsoft_onedrive_files (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(255) NOT NULL,
    item_id VARCHAR(255) NOT NULL,

    -- File info
    name VARCHAR(500) NOT NULL,
    description TEXT,
    mime_type VARCHAR(255),
    size_bytes BIGINT,

    -- Path info
    parent_reference_id VARCHAR(255),
    parent_reference_path TEXT,
    web_url TEXT,

    -- Type
    is_folder BOOLEAN DEFAULT FALSE,
    folder_child_count INT,

    -- File specific
    file_hash VARCHAR(255), -- quickXorHash or sha1Hash

    -- Sharing
    shared BOOLEAN DEFAULT FALSE,
    shared_scope VARCHAR(50), -- anonymous, organization, users
    shared_link JSONB,

    -- Permissions
    permissions JSONB DEFAULT '[]',

    -- Created/Modified by
    created_by_user_email VARCHAR(255),
    created_by_user_name VARCHAR(255),
    last_modified_by_user_email VARCHAR(255),
    last_modified_by_user_name VARCHAR(255),

    -- Timestamps
    created_datetime TIMESTAMPTZ,
    last_modified_datetime TIMESTAMPTZ,

    -- Download
    download_url TEXT,

    -- Thumbnails
    thumbnails JSONB DEFAULT '[]',

    -- Sync tracking
    synced_at TIMESTAMPTZ DEFAULT NOW(),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),

    UNIQUE(user_id, item_id)
);

CREATE INDEX IF NOT EXISTS idx_ms_onedrive_user ON microsoft_onedrive_files(user_id);
CREATE INDEX IF NOT EXISTS idx_ms_onedrive_parent ON microsoft_onedrive_files(user_id, parent_reference_id);
CREATE INDEX IF NOT EXISTS idx_ms_onedrive_folder ON microsoft_onedrive_files(user_id, is_folder);
CREATE INDEX IF NOT EXISTS idx_ms_onedrive_name ON microsoft_onedrive_files(user_id, name);
CREATE INDEX IF NOT EXISTS idx_ms_onedrive_modified ON microsoft_onedrive_files(user_id, last_modified_datetime DESC);

-- ============================================================================
-- MICROSOFT TO DO
-- ============================================================================

-- Task Lists
CREATE TABLE IF NOT EXISTS microsoft_todo_lists (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(255) NOT NULL,
    list_id VARCHAR(255) NOT NULL,

    -- List info
    display_name VARCHAR(255) NOT NULL,
    is_owner BOOLEAN DEFAULT TRUE,
    is_shared BOOLEAN DEFAULT FALSE,
    wellknown_list_name VARCHAR(50), -- none, defaultList, flaggedEmails, unknownFutureValue

    -- Sync tracking
    synced_at TIMESTAMPTZ DEFAULT NOW(),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),

    UNIQUE(user_id, list_id)
);

CREATE INDEX IF NOT EXISTS idx_ms_todo_lists_user ON microsoft_todo_lists(user_id);

-- Tasks
CREATE TABLE IF NOT EXISTS microsoft_todo_tasks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(255) NOT NULL,
    task_id VARCHAR(255) NOT NULL,
    list_id VARCHAR(255) NOT NULL,

    -- Task details
    title VARCHAR(500) NOT NULL,
    body_content TEXT,
    body_content_type VARCHAR(50),
    importance VARCHAR(50), -- low, normal, high
    status VARCHAR(50), -- notStarted, inProgress, completed, waitingOnOthers, deferred

    -- Dates
    due_datetime TIMESTAMPTZ,
    due_timezone VARCHAR(100),
    start_datetime TIMESTAMPTZ,
    start_timezone VARCHAR(100),
    completed_datetime TIMESTAMPTZ,
    completed_timezone VARCHAR(100),

    -- Recurrence
    recurrence JSONB,
    is_reminder_on BOOLEAN DEFAULT FALSE,
    reminder_datetime TIMESTAMPTZ,

    -- Categories
    categories JSONB DEFAULT '[]',

    -- Linked resources
    linked_resources JSONB DEFAULT '[]',

    -- Checklist items
    checklist_items JSONB DEFAULT '[]',

    -- Timestamps
    created_datetime TIMESTAMPTZ,
    last_modified_datetime TIMESTAMPTZ,

    -- Sync tracking
    synced_at TIMESTAMPTZ DEFAULT NOW(),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),

    UNIQUE(user_id, task_id)
);

CREATE INDEX IF NOT EXISTS idx_ms_todo_tasks_user ON microsoft_todo_tasks(user_id);
CREATE INDEX IF NOT EXISTS idx_ms_todo_tasks_list ON microsoft_todo_tasks(user_id, list_id);
CREATE INDEX IF NOT EXISTS idx_ms_todo_tasks_status ON microsoft_todo_tasks(user_id, status);
CREATE INDEX IF NOT EXISTS idx_ms_todo_tasks_due ON microsoft_todo_tasks(user_id, due_datetime);

-- ============================================================================
-- TRIGGERS
-- ============================================================================

DO $$
DECLARE
    tables TEXT[] := ARRAY[
        'microsoft_oauth_tokens', 'microsoft_mail_messages', 'microsoft_calendar_events',
        'microsoft_contacts', 'microsoft_onedrive_files', 'microsoft_todo_lists', 'microsoft_todo_tasks'
    ];
    t TEXT;
BEGIN
    FOREACH t IN ARRAY tables
    LOOP
        EXECUTE format('DROP TRIGGER IF EXISTS %I_updated_at ON %I', t, t);
        EXECUTE format('CREATE TRIGGER %I_updated_at BEFORE UPDATE ON %I FOR EACH ROW EXECUTE FUNCTION update_integration_updated_at()', t, t);
    END LOOP;
END $$;

-- Add comments
COMMENT ON TABLE microsoft_oauth_tokens IS 'Microsoft 365 OAuth tokens for users';
COMMENT ON TABLE microsoft_mail_messages IS 'Synced Outlook email messages';
COMMENT ON TABLE microsoft_calendar_events IS 'Synced Outlook calendar events';
COMMENT ON TABLE microsoft_contacts IS 'Synced Outlook contacts';
COMMENT ON TABLE microsoft_onedrive_files IS 'Synced OneDrive files and folders';
COMMENT ON TABLE microsoft_todo_lists IS 'Microsoft To Do task lists';
COMMENT ON TABLE microsoft_todo_tasks IS 'Microsoft To Do tasks';

-- ============================================================================
-- EMAILS (Gmail integration)
-- ============================================================================

CREATE TABLE IF NOT EXISTS emails (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(255) NOT NULL,
    provider VARCHAR(50) NOT NULL DEFAULT 'gmail',
    external_id VARCHAR(255) NOT NULL,
    thread_id VARCHAR(255),

    -- Email metadata
    subject TEXT,
    snippet TEXT,
    from_email VARCHAR(255),
    from_name VARCHAR(255),
    to_emails JSONB DEFAULT '[]',
    cc_emails JSONB DEFAULT '[]',
    bcc_emails JSONB DEFAULT '[]',
    reply_to VARCHAR(255),

    -- Content
    body_text TEXT,
    body_html TEXT,
    attachments JSONB DEFAULT '[]',

    -- Status flags
    is_read BOOLEAN DEFAULT FALSE,
    is_starred BOOLEAN DEFAULT FALSE,
    is_important BOOLEAN DEFAULT FALSE,
    is_draft BOOLEAN DEFAULT FALSE,
    is_sent BOOLEAN DEFAULT FALSE,
    is_archived BOOLEAN DEFAULT FALSE,
    is_trash BOOLEAN DEFAULT FALSE,
    labels JSONB DEFAULT '[]',

    -- Dates
    date TIMESTAMP WITH TIME ZONE,
    received_at TIMESTAMP WITH TIME ZONE,
    synced_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),

    UNIQUE(user_id, provider, external_id)
);

CREATE INDEX IF NOT EXISTS idx_emails_user_thread ON emails(user_id, thread_id);
CREATE INDEX IF NOT EXISTS idx_emails_user_date ON emails(user_id, date DESC);
CREATE INDEX IF NOT EXISTS idx_emails_user_unread ON emails(user_id, is_read) WHERE is_read = FALSE;
CREATE INDEX IF NOT EXISTS idx_emails_user_starred ON emails(user_id, is_starred) WHERE is_starred = TRUE;
CREATE INDEX IF NOT EXISTS idx_emails_user_provider ON emails(user_id, provider);

-- ============================================================================
-- CHANNELS (Slack/Discord/Teams)
-- ============================================================================

CREATE TABLE IF NOT EXISTS channels (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(255) NOT NULL,
    provider VARCHAR(50) NOT NULL,
    external_id VARCHAR(255) NOT NULL,
    external_workspace_id VARCHAR(255),
    external_workspace_name VARCHAR(255),

    -- Channel metadata
    name VARCHAR(255) NOT NULL,
    description TEXT,
    topic TEXT,
    is_private BOOLEAN DEFAULT FALSE,
    is_archived BOOLEAN DEFAULT FALSE,
    is_dm BOOLEAN DEFAULT FALSE,
    member_count INT DEFAULT 0,
    unread_count INT DEFAULT 0,

    -- Dates
    last_message_at TIMESTAMP WITH TIME ZONE,
    synced_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),

    UNIQUE(user_id, provider, external_id)
);

CREATE INDEX IF NOT EXISTS idx_channels_user_provider ON channels(user_id, provider);

-- ============================================================================
-- CHANNEL MESSAGES
-- ============================================================================

CREATE TABLE IF NOT EXISTS channel_messages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    channel_id UUID REFERENCES channels(id) ON DELETE CASCADE,
    user_id VARCHAR(255) NOT NULL,
    external_id VARCHAR(255) NOT NULL,

    -- Sender info
    sender_id VARCHAR(255),
    sender_name VARCHAR(255),
    sender_avatar VARCHAR(500),

    -- Message content
    content TEXT,
    content_html TEXT,
    attachments JSONB DEFAULT '[]',
    reactions JSONB DEFAULT '[]',
    mentions JSONB DEFAULT '[]',

    -- Thread info
    thread_ts VARCHAR(50),
    parent_message_id UUID REFERENCES channel_messages(id),
    reply_count INT DEFAULT 0,
    is_thread_root BOOLEAN DEFAULT FALSE,

    -- Status
    is_edited BOOLEAN DEFAULT FALSE,
    is_deleted BOOLEAN DEFAULT FALSE,

    -- Dates
    sent_at TIMESTAMP WITH TIME ZONE,
    edited_at TIMESTAMP WITH TIME ZONE,
    synced_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),

    UNIQUE(channel_id, external_id)
);

CREATE INDEX IF NOT EXISTS idx_channel_messages_channel ON channel_messages(channel_id, sent_at DESC);
CREATE INDEX IF NOT EXISTS idx_channel_messages_thread ON channel_messages(channel_id, thread_ts);

-- ============================================================================
-- INTEGRATION SYNC LOG (from migration 030)
-- ============================================================================

CREATE TABLE IF NOT EXISTS integration_sync_log (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(255) NOT NULL,
    provider_id VARCHAR(100) NOT NULL,
    sync_type VARCHAR(50) NOT NULL,
    status VARCHAR(50) NOT NULL,

    -- Sync details
    records_synced INT DEFAULT 0,
    records_failed INT DEFAULT 0,
    error_message TEXT,

    -- Timing
    started_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    completed_at TIMESTAMP WITH TIME ZONE,
    duration_ms INT,

    -- Metadata
    metadata JSONB DEFAULT '{}'
);

CREATE INDEX IF NOT EXISTS idx_sync_log_user_provider ON integration_sync_log(user_id, provider_id, started_at DESC);

-- ============================================================================
-- NOTION DATABASES
-- ============================================================================

CREATE TABLE IF NOT EXISTS notion_databases (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(255) NOT NULL,
    notion_id VARCHAR(255) NOT NULL,

    -- Database metadata
    title VARCHAR(500),
    description TEXT,
    icon VARCHAR(500),
    cover VARCHAR(500),
    url VARCHAR(500),

    -- Properties schema (stored as JSONB)
    properties JSONB DEFAULT '{}',

    -- Sync tracking
    synced_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),

    UNIQUE(user_id, notion_id)
);

CREATE INDEX IF NOT EXISTS idx_notion_databases_user ON notion_databases(user_id);
CREATE INDEX IF NOT EXISTS idx_notion_databases_title ON notion_databases(user_id, title);

-- ============================================================================
-- NOTION PAGES
-- ============================================================================

CREATE TABLE IF NOT EXISTS notion_pages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(255) NOT NULL,
    notion_id VARCHAR(255) NOT NULL,
    database_id UUID REFERENCES notion_databases(id) ON DELETE SET NULL,

    -- Page metadata
    title VARCHAR(500),
    icon VARCHAR(500),
    cover VARCHAR(500),
    url VARCHAR(500),
    archived BOOLEAN DEFAULT FALSE,

    -- Properties (from database schema)
    properties JSONB DEFAULT '{}',

    -- Content (optional - for full page sync)
    content JSONB DEFAULT '[]',

    -- Sync tracking
    synced_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),

    UNIQUE(user_id, notion_id)
);

CREATE INDEX IF NOT EXISTS idx_notion_pages_user ON notion_pages(user_id);
CREATE INDEX IF NOT EXISTS idx_notion_pages_database ON notion_pages(database_id);
CREATE INDEX IF NOT EXISTS idx_notion_pages_title ON notion_pages(user_id, title);
CREATE INDEX IF NOT EXISTS idx_notion_pages_archived ON notion_pages(user_id, archived);

-- ============================================================================
-- SLACK CHANNELS (Slack-specific)
-- ============================================================================

CREATE TABLE IF NOT EXISTS slack_channels (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(255) NOT NULL,
    slack_id VARCHAR(255) NOT NULL,

    -- Channel metadata
    name VARCHAR(255) NOT NULL,
    is_private BOOLEAN DEFAULT FALSE,
    is_dm BOOLEAN DEFAULT FALSE,
    is_mpim BOOLEAN DEFAULT FALSE,
    member_count INT DEFAULT 0,
    topic TEXT,
    purpose TEXT,
    unread_count INT DEFAULT 0,

    -- Activity tracking
    last_activity TIMESTAMP WITH TIME ZONE,

    -- Sync tracking
    synced_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),

    UNIQUE(user_id, slack_id)
);

CREATE INDEX IF NOT EXISTS idx_slack_channels_user ON slack_channels(user_id);
CREATE INDEX IF NOT EXISTS idx_slack_channels_name ON slack_channels(user_id, name);
CREATE INDEX IF NOT EXISTS idx_slack_channels_activity ON slack_channels(user_id, last_activity DESC);

-- ============================================================================
-- SLACK MESSAGES
-- ============================================================================

CREATE TABLE IF NOT EXISTS slack_messages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(255) NOT NULL,
    channel_id UUID NOT NULL REFERENCES slack_channels(id) ON DELETE CASCADE,
    slack_ts VARCHAR(50) NOT NULL,

    -- Sender info
    sender_id VARCHAR(255),
    sender_name VARCHAR(255),

    -- Message content
    content TEXT,
    thread_ts VARCHAR(50),
    reply_count INT DEFAULT 0,
    is_edited BOOLEAN DEFAULT FALSE,

    -- Timestamps
    sent_at TIMESTAMP WITH TIME ZONE,
    synced_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),

    UNIQUE(user_id, channel_id, slack_ts)
);

CREATE INDEX IF NOT EXISTS idx_slack_messages_channel ON slack_messages(channel_id, sent_at DESC);
CREATE INDEX IF NOT EXISTS idx_slack_messages_thread ON slack_messages(channel_id, thread_ts);
CREATE INDEX IF NOT EXISTS idx_slack_messages_sender ON slack_messages(sender_id);

-- ============================================================================
-- LINEAR ISSUES
-- ============================================================================

CREATE TABLE IF NOT EXISTS linear_issues (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(255) NOT NULL,
    external_id TEXT NOT NULL,
    identifier TEXT NOT NULL,
    title TEXT NOT NULL,
    description TEXT,
    state TEXT NOT NULL DEFAULT 'backlog',
    priority INTEGER DEFAULT 0,
    assignee TEXT,
    project TEXT,
    team TEXT NOT NULL,
    due_date DATE,
    external_created_at TIMESTAMPTZ,
    external_updated_at TIMESTAMPTZ,
    synced_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(user_id, external_id)
);

CREATE INDEX IF NOT EXISTS idx_linear_issues_user ON linear_issues(user_id);
CREATE INDEX IF NOT EXISTS idx_linear_issues_identifier ON linear_issues(identifier);
CREATE INDEX IF NOT EXISTS idx_linear_issues_state ON linear_issues(user_id, state);
CREATE INDEX IF NOT EXISTS idx_linear_issues_team ON linear_issues(user_id, team);
CREATE INDEX IF NOT EXISTS idx_linear_issues_updated ON linear_issues(external_updated_at DESC);

-- ============================================================================
-- LINEAR PROJECTS
-- ============================================================================

CREATE TABLE IF NOT EXISTS linear_projects (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(255) NOT NULL,
    external_id TEXT NOT NULL,
    name TEXT NOT NULL,
    description TEXT,
    state TEXT NOT NULL DEFAULT 'planned',
    progress DECIMAL(5,2) DEFAULT 0,
    start_date DATE,
    target_date DATE,
    team TEXT,
    synced_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(user_id, external_id)
);

CREATE INDEX IF NOT EXISTS idx_linear_projects_user ON linear_projects(user_id);
CREATE INDEX IF NOT EXISTS idx_linear_projects_state ON linear_projects(user_id, state);

-- ============================================================================
-- LINEAR TEAMS
-- ============================================================================

CREATE TABLE IF NOT EXISTS linear_teams (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(255) NOT NULL,
    external_id TEXT NOT NULL,
    key TEXT NOT NULL,
    name TEXT NOT NULL,
    description TEXT,
    issue_count INTEGER DEFAULT 0,
    synced_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(user_id, external_id)
);

CREATE INDEX IF NOT EXISTS idx_linear_teams_user ON linear_teams(user_id);
CREATE INDEX IF NOT EXISTS idx_linear_teams_key ON linear_teams(user_id, key);

-- Update timestamp triggers for new tables
CREATE OR REPLACE FUNCTION update_linear_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DO $$
DECLARE
    tables TEXT[] := ARRAY[
        'emails', 'channels', 'channel_messages', 'integration_sync_log',
        'notion_databases', 'notion_pages', 'slack_channels', 'slack_messages',
        'linear_issues', 'linear_projects', 'linear_teams'
    ];
    t TEXT;
BEGIN
    FOREACH t IN ARRAY tables
    LOOP
        EXECUTE format('DROP TRIGGER IF EXISTS %I_updated_at ON %I', t, t);
        BEGIN
            EXECUTE format('CREATE TRIGGER %I_updated_at BEFORE UPDATE ON %I FOR EACH ROW EXECUTE FUNCTION update_integration_updated_at()', t, t);
        EXCEPTION WHEN undefined_function THEN
            EXECUTE format('CREATE TRIGGER %I_updated_at BEFORE UPDATE ON %I FOR EACH ROW EXECUTE FUNCTION update_linear_updated_at()', t, t);
        END;
    END LOOP;
END $$;

-- Comments for new tables
COMMENT ON TABLE emails IS 'Synced emails from Gmail and other providers';
COMMENT ON TABLE channels IS 'Synced communication channels from Slack/Discord/Teams';
COMMENT ON TABLE channel_messages IS 'Messages within communication channels';
COMMENT ON TABLE integration_sync_log IS 'Log of all integration sync operations';
COMMENT ON TABLE notion_databases IS 'Synced Notion databases with their property schemas';
COMMENT ON TABLE notion_pages IS 'Synced Notion pages/database entries';
COMMENT ON TABLE slack_channels IS 'Synced Slack channels, DMs, and group messages';
COMMENT ON TABLE slack_messages IS 'Synced Slack messages from channels';
COMMENT ON TABLE linear_issues IS 'Synced Linear issues';
COMMENT ON TABLE linear_projects IS 'Synced Linear projects';
COMMENT ON TABLE linear_teams IS 'Synced Linear teams';
