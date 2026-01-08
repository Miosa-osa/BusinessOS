-- Consolidated Migration: All missing tables from 003-020
-- Generated: 2026-01-07

-- Migration 003: Custom project statuses
-- Date: 2025-12-21
-- Description: Adds customizable statuses per project for tasks

-- Create project_statuses table
CREATE TABLE IF NOT EXISTS project_statuses (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    name VARCHAR(50) NOT NULL,
    color VARCHAR(7) NOT NULL DEFAULT '#6B7280',
    position INTEGER NOT NULL DEFAULT 0,
    is_done_state BOOLEAN DEFAULT FALSE,
    is_default BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(project_id, name)
);

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_project_statuses_project_id ON project_statuses(project_id);
CREATE INDEX IF NOT EXISTS idx_project_statuses_position ON project_statuses(project_id, position);

-- Add custom_status_id column to tasks
ALTER TABLE tasks ADD COLUMN IF NOT EXISTS custom_status_id UUID REFERENCES project_statuses(id) ON DELETE SET NULL;

-- CREATE INDEX IF NOT EXISTS for custom_status_id
CREATE INDEX IF NOT EXISTS idx_tasks_custom_status ON tasks(custom_status_id);

-- Function to create default statuses for a project
CREATE OR REPLACE FUNCTION create_default_project_statuses(p_project_id UUID) RETURNS VOID AS $$
BEGIN
    INSERT INTO project_statuses (project_id, name, color, position, is_done_state, is_default)
    VALUES
        (p_project_id, 'To Do', '#6B7280', 0, FALSE, TRUE),
        (p_project_id, 'In Progress', '#3B82F6', 1, FALSE, FALSE),
        (p_project_id, 'In Review', '#8B5CF6', 2, FALSE, FALSE),
        (p_project_id, 'Done', '#10B981', 3, TRUE, FALSE),
        (p_project_id, 'Blocked', '#F59E0B', 4, FALSE, FALSE)
    ON CONFLICT (project_id, name) DO NOTHING;
END;
$$ LANGUAGE plpgsql;

-- Trigger function to auto-create default statuses on project creation
CREATE OR REPLACE FUNCTION auto_create_project_statuses() RETURNS TRIGGER AS $$
BEGIN
    PERFORM create_default_project_statuses(NEW.id);
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create trigger for new projects
DROP TRIGGER IF EXISTS create_project_statuses_trigger ON projects;
DROP TRIGGER IF EXISTS create_project_statuses_trigger ON projects;
CREATE TRIGGER create_project_statuses_trigger
    AFTER INSERT ON projects
    FOR EACH ROW EXECUTE FUNCTION auto_create_project_statuses();

-- Create default statuses for existing projects that don't have any
DO $$
DECLARE
    proj_id UUID;
BEGIN
    FOR proj_id IN
        SELECT p.id FROM projects p
        WHERE NOT EXISTS (
            SELECT 1 FROM project_statuses ps WHERE ps.project_id = p.id
        )
    LOOP
        PERFORM create_default_project_statuses(proj_id);
    END LOOP;
END;
$$;
-- Migration 004: Multiple assignees per task
-- Date: 2025-12-21
-- Description: Adds support for assigning multiple team members to a single task

-- Create task_assignees junction table
CREATE TABLE IF NOT EXISTS task_assignees (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    task_id UUID NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    team_member_id UUID NOT NULL REFERENCES team_members(id) ON DELETE CASCADE,
    role VARCHAR(50) DEFAULT 'assignee', -- 'assignee', 'reviewer', 'lead'
    assigned_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    assigned_by VARCHAR(255),
    UNIQUE(task_id, team_member_id)
);

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_task_assignees_task ON task_assignees(task_id);
CREATE INDEX IF NOT EXISTS idx_task_assignees_member ON task_assignees(team_member_id);
CREATE INDEX IF NOT EXISTS idx_task_assignees_role ON task_assignees(role);

-- Migrate existing single assignee data to new junction table
-- This preserves existing assignments
INSERT INTO task_assignees (task_id, team_member_id, role, assigned_at)
SELECT id, assignee_id, 'assignee', updated_at
FROM tasks
WHERE assignee_id IS NOT NULL
ON CONFLICT (task_id, team_member_id) DO NOTHING;
-- Migration 005: Task Dependencies for Gantt View
-- Date: 2025-12-21
-- Description: Adds support for task dependencies (predecessor/successor relationships)
-- Used for Gantt chart visualization and dependency tracking

-- Create task_dependencies table
CREATE TABLE IF NOT EXISTS task_dependencies (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    predecessor_id UUID NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    successor_id UUID NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    dependency_type VARCHAR(20) DEFAULT 'finish_to_start',  -- 'finish_to_start', 'start_to_start', 'finish_to_finish', 'start_to_finish'
    lag_days INTEGER DEFAULT 0,  -- Delay between predecessor end and successor start
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(predecessor_id, successor_id),
    -- Prevent self-referential dependencies
    CONSTRAINT no_self_dependency CHECK (predecessor_id != successor_id)
);

-- Create indexes for efficient lookups
CREATE INDEX IF NOT EXISTS idx_task_dependencies_predecessor ON task_dependencies(predecessor_id);
CREATE INDEX IF NOT EXISTS idx_task_dependencies_successor ON task_dependencies(successor_id);
CREATE INDEX IF NOT EXISTS idx_task_dependencies_type ON task_dependencies(dependency_type);

-- Create function to detect circular dependencies
CREATE OR REPLACE FUNCTION check_circular_dependency()
RETURNS TRIGGER AS $$
DECLARE
    has_cycle BOOLEAN;
BEGIN
    -- Use recursive CTE to detect cycles
    WITH RECURSIVE dependency_chain AS (
        -- Start with the successor of the new dependency
        SELECT successor_id AS task_id, 1 AS depth
        FROM task_dependencies
        WHERE predecessor_id = NEW.successor_id

        UNION ALL

        -- Follow the chain
        SELECT td.successor_id, dc.depth + 1
        FROM dependency_chain dc
        JOIN task_dependencies td ON td.predecessor_id = dc.task_id
        WHERE dc.depth < 100  -- Prevent infinite loops with depth limit
    )
    SELECT EXISTS (
        SELECT 1 FROM dependency_chain WHERE task_id = NEW.predecessor_id
    ) INTO has_cycle;

    IF has_cycle THEN
        RAISE EXCEPTION 'Circular dependency detected: this would create a dependency loop';
    END IF;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create trigger to prevent circular dependencies
DROP TRIGGER IF EXISTS prevent_circular_dependencies ON task_dependencies;
DROP TRIGGER IF EXISTS prevent_circular_dependencies ON task_dependencies;
CREATE TRIGGER prevent_circular_dependencies
    BEFORE INSERT OR UPDATE ON task_dependencies
    FOR EACH ROW
    EXECUTE FUNCTION check_circular_dependency();
-- Migration 006: Vector Embeddings for Knowledge Base RAG
-- Requires pgvector extension to be installed in PostgreSQL

-- Enable pgvector extension
CREATE EXTENSION IF NOT EXISTS vector;

-- Document embeddings table (block-level granularity)
CREATE TABLE IF NOT EXISTS context_embeddings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    context_id UUID NOT NULL REFERENCES contexts(id) ON DELETE CASCADE,
    block_id TEXT NOT NULL,                    -- Editor block ID
    block_type TEXT NOT NULL,                  -- paragraph, heading, list, etc.
    content TEXT NOT NULL,                     -- Raw text content
    embedding vector(768),                     -- nomic-embed-text dimension
    metadata JSONB DEFAULT '{}',               -- Additional block metadata
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),

    UNIQUE(context_id, block_id)
);

-- Create HNSW index for fast cosine similarity search
-- HNSW provides ~10x faster queries than IVFFlat with similar recall
CREATE INDEX IF NOT EXISTS idx_context_embeddings_vector
ON context_embeddings USING hnsw (embedding vector_cosine_ops);

-- Index for context lookups
CREATE INDEX IF NOT EXISTS idx_context_embeddings_context_id
ON context_embeddings(context_id);

-- Index for block type filtering
CREATE INDEX IF NOT EXISTS idx_context_embeddings_block_type
ON context_embeddings(block_type);

-- Add embedding tracking columns to contexts table
ALTER TABLE contexts ADD COLUMN IF NOT EXISTS embedding_status TEXT DEFAULT 'pending';
ALTER TABLE contexts ADD COLUMN IF NOT EXISTS last_embedded_at TIMESTAMPTZ;
ALTER TABLE contexts ADD COLUMN IF NOT EXISTS embedding_count INTEGER DEFAULT 0;

-- CREATE INDEX IF NOT EXISTS on embedding status for batch operations
CREATE INDEX IF NOT EXISTS idx_contexts_embedding_status
ON contexts(embedding_status) WHERE embedding_status != 'indexed';
-- Add client_id column to projects table
-- This allows projects to be linked to clients

ALTER TABLE projects ADD COLUMN IF NOT EXISTS client_id UUID REFERENCES clients(id) ON DELETE SET NULL;
CREATE INDEX IF NOT EXISTS idx_projects_client ON projects(client_id);
-- Migration: Add share_calendar and calendar_user_id columns to team_members
-- These columns are required for calendar sharing functionality

-- Add share_calendar column if it doesn't exist
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'team_members' AND column_name = 'share_calendar'
    ) THEN
        ALTER TABLE team_members ADD COLUMN share_calendar BOOLEAN DEFAULT FALSE;
    END IF;
END $$;

-- Add calendar_user_id column if it doesn't exist
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'team_members' AND column_name = 'calendar_user_id'
    ) THEN
        ALTER TABLE team_members ADD COLUMN calendar_user_id VARCHAR(255);
    END IF;
END $$;


-- Thinking/reasoning tracking
CREATE TABLE IF NOT EXISTS thinking_traces (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL, 
    conversation_id UUID NOT NULL, 
    message_id UUID NOT NULL,

    -- Thinking content
    thinking_content TEXT NOT NULL,        -- The actual thinking/reasoning text
    thinking_type VARCHAR(50),             -- 'analysis', 'planning', 'reflection', 'tool_use'
    step_number INT,                       -- Order in the thinking chain

    -- Timing
    started_at TIMESTAMPTZ DEFAULT NOW(),
    completed_at TIMESTAMPTZ,
    duration_ms INT,

    -- Token tracking
    thinking_tokens INT DEFAULT 0,

    -- Metadata
    model_used VARCHAR(100),
    reasoning_template_id UUID,            -- If using a custom template
    metadata JSONB DEFAULT '{}',

    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Custom reasoning templates/systems
CREATE TABLE IF NOT EXISTS reasoning_templates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,

    name VARCHAR(255) NOT NULL,
    description TEXT,

    -- Template configuration
    system_prompt TEXT,                    -- Base system prompt for reasoning
    thinking_instruction TEXT,             -- How to structure thinking
    output_format VARCHAR(50),             -- 'streaming', 'collapsed', 'step_by_step'

    -- Options
    show_thinking BOOLEAN DEFAULT true,    -- Show thinking in UI
    save_thinking BOOLEAN DEFAULT true,    -- Save to database
    max_thinking_tokens INT DEFAULT 4096,

    -- Usage tracking
    times_used INT DEFAULT 0,

    is_default BOOLEAN DEFAULT false,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);
-- Migration 009: Custom Agents System
-- Allows users to create and configure custom AI agents

-- ===== CUSTOM AGENTS =====

-- User-defined custom agents with custom system prompts and configurations
CREATE TABLE IF NOT EXISTS custom_agents (
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

CREATE INDEX IF NOT EXISTS idx_custom_agents_user_id ON custom_agents(user_id);
CREATE INDEX IF NOT EXISTS idx_custom_agents_name ON custom_agents(user_id, name);
CREATE INDEX IF NOT EXISTS idx_custom_agents_category ON custom_agents(category);
CREATE INDEX IF NOT EXISTS idx_custom_agents_active ON custom_agents(user_id, is_active);

-- ===== AGENT PRESETS (Optional Built-in Templates) =====

-- Store commonly used agent templates that users can copy
CREATE TABLE IF NOT EXISTS agent_presets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    -- Preset Identity
    name VARCHAR(50) NOT NULL UNIQUE,
    display_name VARCHAR(100) NOT NULL,
    description TEXT,
    avatar VARCHAR(50),

    -- Preset Configuration (same as custom_agents)
    system_prompt TEXT NOT NULL,
    model_preference VARCHAR(100),
    temperature DECIMAL(3,2) DEFAULT 0.7,
    max_tokens INTEGER DEFAULT 4096,
    capabilities TEXT[] DEFAULT '{}',
    tools_enabled TEXT[] DEFAULT '{}',
    context_sources TEXT[] DEFAULT '{}',
    thinking_enabled BOOLEAN DEFAULT FALSE,
    category VARCHAR(50) DEFAULT 'general',

    -- Usage tracking
    times_copied INTEGER DEFAULT 0,

    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Insert default agent presets
INSERT INTO agent_presets (name, display_name, description, avatar, system_prompt, category, capabilities, tools_enabled, thinking_enabled)
VALUES
    ('code-reviewer', 'Code Reviewer', 'Reviews code for bugs, best practices, and improvements', 'magnifying-glass',
     'You are an expert code reviewer. Analyze code for:
1. **Bugs & Errors**: Identify potential bugs, edge cases, and runtime errors
2. **Best Practices**: Check adherence to coding standards and conventions
3. **Performance**: Spot inefficiencies and suggest optimizations
4. **Security**: Flag potential security vulnerabilities
5. **Maintainability**: Assess code readability and suggest improvements

Provide specific, actionable feedback with code examples when suggesting changes.',
     'coding', ARRAY['code_review', 'analysis'], ARRAY['read_file', 'search_code'], TRUE),

    ('technical-writer', 'Technical Writer', 'Creates clear documentation and technical content', 'pencil',
     'You are an expert technical writer. Create clear, well-structured documentation that:
1. Uses simple, precise language
2. Includes relevant code examples
3. Follows standard documentation patterns
4. Anticipates reader questions
5. Provides both quick-start guides and detailed references

Adapt your writing style to the audience - from beginner-friendly tutorials to expert reference docs.',
     'writing', ARRAY['documentation', 'writing'], ARRAY[]::text[], FALSE),

    ('data-analyst', 'Data Analyst', 'Analyzes data and creates insights', 'chart',
     'You are an expert data analyst. When analyzing data:
1. Start with exploratory analysis to understand the data
2. Identify key patterns, trends, and anomalies
3. Use appropriate statistical methods
4. Create clear visualizations (describe them in detail)
5. Provide actionable insights and recommendations

Be precise with numbers and transparent about limitations or assumptions.',
     'analysis', ARRAY['data_analysis', 'visualization'], ARRAY[]::text[], TRUE),

    ('business-strategist', 'Business Strategist', 'Provides strategic business advice and analysis', 'briefcase',
     'You are a senior business strategist. Provide strategic advice by:
1. Understanding the business context and objectives
2. Analyzing market conditions and competition
3. Identifying opportunities and risks
4. Developing actionable recommendations
5. Considering implementation feasibility

Use frameworks like SWOT, Porter''s Five Forces, and business model canvas when appropriate.',
     'business', ARRAY['strategy', 'analysis', 'planning'], ARRAY[]::text[], TRUE),

    ('creative-writer', 'Creative Writer', 'Helps with creative writing and content creation', 'sparkles',
     'You are a talented creative writer. Help with:
1. Generating creative ideas and concepts
2. Writing engaging narratives and copy
3. Developing compelling characters and stories
4. Crafting persuasive marketing content
5. Editing and improving existing content

Match the desired tone, style, and voice. Be creative while staying on-brand.',
     'writing', ARRAY['creative_writing', 'content_creation'], ARRAY[]::text[], FALSE)
ON CONFLICT (name) DO NOTHING;
-- Migration 010: Custom Commands System
-- Maps /slash commands to agents or prompt templates

-- ===== CUSTOM COMMANDS =====

-- Command registry mapping /triggers to actions
CREATE TABLE IF NOT EXISTS custom_commands (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(255) NOT NULL,

    -- Command Identity
    trigger VARCHAR(50) NOT NULL,           -- e.g., "/review" (must start with /)
    display_name VARCHAR(100) NOT NULL,     -- e.g., "Code Review"
    description TEXT,                       -- What the command does
    
    -- Command Action
    action_type VARCHAR(50) NOT NULL,       -- 'agent', 'template', 'tool'
    target_agent_id UUID,                   -- FK to custom_agents (if action_type = 'agent')
    prompt_template TEXT,                   -- Template with {{placeholders}} (if action_type = 'template')
    tool_name VARCHAR(100),                 -- Tool to execute (if action_type = 'tool')
    
    -- Command Behavior
    requires_input BOOLEAN DEFAULT FALSE,   -- Whether command needs user input after trigger
    input_placeholder TEXT,                 -- Placeholder text for input (e.g., "Enter code to review...")
    
    -- Command Configuration
    parameters JSONB DEFAULT '{}',          -- Additional configuration
    streaming_enabled BOOLEAN DEFAULT TRUE, -- Enable streaming for this command
    thinking_enabled BOOLEAN DEFAULT FALSE, -- Enable COT for this command
    
    -- Metadata
    category VARCHAR(50) DEFAULT 'general', -- general, coding, writing, analysis, productivity
    is_active BOOLEAN DEFAULT TRUE,
    is_system BOOLEAN DEFAULT FALSE,        -- System commands cannot be deleted by users
    times_used INTEGER DEFAULT 0,
    last_used_at TIMESTAMP WITH TIME ZONE,
    
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    UNIQUE(user_id, trigger)
);

CREATE INDEX IF NOT EXISTS idx_custom_commands_user_id ON custom_commands(user_id);
CREATE INDEX IF NOT EXISTS idx_custom_commands_trigger ON custom_commands(user_id, trigger);
CREATE INDEX IF NOT EXISTS idx_custom_commands_active ON custom_commands(user_id, is_active);
CREATE INDEX IF NOT EXISTS idx_custom_commands_category ON custom_commands(category);

-- Add FK constraint
ALTER TABLE custom_commands 
ADD CONSTRAINT fk_custom_commands_agent 
FOREIGN KEY (target_agent_id) REFERENCES custom_agents(id) ON DELETE SET NULL;

-- ===== AGENT MENTIONS TRACKING =====

-- Track @agent mentions in messages for context
CREATE TABLE IF NOT EXISTS agent_mentions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(255) NOT NULL,
    conversation_id UUID NOT NULL,
    message_id UUID NOT NULL,
    
    -- Mention details
    mentioned_agent_id UUID,                -- FK to custom_agents
    mention_text VARCHAR(100) NOT NULL,     -- The actual @mention (e.g., "@code-reviewer")
    position_in_message INT,                -- Character position of mention
    
    -- Resolution
    resolved BOOLEAN DEFAULT TRUE,          -- Whether agent was invoked
    resolution_note TEXT,                   -- Why it failed (if any)
    
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_agent_mentions_user_id ON agent_mentions(user_id);
CREATE INDEX IF NOT EXISTS idx_agent_mentions_conversation ON agent_mentions(conversation_id);
CREATE INDEX IF NOT EXISTS idx_agent_mentions_agent ON agent_mentions(mentioned_agent_id);

-- Add FK constraint
ALTER TABLE agent_mentions 
ADD CONSTRAINT fk_agent_mentions_agent 
FOREIGN KEY (mentioned_agent_id) REFERENCES custom_agents(id) ON DELETE CASCADE;

-- ===== SEED SYSTEM COMMANDS =====

-- Insert default system commands (placeholder - will reference actual agents after they're created)
INSERT INTO custom_commands (user_id, trigger, display_name, description, action_type, prompt_template, category, is_system, streaming_enabled, thinking_enabled)
VALUES
    ('SYSTEM', '/help', 'Show Help', 'Display available commands and agents', 'template', 
     'Here are the available commands and agents:\n\n**Commands:**\n{{commands_list}}\n\n**Agents:**\n{{agents_list}}', 
     'productivity', TRUE, TRUE, FALSE),
    
    ('SYSTEM', '/clear', 'Clear Context', 'Clear conversation context', 'tool', 
     NULL, 
     'productivity', TRUE, FALSE, FALSE),
    
    ('SYSTEM', '/summarize', 'Summarize Conversation', 'Create a summary of the current conversation', 'template',
     'Please provide a concise summary of this conversation, highlighting:\n1. Key topics discussed\n2. Decisions made\n3. Action items identified\n4. Open questions remaining',
     'productivity', TRUE, TRUE, TRUE)
ON CONFLICT (user_id, trigger) DO NOTHING;
-- Migration 013: Focus Configurations System
-- Stores user-specific overrides for Quick, Deep, and Creative modes

-- Focus configuration templates (system-level defaults)
CREATE TABLE IF NOT EXISTS focus_mode_templates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(50) NOT NULL UNIQUE,
    display_name VARCHAR(100) NOT NULL,
    description TEXT,
    icon VARCHAR(50),

    -- Behavior settings
    default_model VARCHAR(100),
    temperature DECIMAL(3,2) DEFAULT 0.7,
    max_tokens INTEGER DEFAULT 4096,

    -- Output constraints
    output_style VARCHAR(50) DEFAULT 'balanced', -- concise, balanced, detailed, structured
    response_format VARCHAR(50) DEFAULT 'markdown', -- markdown, plain, json, artifact
    max_response_length INTEGER, -- null = no limit
    require_sources BOOLEAN DEFAULT false,

    -- Context settings
    auto_search BOOLEAN DEFAULT false,
    search_depth VARCHAR(20) DEFAULT 'quick', -- quick, standard, deep
    kb_context_limit INTEGER DEFAULT 5, -- max KB items to inject
    include_history_count INTEGER DEFAULT 10, -- conversation history to include

    -- Thinking/COT settings
    thinking_enabled BOOLEAN DEFAULT false,
    thinking_style VARCHAR(50), -- analytical, creative, step-by-step

    -- System prompt additions
    system_prompt_prefix TEXT,
    system_prompt_suffix TEXT,

    -- Metadata
    sort_order INTEGER DEFAULT 0,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- User-specific focus configuration overrides
CREATE TABLE IF NOT EXISTS focus_configurations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(255) NOT NULL,
    template_id UUID REFERENCES focus_mode_templates(id) ON DELETE CASCADE,

    -- Override settings (null = use template default)
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

    -- Preferred model override
    preferred_model VARCHAR(100),

    -- Auto-load KB categories
    auto_load_kb_categories TEXT[], -- array of category slugs to auto-include

    -- Keyboard shortcut
    keyboard_shortcut VARCHAR(20),

    -- Metadata
    is_favorite BOOLEAN DEFAULT false,
    use_count INTEGER DEFAULT 0,
    last_used_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),

    UNIQUE(user_id, template_id)
);

-- Focus mode context presets (for auto-load)
CREATE TABLE IF NOT EXISTS focus_context_presets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(255) NOT NULL,
    name VARCHAR(100) NOT NULL,
    description TEXT,

    -- What to include
    kb_artifact_ids UUID[], -- specific artifacts to include
    kb_categories TEXT[], -- category slugs
    project_ids UUID[], -- projects to include context from

    -- Search settings
    default_search_queries TEXT[], -- pre-defined searches to run
    search_domains TEXT[], -- domains to search

    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Link presets to focus configurations
CREATE TABLE IF NOT EXISTS focus_configuration_presets (
    focus_config_id UUID REFERENCES focus_configurations(id) ON DELETE CASCADE,
    preset_id UUID REFERENCES focus_context_presets(id) ON DELETE CASCADE,
    sort_order INTEGER DEFAULT 0,
    PRIMARY KEY (focus_config_id, preset_id)
);

-- Indexes
CREATE INDEX IF NOT EXISTS idx_focus_configurations_user ON focus_configurations(user_id);
CREATE INDEX IF NOT EXISTS idx_focus_configurations_template ON focus_configurations(template_id);
CREATE INDEX IF NOT EXISTS idx_focus_context_presets_user ON focus_context_presets(user_id);

-- Insert default focus mode templates
INSERT INTO focus_mode_templates (name, display_name, description, icon, default_model, temperature, output_style, auto_search, search_depth, thinking_enabled, thinking_style, system_prompt_prefix, sort_order) VALUES
('quick', 'Quick', 'Fast, concise responses for simple questions', 'zap', NULL, 0.5, 'concise', false, 'quick', false, NULL,
'You are in Quick Mode. Provide brief, direct answers. Be concise and to the point. Avoid unnecessary elaboration.', 1),

('deep', 'Deep Research', 'Thorough research with sources and citations', 'search', 'claude-sonnet-4-20250514', 0.7, 'detailed', true, 'deep', true, 'analytical',
'You are in Deep Research Mode. Conduct thorough research and provide comprehensive, well-sourced answers. Include citations where possible. Analyze multiple perspectives.', 2),

('creative', 'Creative', 'Imaginative and exploratory responses', 'sparkles', NULL, 0.9, 'balanced', false, 'quick', true, 'creative',
'You are in Creative Mode. Think outside the box. Explore unconventional ideas and approaches. Be imaginative and innovative in your responses.', 3),

('analyze', 'Analysis', 'Data-driven analysis and insights', 'chart-bar', 'claude-sonnet-4-20250514', 0.6, 'structured', false, 'standard', true, 'analytical',
'You are in Analysis Mode. Focus on data-driven insights. Structure your response with clear sections. Use quantitative reasoning where applicable.', 4),

('write', 'Writing', 'Document creation and editing', 'file-text', NULL, 0.7, 'detailed', false, 'quick', false, NULL,
'You are in Writing Mode. Create well-structured, polished content. Focus on clarity, flow, and appropriate tone. Generate artifacts for longer documents.', 5),

('plan', 'Planning', 'Strategic planning and project organization', 'clipboard-list', NULL, 0.6, 'structured', false, 'standard', true, 'step-by-step',
'You are in Planning Mode. Create actionable plans with clear steps. Consider dependencies and timelines. Structure output as organized lists or project artifacts.', 6),

('code', 'Coding', 'Software development assistance', 'code', 'claude-sonnet-4-20250514', 0.4, 'structured', false, 'quick', true, 'step-by-step',
'You are in Coding Mode. Write clean, efficient code. Follow best practices. Include comments where helpful. Generate code artifacts for complete implementations.', 7)

ON CONFLICT (name) DO UPDATE SET
    display_name = EXCLUDED.display_name,
    description = EXCLUDED.description,
    icon = EXCLUDED.icon,
    default_model = EXCLUDED.default_model,
    temperature = EXCLUDED.temperature,
    output_style = EXCLUDED.output_style,
    auto_search = EXCLUDED.auto_search,
    search_depth = EXCLUDED.search_depth,
    thinking_enabled = EXCLUDED.thinking_enabled,
    thinking_style = EXCLUDED.thinking_style,
    system_prompt_prefix = EXCLUDED.system_prompt_prefix,
    sort_order = EXCLUDED.sort_order,
    updated_at = NOW();
-- Migration 014: Web Search Results Cache
-- Prevents redundant API calls within the same conversation or time window

-- Cache table for web search results
CREATE TABLE IF NOT EXISTS web_search_results (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    -- Query identification
    query_hash VARCHAR(64) NOT NULL, -- SHA256 hash of normalized query
    original_query TEXT NOT NULL,
    optimized_query TEXT, -- Query after optimization

    -- Context (optional)
    user_id VARCHAR(255),
    conversation_id UUID,

    -- Results
    results JSONB NOT NULL DEFAULT '[]', -- Array of search results
    result_count INTEGER DEFAULT 0,
    provider VARCHAR(50) DEFAULT 'duckduckgo',

    -- Performance metrics
    search_time_ms FLOAT,

    -- Cache management
    expires_at TIMESTAMPTZ NOT NULL, -- When this cache entry expires
    hit_count INTEGER DEFAULT 0, -- Number of times this cache was used
    last_hit_at TIMESTAMPTZ,

    -- Metadata
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Index for fast query hash lookups
CREATE INDEX IF NOT EXISTS idx_web_search_query_hash ON web_search_results(query_hash);

-- Index for user-specific cache lookups
CREATE INDEX IF NOT EXISTS idx_web_search_user ON web_search_results(user_id) WHERE user_id IS NOT NULL;

-- Index for conversation-specific cache
CREATE INDEX IF NOT EXISTS idx_web_search_conversation ON web_search_results(conversation_id) WHERE conversation_id IS NOT NULL;

-- Index for cache expiration cleanup
CREATE INDEX IF NOT EXISTS idx_web_search_expires ON web_search_results(expires_at);

-- Composite index for most common lookup pattern
CREATE INDEX IF NOT EXISTS idx_web_search_lookup ON web_search_results(query_hash, expires_at);

-- Function to clean expired cache entries (call periodically)
CREATE OR REPLACE FUNCTION cleanup_expired_search_cache()
RETURNS INTEGER AS $$
DECLARE
    deleted_count INTEGER;
BEGIN
    DELETE FROM web_search_results WHERE expires_at < NOW();
    GET DIAGNOSTICS deleted_count = ROW_COUNT;
    RETURN deleted_count;
END;
$$ LANGUAGE plpgsql;

-- Add comment explaining the table
COMMENT ON TABLE web_search_results IS 'Cache for web search results to prevent redundant API calls';
COMMENT ON COLUMN web_search_results.query_hash IS 'SHA256 hash of normalized (lowercased, trimmed) query for fast lookups';
COMMENT ON COLUMN web_search_results.expires_at IS 'Cache entries expire after a configurable duration (default 1 hour for general, 15 min for news)';
-- Migration 015: Command Workflows
-- Enables multi-step command sequences with dependencies and parallel execution

-- Workflow definition table
CREATE TABLE IF NOT EXISTS command_workflows (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(255) NOT NULL,

    -- Workflow metadata
    name VARCHAR(255) NOT NULL,
    display_name VARCHAR(255) NOT NULL,
    description TEXT,
    trigger VARCHAR(100) NOT NULL, -- e.g., "/deploy" or "/release"

    -- Execution settings
    execution_mode VARCHAR(50) DEFAULT 'sequential', -- sequential, parallel, smart
    stop_on_failure BOOLEAN DEFAULT TRUE,
    timeout_seconds INTEGER DEFAULT 300,

    -- Status
    is_active BOOLEAN DEFAULT TRUE,
    is_system BOOLEAN DEFAULT FALSE,

    -- Metadata
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Workflow steps table
CREATE TABLE IF NOT EXISTS workflow_steps (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workflow_id UUID NOT NULL REFERENCES command_workflows(id) ON DELETE CASCADE,

    -- Step metadata
    name VARCHAR(255) NOT NULL,
    description TEXT,
    step_order INTEGER NOT NULL,

    -- Step action
    action_type VARCHAR(50) NOT NULL, -- command, agent, tool, condition, wait

    -- For action_type = 'command'
    command_trigger VARCHAR(100),
    command_args TEXT,

    -- For action_type = 'agent'
    target_agent_id UUID,
    prompt_template TEXT,

    -- For action_type = 'tool'
    tool_name VARCHAR(100),
    tool_params JSONB DEFAULT '{}',

    -- For action_type = 'condition'
    condition_expression TEXT, -- e.g., "{{previous.success}} == true"
    on_true_step UUID, -- step to go to if true
    on_false_step UUID, -- step to go to if false

    -- For action_type = 'wait'
    wait_seconds INTEGER DEFAULT 0,

    -- Dependencies
    depends_on UUID[], -- array of step IDs that must complete first
    can_parallel BOOLEAN DEFAULT FALSE,

    -- Error handling
    on_failure VARCHAR(50) DEFAULT 'stop', -- stop, continue, retry, skip
    max_retries INTEGER DEFAULT 0,
    retry_delay_seconds INTEGER DEFAULT 5,

    -- Context
    input_mapping JSONB DEFAULT '{}', -- map input from previous steps
    output_key VARCHAR(100), -- key to store output in context

    -- Metadata
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Workflow execution logs
CREATE TABLE IF NOT EXISTS workflow_executions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workflow_id UUID NOT NULL REFERENCES command_workflows(id) ON DELETE CASCADE,
    user_id VARCHAR(255) NOT NULL,

    -- Execution context
    conversation_id UUID,
    initial_input TEXT,
    context JSONB DEFAULT '{}',

    -- Status
    status VARCHAR(50) DEFAULT 'pending', -- pending, running, completed, failed, cancelled
    current_step_id UUID,

    -- Results
    result JSONB DEFAULT '{}',
    error_message TEXT,

    -- Timing
    started_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,

    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Step execution logs
CREATE TABLE IF NOT EXISTS workflow_step_executions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    execution_id UUID NOT NULL REFERENCES workflow_executions(id) ON DELETE CASCADE,
    step_id UUID NOT NULL REFERENCES workflow_steps(id) ON DELETE CASCADE,

    -- Status
    status VARCHAR(50) DEFAULT 'pending', -- pending, running, completed, failed, skipped
    attempt_number INTEGER DEFAULT 1,

    -- Results
    input JSONB DEFAULT '{}',
    output JSONB DEFAULT '{}',
    error_message TEXT,

    -- Timing
    started_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    duration_ms FLOAT,

    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Indexes
CREATE INDEX IF NOT EXISTS idx_workflows_user ON command_workflows(user_id);
CREATE INDEX IF NOT EXISTS idx_workflows_trigger ON command_workflows(trigger);
CREATE INDEX IF NOT EXISTS idx_workflow_steps_workflow ON workflow_steps(workflow_id);
CREATE INDEX IF NOT EXISTS idx_workflow_steps_order ON workflow_steps(workflow_id, step_order);
CREATE INDEX IF NOT EXISTS idx_workflow_executions_workflow ON workflow_executions(workflow_id);
CREATE INDEX IF NOT EXISTS idx_workflow_executions_user ON workflow_executions(user_id);
CREATE INDEX IF NOT EXISTS idx_workflow_executions_status ON workflow_executions(status);
CREATE INDEX IF NOT EXISTS idx_step_executions_execution ON workflow_step_executions(execution_id);

-- Unique constraint for workflow triggers per user
CREATE UNIQUE INDEX IF NOT EXISTS idx_workflows_user_trigger ON command_workflows(user_id, trigger) WHERE is_active = TRUE;

-- Comments
COMMENT ON TABLE command_workflows IS 'Multi-step command workflows for complex automation';
COMMENT ON TABLE workflow_steps IS 'Individual steps within a workflow';
COMMENT ON TABLE workflow_executions IS 'Execution history and status of workflow runs';
COMMENT ON TABLE workflow_step_executions IS 'Execution history of individual steps';
-- Migration 016: Notifications System
-- Real-time notification system with batching, preferences, and multi-channel delivery

-- ===== NOTIFICATIONS TABLE =====
CREATE TABLE IF NOT EXISTS notifications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(255) NOT NULL,
    workspace_id UUID,  -- Nullable for future workspace support

    -- Notification Content
    type VARCHAR(100) NOT NULL,
    title VARCHAR(255) NOT NULL,
    body TEXT,

    -- Reference to source entity
    entity_type VARCHAR(50),
    entity_id UUID,

    -- Sender info (for social notifications)
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
    workspace_id UUID,  -- Nullable for future workspace support

    -- Global channel toggles
    email_enabled BOOLEAN DEFAULT TRUE,
    push_enabled BOOLEAN DEFAULT TRUE,
    in_app_enabled BOOLEAN DEFAULT TRUE,

    -- Per-type settings (overrides globals)
    -- Example: {"task.assigned": {"email": true, "push": true, "in_app": true}}
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

    -- Unique per user per workspace (NULL workspace = global prefs)
    UNIQUE(user_id, workspace_id)
);

-- Indexes for preferences
CREATE INDEX IF NOT EXISTS idx_notification_prefs_user
    ON notification_preferences(user_id);

-- ===== NOTIFICATION BATCHES TABLE =====
CREATE TABLE IF NOT EXISTS notification_batches (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(255) NOT NULL,

    -- Grouping key (user_id:type:entity_id or user_id:type)
    batch_key VARCHAR(255) NOT NULL,
    type VARCHAR(100) NOT NULL,
    entity_type VARCHAR(50),
    entity_id UUID,

    -- Pending notifications
    pending_ids UUID[] DEFAULT '{}',
    pending_count INTEGER DEFAULT 0,

    -- Timing
    first_at TIMESTAMPTZ DEFAULT NOW(),
    dispatch_at TIMESTAMPTZ NOT NULL,

    -- Status: pending, dispatched
    status VARCHAR(20) DEFAULT 'pending',

    -- Unique batch per user per key
    UNIQUE(user_id, batch_key)
);

-- Indexes for batches
CREATE INDEX IF NOT EXISTS idx_notification_batches_dispatch
    ON notification_batches(dispatch_at) WHERE status = 'pending';
CREATE INDEX IF NOT EXISTS idx_notification_batches_user
    ON notification_batches(user_id);

-- Comments
COMMENT ON TABLE notifications IS 'User notifications with multi-channel delivery support';
COMMENT ON TABLE notification_preferences IS 'Per-user notification channel and timing preferences';
COMMENT ON TABLE notification_batches IS 'Pending notification batches for spam reduction';
COMMENT ON COLUMN notifications.workspace_id IS 'Nullable - for future workspace/team support';
COMMENT ON COLUMN notifications.batch_id IS 'Links to batch if this notification was batched';
COMMENT ON COLUMN notifications.channels_sent IS 'Tracks which channels received this notification';

CREATE TABLE IF NOT EXISTS user_dashboards (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(255) NOT NULL,
    workspace_id UUID REFERENCES workspaces(id) ON DELETE CASCADE,
    
    -- Identity
    name VARCHAR(100) NOT NULL,
    description TEXT,
    is_default BOOLEAN DEFAULT FALSE,
    
    -- Layout (JSON array of widget configs)
    -- Structure: [{ widget_id, widget_type, position: {x,y,w,h}, config: {} }]
    layout JSONB NOT NULL DEFAULT '[]',
    
    -- Sharing
    visibility VARCHAR(50) DEFAULT 'private' CHECK (visibility IN ('private', 'workspace', 'public_link')),
    share_token VARCHAR(100) UNIQUE,
    is_enforced BOOLEAN DEFAULT FALSE,
    enforced_for_roles TEXT[],
    
    -- Metadata
    created_via VARCHAR(50) DEFAULT 'agent' CHECK (created_via IN ('agent', 'manual', 'template')),
    
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Indexes for user_dashboards
CREATE INDEX IF NOT EXISTS idx_dashboards_user ON user_dashboards(user_id);
CREATE INDEX IF NOT EXISTS idx_dashboards_workspace ON user_dashboards(workspace_id);
CREATE INDEX IF NOT EXISTS idx_dashboards_share_token ON user_dashboards(share_token) WHERE share_token IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_dashboards_default ON user_dashboards(user_id, is_default) WHERE is_default = TRUE;

-- ============================================================================
-- DASHBOARD WIDGETS TABLE (Widget Type Registry)
-- Defines available widget types and their configuration schemas
-- ============================================================================
CREATE TABLE IF NOT EXISTS dashboard_widgets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    widget_type VARCHAR(100) NOT NULL UNIQUE,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    category VARCHAR(50),
    
    -- Configuration
    config_schema JSONB NOT NULL,           -- JSON Schema for widget config validation
    default_config JSONB DEFAULT '{}',      -- Default config values
    default_size JSONB DEFAULT '{"w": 4, "h": 3}',
    min_size JSONB DEFAULT '{"w": 2, "h": 2}',
    
    -- Real-time subscriptions
    sse_events TEXT[],                      -- Events this widget listens to
    
    -- Availability
    is_enabled BOOLEAN DEFAULT TRUE,
    requires_feature VARCHAR(100),          -- Feature flag dependency (e.g., 'team', 'client_portal')
    
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- ============================================================================
-- DASHBOARD TEMPLATES TABLE
-- Predefined dashboard templates users can clone
-- ============================================================================
CREATE TABLE IF NOT EXISTS dashboard_templates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    description TEXT,
    category VARCHAR(50),
    
    -- Layout configuration (same structure as user_dashboards.layout)
    layout JSONB NOT NULL,
    
    -- Display
    thumbnail_url TEXT,
    is_default BOOLEAN DEFAULT FALSE,       -- Show as "recommended" template
    sort_order INTEGER DEFAULT 0,
    
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- ============================================================================
-- SEED DATA: Widget Types
-- ============================================================================
INSERT INTO dashboard_widgets (widget_type, name, description, category, config_schema, default_config, default_size, min_size, sse_events, is_enabled) VALUES

-- Core Widgets (Priority 1)
('task_summary', 'Task Summary', 'Tasks grouped by status (todo/in-progress/done)', 'tasks', 
 '{"type": "object", "properties": {"group_by": {"type": "string", "enum": ["status", "project", "priority"]}, "show_completed": {"type": "boolean"}}}',
 '{"group_by": "status", "show_completed": false}',
 '{"w": 4, "h": 3}', '{"w": 2, "h": 2}',
 ARRAY['task.created', 'task.updated', 'task.deleted'], TRUE),

('task_list', 'Task List', 'Filterable, sortable task list', 'tasks',
 '{"type": "object", "properties": {"filter": {"type": "object"}, "sort": {"type": "string"}, "limit": {"type": "integer"}}}',
 '{"filter": {}, "sort": "due_date", "limit": 10}',
 '{"w": 6, "h": 4}', '{"w": 3, "h": 2}',
 ARRAY['task.created', 'task.updated', 'task.deleted'], TRUE),

('project_progress', 'Project Progress', 'Project completion percentage bars', 'projects',
 '{"type": "object", "properties": {"project_ids": {"type": "array", "items": {"type": "string"}}, "show_all": {"type": "boolean"}}}',
 '{"show_all": true}',
 '{"w": 4, "h": 3}', '{"w": 2, "h": 2}',
 ARRAY['task.completed', 'project.updated'], TRUE),

('upcoming_deadlines', 'Upcoming Deadlines', 'Tasks due in the next X days', 'tasks',
 '{"type": "object", "properties": {"days_ahead": {"type": "integer"}, "show_overdue": {"type": "boolean"}}}',
 '{"days_ahead": 7, "show_overdue": true}',
 '{"w": 4, "h": 4}', '{"w": 2, "h": 2}',
 ARRAY['task.created', 'task.updated', 'task.deleted'], TRUE),

('recent_activity', 'Recent Activity', 'Activity feed of recent actions', 'activity',
 '{"type": "object", "properties": {"limit": {"type": "integer"}, "types": {"type": "array", "items": {"type": "string"}}}}',
 '{"limit": 10, "types": []}',
 '{"w": 4, "h": 4}', '{"w": 2, "h": 3}',
 ARRAY['activity.*'], TRUE),

('metric_card', 'Metric Card', 'Single number KPI display', 'analytics',
 '{"type": "object", "properties": {"metric": {"type": "string", "enum": ["tasks_due_today", "tasks_overdue", "tasks_completed_week", "active_projects"]}, "comparison_period": {"type": "string"}}}',
 '{"metric": "tasks_due_today", "comparison_period": "yesterday"}',
 '{"w": 2, "h": 2}', '{"w": 2, "h": 2}',
 ARRAY['task.*', 'project.*'], TRUE),

-- Extended Widgets (Priority 2)
('task_burndown', 'Task Burndown', 'Chart showing tasks completed over time', 'analytics',
 '{"type": "object", "properties": {"project_id": {"type": "string"}, "days": {"type": "integer"}}}',
 '{"days": 30}',
 '{"w": 6, "h": 4}', '{"w": 4, "h": 3}',
 ARRAY['task.created', 'task.completed'], TRUE),

('workload_heatmap', 'Workload Heatmap', 'Calendar heatmap of task density', 'analytics',
 '{"type": "object", "properties": {"date_range": {"type": "string", "enum": ["month", "quarter", "year"]}}}',
 '{"date_range": "month"}',
 '{"w": 6, "h": 3}', '{"w": 4, "h": 2}',
 ARRAY['task.created', 'task.updated'], TRUE),

('client_overview', 'Client Overview', 'Client cards with project/task counts', 'clients',
 '{"type": "object", "properties": {"limit": {"type": "integer"}, "sort_by": {"type": "string"}}}',
 '{"limit": 5, "sort_by": "recent_activity"}',
 '{"w": 4, "h": 3}', '{"w": 3, "h": 2}',
 ARRAY['client.*', 'project.*'], TRUE),

('notes_pinned', 'Pinned Notes', 'Display pinned notes/artifacts', 'notes',
 '{"type": "object", "properties": {"limit": {"type": "integer"}}}',
 '{"limit": 5}',
 '{"w": 4, "h": 3}', '{"w": 2, "h": 2}',
 ARRAY['artifact.created', 'artifact.updated'], TRUE),

('quick_actions', 'Quick Actions', 'Buttons for common actions', 'utility',
 '{"type": "object", "properties": {"actions": {"type": "array", "items": {"type": "object"}}}}',
 '{"actions": [{"label": "New Task", "action": "create_task"}, {"label": "New Note", "action": "create_note"}]}',
 '{"w": 2, "h": 2}', '{"w": 2, "h": 1}',
 NULL, TRUE),

('agent_shortcuts', 'Agent Shortcuts', 'Quick prompts to OS agent', 'utility',
 '{"type": "object", "properties": {"shortcuts": {"type": "array", "items": {"type": "object"}}}}',
 '{"shortcuts": [{"label": "Summarize my day", "prompt": "What do I have on my plate today?"}, {"label": "What''s overdue?", "prompt": "Show me all overdue tasks"}]}',
 '{"w": 3, "h": 2}', '{"w": 2, "h": 2}',
 NULL, TRUE),

-- Future Widgets (Disabled until dependencies exist)
('team_capacity', 'Team Capacity', 'Shows team member availability/bandwidth', 'team',
 '{"type": "object", "properties": {"team_id": {"type": "string"}}}',
 '{}',
 '{"w": 4, "h": 3}', '{"w": 3, "h": 2}',
 ARRAY['team.*'], FALSE),

('team_workload', 'Team Workload', 'Work distribution across team members', 'team',
 '{"type": "object", "properties": {"team_id": {"type": "string"}}}',
 '{}',
 '{"w": 6, "h": 4}', '{"w": 4, "h": 3}',
 ARRAY['team.*', 'task.*'], FALSE),

('project_timeline', 'Project Timeline', 'Gantt-style project timeline view', 'projects',
 '{"type": "object", "properties": {"project_ids": {"type": "array"}}}',
 '{}',
 '{"w": 12, "h": 4}', '{"w": 6, "h": 3}',
 ARRAY['project.*', 'task.*'], FALSE),

('client_pipeline', 'Client Pipeline', 'Sales pipeline visualization', 'clients',
 '{"type": "object", "properties": {"stages": {"type": "array"}}}',
 '{}',
 '{"w": 8, "h": 4}', '{"w": 6, "h": 3}',
 ARRAY['client.*'], FALSE),

('client_activity', 'Client Activity', 'Client interaction history', 'clients',
 '{"type": "object", "properties": {"client_id": {"type": "string"}}}',
 '{}',
 '{"w": 4, "h": 4}', '{"w": 3, "h": 3}',
 ARRAY['client.*'], FALSE),

('task_calendar', 'Task Calendar', 'Calendar view of tasks by due date', 'tasks',
 '{"type": "object", "properties": {"view": {"type": "string", "enum": ["month", "week"]}}}',
 '{"view": "month"}',
 '{"w": 8, "h": 5}', '{"w": 6, "h": 4}',
 ARRAY['task.*'], FALSE),

('custom_query', 'Custom Query', 'Power user SQL/filter widget', 'advanced',
 '{"type": "object", "properties": {"query": {"type": "string"}, "display": {"type": "string"}}}',
 '{}',
 '{"w": 6, "h": 4}', '{"w": 4, "h": 3}',
 NULL, FALSE)

ON CONFLICT (widget_type) DO NOTHING;

-- ============================================================================
-- SEED DATA: Dashboard Templates
-- ============================================================================
INSERT INTO dashboard_templates (name, description, category, layout, is_default, sort_order) VALUES

('My Day', 'Perfect for daily planning and task management', 'personal',
 '[
   {"widget_id": "gen_1", "widget_type": "task_summary", "position": {"x": 0, "y": 0, "w": 4, "h": 3}, "config": {"group_by": "status"}},
   {"widget_id": "gen_2", "widget_type": "upcoming_deadlines", "position": {"x": 4, "y": 0, "w": 4, "h": 3}, "config": {"days_ahead": 3}},
   {"widget_id": "gen_3", "widget_type": "quick_actions", "position": {"x": 8, "y": 0, "w": 4, "h": 2}, "config": {}},
   {"widget_id": "gen_4", "widget_type": "recent_activity", "position": {"x": 8, "y": 2, "w": 4, "h": 3}, "config": {"limit": 5}}
 ]'::jsonb,
 TRUE, 1),

('Project Manager', 'Track project progress and team workload', 'management',
 '[
   {"widget_id": "gen_1", "widget_type": "project_progress", "position": {"x": 0, "y": 0, "w": 6, "h": 3}, "config": {"show_all": true}},
   {"widget_id": "gen_2", "widget_type": "task_burndown", "position": {"x": 6, "y": 0, "w": 6, "h": 3}, "config": {"days": 30}},
   {"widget_id": "gen_3", "widget_type": "upcoming_deadlines", "position": {"x": 0, "y": 3, "w": 6, "h": 4}, "config": {"days_ahead": 14}},
   {"widget_id": "gen_4", "widget_type": "workload_heatmap", "position": {"x": 6, "y": 3, "w": 6, "h": 3}, "config": {"date_range": "month"}}
 ]'::jsonb,
 FALSE, 2),

('Executive', 'High-level KPIs and project overview', 'leadership',
 '[
   {"widget_id": "gen_1", "widget_type": "metric_card", "position": {"x": 0, "y": 0, "w": 3, "h": 2}, "config": {"metric": "tasks_due_today"}},
   {"widget_id": "gen_2", "widget_type": "metric_card", "position": {"x": 3, "y": 0, "w": 3, "h": 2}, "config": {"metric": "tasks_overdue"}},
   {"widget_id": "gen_3", "widget_type": "metric_card", "position": {"x": 6, "y": 0, "w": 3, "h": 2}, "config": {"metric": "tasks_completed_week"}},
   {"widget_id": "gen_4", "widget_type": "metric_card", "position": {"x": 9, "y": 0, "w": 3, "h": 2}, "config": {"metric": "active_projects"}},
   {"widget_id": "gen_5", "widget_type": "project_progress", "position": {"x": 0, "y": 2, "w": 6, "h": 3}, "config": {"show_all": true}},
   {"widget_id": "gen_6", "widget_type": "task_burndown", "position": {"x": 6, "y": 2, "w": 6, "h": 3}, "config": {"days": 30}}
 ]'::jsonb,
 FALSE, 3),

('Developer', 'Focus on your assigned tasks and notes', 'personal',
 '[
   {"widget_id": "gen_1", "widget_type": "task_list", "position": {"x": 0, "y": 0, "w": 6, "h": 4}, "config": {"filter": {"assigned_to_me": true}, "sort": "priority", "limit": 15}},
   {"widget_id": "gen_2", "widget_type": "notes_pinned", "position": {"x": 6, "y": 0, "w": 6, "h": 3}, "config": {"limit": 5}},
   {"widget_id": "gen_3", "widget_type": "recent_activity", "position": {"x": 6, "y": 3, "w": 6, "h": 3}, "config": {"limit": 8}},
   {"widget_id": "gen_4", "widget_type": "agent_shortcuts", "position": {"x": 0, "y": 4, "w": 4, "h": 2}, "config": {}}
 ]'::jsonb,
 FALSE, 4),

('Client Focus', 'Manage client relationships and projects', 'business',
 '[
   {"widget_id": "gen_1", "widget_type": "client_overview", "position": {"x": 0, "y": 0, "w": 4, "h": 3}, "config": {"limit": 5}},
   {"widget_id": "gen_2", "widget_type": "project_progress", "position": {"x": 4, "y": 0, "w": 4, "h": 3}, "config": {"show_all": true}},
   {"widget_id": "gen_3", "widget_type": "upcoming_deadlines", "position": {"x": 8, "y": 0, "w": 4, "h": 3}, "config": {"days_ahead": 14}},
   {"widget_id": "gen_4", "widget_type": "recent_activity", "position": {"x": 0, "y": 3, "w": 6, "h": 3}, "config": {"limit": 10}}
 ]'::jsonb,
 FALSE, 5)

ON CONFLICT DO NOTHING;

-- ============================================================================
-- TRIGGER: Update updated_at timestamp
-- ============================================================================
CREATE OR REPLACE FUNCTION update_dashboard_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trigger_update_dashboard_timestamp ON user_dashboards;
CREATE TRIGGER trigger_update_dashboard_timestamp
    BEFORE UPDATE ON user_dashboards
    FOR EACH ROW
    EXECUTE FUNCTION update_dashboard_updated_at();
-- Push devices for mobile notifications
CREATE TABLE IF NOT EXISTS push_devices (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id TEXT NOT NULL REFERENCES "user"(id) ON DELETE CASCADE,
    device_id TEXT NOT NULL,
    platform TEXT NOT NULL CHECK (platform IN ('ios', 'android', 'web')),
    push_token TEXT NOT NULL,
    app_version TEXT,
    os_version TEXT,
    device_model TEXT,
    is_active BOOLEAN DEFAULT true,
    last_used_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(user_id, device_id)
);

CREATE INDEX IF NOT EXISTS idx_push_devices_user ON push_devices(user_id);
CREATE INDEX IF NOT EXISTS idx_push_devices_token ON push_devices(push_token);
CREATE INDEX IF NOT EXISTS idx_push_devices_active ON push_devices(user_id, is_active) WHERE is_active = true;
-- Web Push subscriptions for browser notifications (Web Push API)
-- This is separate from push_devices which is for mobile (FCM/APNs)
CREATE TABLE IF NOT EXISTS push_subscriptions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id TEXT NOT NULL REFERENCES "user"(id) ON DELETE CASCADE,
    endpoint TEXT NOT NULL UNIQUE,  -- Web Push endpoint URL
    p256dh TEXT NOT NULL,           -- Public key for encryption
    auth TEXT NOT NULL,             -- Auth secret
    user_agent TEXT,                -- Browser/device info
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_push_subscriptions_user ON push_subscriptions(user_id);
CREATE INDEX IF NOT EXISTS idx_push_subscriptions_endpoint ON push_subscriptions(endpoint);
-- ============================================================================
-- Migration: 020_workspaces.sql
-- Description: Workspace & Team Collaboration System
-- Created: 2026-01-07
-- Author: Javaris
-- ============================================================================

-- ============================================================================
-- WORKSPACES TABLE
-- Multi-tenant containers for team collaboration
-- ============================================================================
CREATE TABLE IF NOT EXISTS workspaces (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    -- Workspace Identity
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(100) NOT NULL UNIQUE,
    description TEXT,
    logo_url VARCHAR(500),

    -- Billing & Limits
    plan_type VARCHAR(50) DEFAULT 'free' CHECK (plan_type IN ('free', 'starter', 'professional', 'enterprise')),
    max_members INTEGER DEFAULT 5,
    max_projects INTEGER DEFAULT 10,
    max_storage_gb INTEGER DEFAULT 5,

    -- Settings
    settings JSONB DEFAULT '{}',

    -- Owner
    owner_id VARCHAR(255) NOT NULL,

    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_workspaces_owner ON workspaces(owner_id);
CREATE INDEX IF NOT EXISTS idx_workspaces_slug ON workspaces(slug);

-- ============================================================================
-- WORKSPACE ROLES TABLE
-- Customizable roles per workspace with permissions matrix
-- ============================================================================
CREATE TABLE IF NOT EXISTS workspace_roles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,

    -- Role Identity
    name VARCHAR(100) NOT NULL,
    display_name VARCHAR(100) NOT NULL,
    description TEXT,
    color VARCHAR(20),
    icon VARCHAR(50),

    -- Permissions Matrix (JSONB)
    permissions JSONB NOT NULL DEFAULT '{}',

    -- Role Type
    is_system BOOLEAN DEFAULT FALSE,
    is_default BOOLEAN DEFAULT FALSE,

    -- Hierarchy (higher = more permissions)
    hierarchy_level INTEGER DEFAULT 0 CHECK (hierarchy_level >= 0 AND hierarchy_level <= 100),

    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),

    UNIQUE(workspace_id, name)
);

CREATE INDEX IF NOT EXISTS idx_workspace_roles_workspace ON workspace_roles(workspace_id);

-- ============================================================================
-- WORKSPACE MEMBERS TABLE
-- Links users to workspaces with role assignments
-- ============================================================================
CREATE TABLE IF NOT EXISTS workspace_members (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    user_id VARCHAR(255) NOT NULL,

    -- Role Assignment
    role_id UUID REFERENCES workspace_roles(id) ON DELETE SET NULL,
    role_name VARCHAR(100),

    -- Status
    status VARCHAR(50) DEFAULT 'active' CHECK (status IN ('active', 'invited', 'suspended')),
    invited_by VARCHAR(255),
    invited_at TIMESTAMPTZ,
    joined_at TIMESTAMPTZ,

    -- Permissions Override
    custom_permissions JSONB,

    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),

    UNIQUE(workspace_id, user_id)
);

CREATE INDEX IF NOT EXISTS idx_workspace_members_workspace ON workspace_members(workspace_id);
CREATE INDEX IF NOT EXISTS idx_workspace_members_user ON workspace_members(user_id);
CREATE INDEX IF NOT EXISTS idx_workspace_members_status ON workspace_members(status);

-- ============================================================================
-- WORKSPACE INVITATIONS TABLE
-- Magic link invitations for workspace membership
-- ============================================================================
CREATE TABLE IF NOT EXISTS workspace_invitations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,

    -- Invitation Details
    email VARCHAR(255) NOT NULL,
    token VARCHAR(64) NOT NULL UNIQUE,

    -- Role to assign on accept
    role_id UUID REFERENCES workspace_roles(id) ON DELETE SET NULL,
    role_name VARCHAR(100) NOT NULL,

    -- Inviter Information
    invited_by_id VARCHAR(255) NOT NULL,
    invited_by_name VARCHAR(255),

    -- Status
    status VARCHAR(50) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'accepted', 'expired', 'revoked')),

    -- Timestamps
    expires_at TIMESTAMPTZ NOT NULL,
    accepted_at TIMESTAMPTZ,
    accepted_by_user_id VARCHAR(255),

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_workspace_invitations_token ON workspace_invitations(token);
CREATE INDEX IF NOT EXISTS idx_workspace_invitations_workspace ON workspace_invitations(workspace_id);
CREATE INDEX IF NOT EXISTS idx_workspace_invitations_email ON workspace_invitations(email);
CREATE INDEX IF NOT EXISTS idx_workspace_invitations_status ON workspace_invitations(status) WHERE status = 'pending';

-- Partial unique index: only one pending invitation per email per workspace
CREATE UNIQUE INDEX IF NOT EXISTS idx_workspace_invitations_pending_unique 
ON workspace_invitations(workspace_id, email) 
WHERE status = 'pending';

-- ============================================================================
-- WORKSPACE MEMORIES TABLE
-- Shared knowledge/context across team
-- ============================================================================
CREATE TABLE IF NOT EXISTS workspace_memories (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,

    -- Memory Identity
    title VARCHAR(255) NOT NULL,
    summary TEXT NOT NULL,
    content TEXT NOT NULL,

    -- Memory Type
    memory_type VARCHAR(50) NOT NULL CHECK (memory_type IN ('decision', 'process', 'knowledge', 'pattern', 'policy')),
    category VARCHAR(100),

    -- Scope
    scope_type VARCHAR(50) DEFAULT 'workspace' CHECK (scope_type IN ('workspace', 'project', 'node')),
    scope_id UUID,

    -- Access Control
    visibility VARCHAR(50) DEFAULT 'team' CHECK (visibility IN ('team', 'managers', 'admins')),
    created_by VARCHAR(255) NOT NULL,

    -- Relevance
    importance_score DECIMAL(3,2) DEFAULT 0.5 CHECK (importance_score >= 0 AND importance_score <= 1),
    access_count INTEGER DEFAULT 0,

    -- Embeddings (nullable for Phase 2)
    -- embedding vector(1536),

    -- Metadata
    tags TEXT[] DEFAULT '{}',
    metadata JSONB DEFAULT '{}',

    is_active BOOLEAN DEFAULT TRUE,
    is_pinned BOOLEAN DEFAULT FALSE,

    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_workspace_memories_workspace ON workspace_memories(workspace_id);
CREATE INDEX IF NOT EXISTS idx_workspace_memories_scope ON workspace_memories(scope_type, scope_id);
CREATE INDEX IF NOT EXISTS idx_workspace_memories_type ON workspace_memories(memory_type);
CREATE INDEX IF NOT EXISTS idx_workspace_memories_active ON workspace_memories(workspace_id, is_active) WHERE is_active = TRUE;

-- ============================================================================
-- USER WORKSPACE PROFILES TABLE
-- Per-workspace user profiles (can differ between workspaces)
-- ============================================================================
CREATE TABLE IF NOT EXISTS user_workspace_profiles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    user_id VARCHAR(255) NOT NULL,

    -- Profile Info
    display_name VARCHAR(255),
    title VARCHAR(100),
    department VARCHAR(100),
    avatar_url VARCHAR(500),

    -- Contact
    work_email VARCHAR(255),
    phone VARCHAR(50),

    -- Work Preferences
    timezone VARCHAR(50),
    working_hours JSONB,
    notification_preferences JSONB,

    -- Agent Personalization
    preferred_output_style VARCHAR(50),
    communication_preferences JSONB,
    expertise_areas TEXT[] DEFAULT '{}',

    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),

    UNIQUE(workspace_id, user_id)
);

CREATE INDEX IF NOT EXISTS idx_user_workspace_profiles_workspace ON user_workspace_profiles(workspace_id);
CREATE INDEX IF NOT EXISTS idx_user_workspace_profiles_user ON user_workspace_profiles(user_id);

-- ============================================================================
-- PROJECT MEMBERS TABLE (WORKSPACE-AWARE)
-- Team assignment to projects with workspace context
-- ============================================================================
CREATE TABLE IF NOT EXISTS workspace_project_members (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    user_id VARCHAR(255) NOT NULL,
    workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,

    -- Project-specific role
    project_role VARCHAR(50) DEFAULT 'contributor' CHECK (project_role IN ('lead', 'contributor', 'reviewer', 'viewer')),

    -- Assignment details
    assigned_by VARCHAR(255),
    assigned_at TIMESTAMPTZ DEFAULT NOW(),

    -- Notification preferences
    notification_level VARCHAR(50) DEFAULT 'all' CHECK (notification_level IN ('all', 'mentions', 'none')),

    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),

    UNIQUE(project_id, user_id)
);

CREATE INDEX IF NOT EXISTS idx_workspace_project_members_project ON workspace_project_members(project_id);
CREATE INDEX IF NOT EXISTS idx_workspace_project_members_user ON workspace_project_members(user_id);
CREATE INDEX IF NOT EXISTS idx_workspace_project_members_workspace ON workspace_project_members(workspace_id);

-- ============================================================================
-- FUNCTION: Seed default roles for a new workspace
-- ============================================================================
CREATE OR REPLACE FUNCTION seed_workspace_default_roles(p_workspace_id UUID)
RETURNS void AS $$
BEGIN
    -- Owner (level 100)
    INSERT INTO workspace_roles (workspace_id, name, display_name, description, color, icon, hierarchy_level, is_system, permissions)
    VALUES (p_workspace_id, 'owner', 'Owner', 'Full access to everything including billing and ownership transfer', '#ef4444', 'crown', 100, TRUE, 
    '{
        "projects": {"create": true, "read": true, "update": true, "delete": true, "manage_members": true},
        "tasks": {"create": true, "read": true, "update": true, "delete": true, "assign": true},
        "clients": {"create": true, "read": true, "update": true, "delete": true},
        "artifacts": {"create": true, "read": true, "update": true, "delete": true},
        "settings": {"view": true, "manage": true},
        "team": {"view": true, "invite": true, "manage_roles": true},
        "billing": {"view": true, "manage": true},
        "agent": {"use_all_agents": true, "create_custom_agents": true, "access_workspace_memory": true, "modify_workspace_memory": true}
    }'::jsonb);

    -- Admin (level 90)
    INSERT INTO workspace_roles (workspace_id, name, display_name, description, color, icon, hierarchy_level, is_system, permissions)
    VALUES (p_workspace_id, 'admin', 'Administrator', 'Full access except billing and ownership transfer', '#f97316', 'shield', 90, TRUE,
    '{
        "projects": {"create": true, "read": true, "update": true, "delete": true, "manage_members": true},
        "tasks": {"create": true, "read": true, "update": true, "delete": true, "assign": true},
        "clients": {"create": true, "read": true, "update": true, "delete": true},
        "artifacts": {"create": true, "read": true, "update": true, "delete": true},
        "settings": {"view": true, "manage": true},
        "team": {"view": true, "invite": true, "manage_roles": true},
        "billing": {"view": true, "manage": false},
        "agent": {"use_all_agents": true, "create_custom_agents": true, "access_workspace_memory": true, "modify_workspace_memory": true}
    }'::jsonb);

    -- Manager (level 70)
    INSERT INTO workspace_roles (workspace_id, name, display_name, description, color, icon, hierarchy_level, is_system, permissions)
    VALUES (p_workspace_id, 'manager', 'Manager', 'Can manage projects, tasks, and team members', '#eab308', 'users', 70, TRUE,
    '{
        "projects": {"create": true, "read": true, "update": true, "delete": false, "manage_members": true},
        "tasks": {"create": true, "read": true, "update": true, "delete": true, "assign": true},
        "clients": {"create": true, "read": true, "update": true, "delete": false},
        "artifacts": {"create": true, "read": true, "update": true, "delete": false},
        "settings": {"view": true, "manage": false},
        "team": {"view": true, "invite": true, "manage_roles": false},
        "billing": {"view": false, "manage": false},
        "agent": {"use_all_agents": true, "create_custom_agents": false, "access_workspace_memory": true, "modify_workspace_memory": true}
    }'::jsonb);

    -- Member (level 50) - Default role
    INSERT INTO workspace_roles (workspace_id, name, display_name, description, color, icon, hierarchy_level, is_system, is_default, permissions)
    VALUES (p_workspace_id, 'member', 'Member', 'Standard access - create and edit own work', '#22c55e', 'user', 50, TRUE, TRUE,
    '{
        "projects": {"create": true, "read": true, "update": "own", "delete": false, "manage_members": false},
        "tasks": {"create": true, "read": true, "update": "own", "delete": "own", "assign": false},
        "clients": {"create": false, "read": true, "update": false, "delete": false},
        "artifacts": {"create": true, "read": true, "update": "own", "delete": "own"},
        "settings": {"view": false, "manage": false},
        "team": {"view": true, "invite": false, "manage_roles": false},
        "billing": {"view": false, "manage": false},
        "agent": {"use_all_agents": true, "create_custom_agents": false, "access_workspace_memory": true, "modify_workspace_memory": false}
    }'::jsonb);

    -- Viewer (level 20)
    INSERT INTO workspace_roles (workspace_id, name, display_name, description, color, icon, hierarchy_level, is_system, permissions)
    VALUES (p_workspace_id, 'viewer', 'Viewer', 'Read-only access', '#3b82f6', 'eye', 20, TRUE,
    '{
        "projects": {"create": false, "read": true, "update": false, "delete": false, "manage_members": false},
        "tasks": {"create": false, "read": true, "update": false, "delete": false, "assign": false},
        "clients": {"create": false, "read": true, "update": false, "delete": false},
        "artifacts": {"create": false, "read": true, "update": false, "delete": false},
        "settings": {"view": false, "manage": false},
        "team": {"view": true, "invite": false, "manage_roles": false},
        "billing": {"view": false, "manage": false},
        "agent": {"use_all_agents": false, "create_custom_agents": false, "access_workspace_memory": true, "modify_workspace_memory": false}
    }'::jsonb);

    -- Guest (level 10)
    INSERT INTO workspace_roles (workspace_id, name, display_name, description, color, icon, hierarchy_level, is_system, permissions)
    VALUES (p_workspace_id, 'guest', 'Guest', 'Limited read access to specific projects', '#6b7280', 'user-x', 10, TRUE,
    '{
        "projects": {"create": false, "read": "assigned", "update": false, "delete": false, "manage_members": false},
        "tasks": {"create": false, "read": "assigned", "update": false, "delete": false, "assign": false},
        "clients": {"create": false, "read": false, "update": false, "delete": false},
        "artifacts": {"create": false, "read": "assigned", "update": false, "delete": false},
        "settings": {"view": false, "manage": false},
        "team": {"view": false, "invite": false, "manage_roles": false},
        "billing": {"view": false, "manage": false},
        "agent": {"use_all_agents": false, "create_custom_agents": false, "access_workspace_memory": false, "modify_workspace_memory": false}
    }'::jsonb);
END;
$$ LANGUAGE plpgsql;

-- ============================================================================
-- TRIGGER: Auto-update updated_at timestamps
-- ============================================================================
CREATE OR REPLACE FUNCTION update_workspace_timestamp()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Apply triggers
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_trigger WHERE tgname = 'trigger_workspaces_updated_at') THEN
        DROP TRIGGER IF EXISTS trigger_workspaces_updated_at ON workspaces;
CREATE TRIGGER trigger_workspaces_updated_at
    BEFORE UPDATE ON workspaces
            FOR EACH ROW EXECUTE FUNCTION update_workspace_timestamp();
    END IF;

    IF NOT EXISTS (SELECT 1 FROM pg_trigger WHERE tgname = 'trigger_workspace_roles_updated_at') THEN
        DROP TRIGGER IF EXISTS trigger_workspace_roles_updated_at ON workspace_roles;
CREATE TRIGGER trigger_workspace_roles_updated_at
    BEFORE UPDATE ON workspace_roles
            FOR EACH ROW EXECUTE FUNCTION update_workspace_timestamp();
    END IF;

    IF NOT EXISTS (SELECT 1 FROM pg_trigger WHERE tgname = 'trigger_workspace_members_updated_at') THEN
        DROP TRIGGER IF EXISTS trigger_workspace_members_updated_at ON workspace_members;
CREATE TRIGGER trigger_workspace_members_updated_at
    BEFORE UPDATE ON workspace_members
            FOR EACH ROW EXECUTE FUNCTION update_workspace_timestamp();
    END IF;

    IF NOT EXISTS (SELECT 1 FROM pg_trigger WHERE tgname = 'trigger_workspace_invitations_updated_at') THEN
        DROP TRIGGER IF EXISTS trigger_workspace_invitations_updated_at ON workspace_invitations;
CREATE TRIGGER trigger_workspace_invitations_updated_at
    BEFORE UPDATE ON workspace_invitations
            FOR EACH ROW EXECUTE FUNCTION update_workspace_timestamp();
    END IF;

    IF NOT EXISTS (SELECT 1 FROM pg_trigger WHERE tgname = 'trigger_workspace_memories_updated_at') THEN
        DROP TRIGGER IF EXISTS trigger_workspace_memories_updated_at ON workspace_memories;
CREATE TRIGGER trigger_workspace_memories_updated_at
    BEFORE UPDATE ON workspace_memories
            FOR EACH ROW EXECUTE FUNCTION update_workspace_timestamp();
    END IF;

    IF NOT EXISTS (SELECT 1 FROM pg_trigger WHERE tgname = 'trigger_user_workspace_profiles_updated_at') THEN
        DROP TRIGGER IF EXISTS trigger_user_workspace_profiles_updated_at ON user_workspace_profiles;
CREATE TRIGGER trigger_user_workspace_profiles_updated_at
    BEFORE UPDATE ON user_workspace_profiles
            FOR EACH ROW EXECUTE FUNCTION update_workspace_timestamp();
    END IF;
END $$;

-- ============================================================================
-- ROLLBACK SCRIPT (for reference)
-- ============================================================================
-- DROP TRIGGER IF EXISTS trigger_workspaces_updated_at ON workspaces;
-- DROP TRIGGER IF EXISTS trigger_workspace_roles_updated_at ON workspace_roles;
-- DROP TRIGGER IF EXISTS trigger_workspace_members_updated_at ON workspace_members;
-- DROP TRIGGER IF EXISTS trigger_workspace_invitations_updated_at ON workspace_invitations;
-- DROP TRIGGER IF EXISTS trigger_workspace_memories_updated_at ON workspace_memories;
-- DROP TRIGGER IF EXISTS trigger_user_workspace_profiles_updated_at ON user_workspace_profiles;
-- DROP FUNCTION IF EXISTS update_workspace_timestamp();
-- DROP FUNCTION IF EXISTS seed_workspace_default_roles(UUID);
-- DROP TABLE IF EXISTS workspace_project_members;
-- DROP TABLE IF EXISTS user_workspace_profiles;
-- DROP TABLE IF EXISTS workspace_memories;
-- DROP TABLE IF EXISTS workspace_invitations;
-- DROP TABLE IF EXISTS workspace_members;
-- DROP TABLE IF EXISTS workspace_roles;
-- DROP TABLE IF EXISTS workspaces;

