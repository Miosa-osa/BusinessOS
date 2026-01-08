# Backend Development Tasks for Pedro

**Created:** December 28, 2025
**Priority:** HIGH
**Context:** Building on recent artifact versioning, KB sync, and model routing improvements

---

## Overview

We need to extend the AI system with three major feature sets:

1. **Chain of Thought (COT) / Thinking System** - Visible reasoning with enable/disable
2. **Agents & Commands Customization** - User-configurable agents in AI Settings
3. **Focus Mode Enhancement** - Real web search, proper context loading

These features need backend API support and will be consumed by the frontend.

---

## PART 1: Chain of Thought (COT) Thinking System

### 1.1 Requirements

Users need to:
- See the AI's thinking/reasoning process in real-time
- Enable/disable thinking mode per conversation or globally
- Track thinking steps and reasoning chains
- Create custom reasoning templates/systems

### 1.2 Database Schema

```sql
-- Thinking/reasoning tracking
CREATE TABLE thinking_traces (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    conversation_id UUID REFERENCES conversations(id) ON DELETE CASCADE,
    message_id UUID REFERENCES messages(id) ON DELETE CASCADE,

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

CREATE INDEX idx_thinking_traces_conversation ON thinking_traces(conversation_id);
CREATE INDEX idx_thinking_traces_message ON thinking_traces(message_id);

-- Custom reasoning templates/systems
CREATE TABLE reasoning_templates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,

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
```

### 1.3 API Endpoints Needed

```
POST   /api/chat/message
       - Add: thinking_enabled (bool), reasoning_template_id (optional)
       - Response: Include thinking chunks in SSE stream

GET    /api/thinking/traces/:conversation_id
       - Get all thinking traces for a conversation

GET    /api/thinking/trace/:message_id
       - Get thinking trace for a specific message

POST   /api/reasoning/templates
       - Create custom reasoning template

GET    /api/reasoning/templates
       - List user's reasoning templates

PUT    /api/reasoning/templates/:id
       - Update reasoning template

DELETE /api/reasoning/templates/:id
       - Delete reasoning template

PUT    /api/settings
       - Add: thinking_settings object
         {
           enabled: boolean,
           show_in_ui: boolean,
           save_traces: boolean,
           default_template_id: string | null,
           max_thinking_tokens: number
         }
```

### 1.4 SSE Stream Format for Thinking

The chat stream needs to include thinking content:

```json
// New event types for SSE stream
{ "type": "thinking_start", "data": { "step": 1 } }
{ "type": "thinking_chunk", "data": { "content": "Let me analyze...", "step": 1 } }
{ "type": "thinking_end", "data": { "step": 1, "duration_ms": 1200 } }
{ "type": "content_start", "data": {} }
{ "type": "content_chunk", "data": { "content": "Based on my analysis..." } }
{ "type": "content_end", "data": {} }
```

### 1.5 Integration with LLM Services

Update `services/llm.go` and provider services to:
- Accept `thinking_enabled` parameter
- For models that support thinking (Claude with extended thinking, DeepSeek-R1):
  - Enable native thinking mode
- For models without native support:
  - Inject thinking instruction into system prompt
  - Parse thinking from response (e.g., `<thinking>...</thinking>` tags)

---

## PART 2: Agents & Commands Customization

### 2.1 Requirements

Users need to:
- View all available agents in AI Settings
- Create custom agents with specific capabilities
- Configure agent parameters (model, temperature, system prompt)
- Enable/disable agents
- Create and customize slash commands
- Map commands to agents or workflows

### 2.2 Database Schema

```sql
-- Custom agents
CREATE TABLE custom_agents (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,

    -- Identity
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(100) NOT NULL,            -- For @agent-name invocation
    description TEXT,
    icon VARCHAR(50),                       -- Lucide icon name
    color VARCHAR(20),                      -- Accent color

    -- Configuration
    model VARCHAR(100),                     -- Preferred model
    temperature FLOAT DEFAULT 0.7,
    max_tokens INT DEFAULT 8192,
    system_prompt TEXT,                     -- Agent's base personality/instructions

    -- Capabilities
    capabilities JSONB DEFAULT '[]',        -- ['web_search', 'code_execution', 'file_access']
    tools_enabled JSONB DEFAULT '[]',       -- Which MCP tools can this agent use

    -- Behavior
    thinking_enabled BOOLEAN DEFAULT false,
    auto_delegate BOOLEAN DEFAULT false,    -- Can delegate to other agents

    -- Status
    is_enabled BOOLEAN DEFAULT true,
    is_system BOOLEAN DEFAULT false,        -- System agents can't be deleted

    -- Usage tracking
    times_used INT DEFAULT 0,
    last_used_at TIMESTAMPTZ,

    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),

    UNIQUE(user_id, slug)
);

-- Custom slash commands
CREATE TABLE custom_commands (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,

    -- Identity
    name VARCHAR(100) NOT NULL,             -- Display name
    command VARCHAR(50) NOT NULL,           -- The /command trigger
    description TEXT,

    -- Execution
    agent_id UUID REFERENCES custom_agents(id),  -- Which agent handles this
    action_type VARCHAR(50) NOT NULL,       -- 'agent', 'workflow', 'prompt_template'

    -- For prompt templates
    prompt_template TEXT,                   -- Template with {{input}} placeholder

    -- For workflows
    workflow_steps JSONB,                   -- Ordered steps to execute

    -- Options
    requires_input BOOLEAN DEFAULT true,    -- Does command need user input?
    show_in_menu BOOLEAN DEFAULT true,      -- Show in command palette

    -- Usage
    times_used INT DEFAULT 0,

    is_enabled BOOLEAN DEFAULT true,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),

    UNIQUE(user_id, command)
);

-- Agent delegation rules
CREATE TABLE agent_delegations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    from_agent_id UUID NOT NULL REFERENCES custom_agents(id) ON DELETE CASCADE,
    to_agent_id UUID NOT NULL REFERENCES custom_agents(id) ON DELETE CASCADE,

    -- When to delegate
    trigger_pattern TEXT,                   -- Regex or keyword pattern
    trigger_capability VARCHAR(100),        -- When specific capability needed

    -- How to delegate
    delegation_prompt TEXT,                 -- Instructions for handoff

    created_at TIMESTAMPTZ DEFAULT NOW()
);
```

### 2.3 API Endpoints Needed

```
# Agents
GET    /api/agents
       - List all agents (system + custom)
       - Include usage stats

POST   /api/agents
       - Create custom agent

GET    /api/agents/:id
       - Get agent details

PUT    /api/agents/:id
       - Update agent

DELETE /api/agents/:id
       - Delete custom agent (not system agents)

POST   /api/agents/:id/test
       - Test agent with a sample prompt

# Commands
GET    /api/commands
       - List all commands (system + custom)

POST   /api/commands
       - Create custom command

PUT    /api/commands/:id
       - Update command

DELETE /api/commands/:id
       - Delete command

# In chat
POST   /api/chat/message
       - Add: agent_id (optional) - explicitly use specific agent
       - Add: command (optional) - trigger command execution
```

### 2.4 System Agents to Pre-populate

Create these as system agents (is_system = true):

| Slug | Name | Description | Capabilities |
|------|------|-------------|--------------|
| researcher | Research Agent | Deep web research and analysis | web_search, summarize |
| writer | Content Writer | Long-form content creation | document_generation |
| coder | Code Assistant | Code generation and review | code_execution, code_review |
| analyst | Data Analyst | Data analysis and insights | data_processing, charts |
| planner | Task Planner | Break down complex tasks | task_decomposition |

### 2.5 System Commands to Pre-populate

| Command | Action | Agent |
|---------|--------|-------|
| /research | Deep research on topic | researcher |
| /write | Generate document | writer |
| /code | Generate/explain code | coder |
| /analyze | Analyze data or text | analyst |
| /plan | Create action plan | planner |
| /summarize | Summarize content | researcher |
| /brainstorm | Generate ideas | (default) |

---

## PART 3: Focus Mode Enhancement

### 3.1 Requirements

Focus Mode needs to:
- Actually perform web search when research mode selected
- Load proper context based on selected focus options
- Configure how context is injected into the conversation
- Track which focus modes are active and their results

### 3.2 Current Focus Mode Options (from frontend)

```typescript
// Focus cards currently in UI
- quick: Quick Answer (fast, concise)
- deep: Deep Research (thorough, web search)
- create: Create Content (document generation)
- analyze: Analyze Data (data processing)
- code: Code Assistant (coding tasks)
- plan: Strategic Planning (task breakdown)
```

### 3.3 Database Schema Additions

```sql
-- Focus mode configurations
CREATE TABLE focus_configurations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,

    focus_id VARCHAR(50) NOT NULL,          -- 'quick', 'deep', 'create', etc.

    -- Behavior configuration
    web_search_enabled BOOLEAN DEFAULT false,
    web_search_depth INT DEFAULT 3,         -- Number of results to fetch
    web_search_domains TEXT[],              -- Allowed domains (empty = all)

    context_loading JSONB DEFAULT '{}',     -- Which contexts to auto-load
    agent_id UUID REFERENCES custom_agents(id),  -- Which agent to use

    -- Prompting
    system_prompt_addition TEXT,            -- Added to system prompt in this mode

    -- Output preferences
    response_format VARCHAR(50),            -- 'detailed', 'concise', 'structured'
    include_sources BOOLEAN DEFAULT true,

    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),

    UNIQUE(user_id, focus_id)
);

-- Web search results cache
CREATE TABLE web_search_results (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    conversation_id UUID NOT NULL REFERENCES conversations(id) ON DELETE CASCADE,
    message_id UUID REFERENCES messages(id) ON DELETE CASCADE,

    query TEXT NOT NULL,

    -- Results
    results JSONB NOT NULL,                 -- Array of {title, url, snippet, content}
    result_count INT,

    -- Source tracking
    search_provider VARCHAR(50),            -- 'tavily', 'serper', 'brave', etc.

    -- Timing
    searched_at TIMESTAMPTZ DEFAULT NOW(),
    duration_ms INT
);

CREATE INDEX idx_web_search_conversation ON web_search_results(conversation_id);
```

### 3.4 API Endpoints Needed

```
# Focus configurations
GET    /api/focus/configurations
       - Get user's focus mode configurations

PUT    /api/focus/configurations/:focus_id
       - Update focus mode configuration

# Web search
POST   /api/search/web
       - Perform web search
       - Body: { query, depth, domains, conversation_id }
       - Returns: { results: [...], sources: [...] }

GET    /api/search/results/:conversation_id
       - Get cached search results for conversation

# In chat
POST   /api/chat/message
       - Add: focus_mode (string)
       - Add: focus_options (object) - the selected options from focus cards
       - Backend should:
         1. Load focus configuration
         2. If web_search_enabled, perform search first
         3. Inject search results into context
         4. Load any auto-load contexts
         5. Use configured agent
         6. Apply system_prompt_addition
```

### 3.5 Web Search Integration

Need to integrate a web search provider. Options:

1. **Tavily API** (Recommended)
   - Built for AI/LLM use cases
   - Returns clean, structured content
   - Handles scraping automatically

2. **Serper API**
   - Google search results
   - Need separate scraping

3. **Brave Search API**
   - Privacy-focused
   - Good API

**Implementation:**

```go
// services/web_search.go

type WebSearchService interface {
    Search(ctx context.Context, query string, opts SearchOptions) (*SearchResults, error)
}

type SearchOptions struct {
    MaxResults    int
    SearchDepth   string   // 'basic', 'advanced'
    Domains       []string // Limit to specific domains
    ExcludeDomains []string
    TimeRange     string   // 'day', 'week', 'month', 'year'
}

type SearchResult struct {
    Title   string `json:"title"`
    URL     string `json:"url"`
    Snippet string `json:"snippet"`
    Content string `json:"content"`  // Full scraped content
    Score   float64 `json:"score"`
}

type SearchResults struct {
    Query   string         `json:"query"`
    Results []SearchResult `json:"results"`
    Sources []string       `json:"sources"`
}
```

### 3.6 Context Loading Flow

When a chat message is sent with focus_mode:

```
1. Parse focus_mode and focus_options from request

2. Load FocusConfiguration for this focus_mode
   - Get web_search settings
   - Get context_loading settings
   - Get agent_id

3. If web_search_enabled:
   a. Build search query from user message
   b. Call WebSearchService.Search()
   c. Save results to web_search_results table
   d. Format results for context injection

4. Load auto-load contexts:
   - Based on focus_options selections
   - Based on focus_configuration.context_loading

5. Build final system prompt:
   - Base system prompt
   - + focus_configuration.system_prompt_addition
   - + web search results (if any)
   - + loaded contexts

6. Send to LLM with configured agent settings

7. Stream response back with:
   - thinking (if enabled)
   - content
   - sources (if web search was used)
```

---

## PART 4: Settings API Extensions

### 4.1 Extended Settings Object

The `/api/settings` endpoint needs to handle:

```json
{
  // Existing
  "default_model": "string",
  "model_settings": {
    "temperature": 0.7,
    "maxTokens": 8192,
    "contextWindow": 131072,
    "topP": 0.95,
    "streamResponses": true,
    "showUsageInChat": true
  },

  // NEW: Thinking settings
  "thinking_settings": {
    "enabled": true,
    "show_in_ui": true,
    "save_traces": true,
    "default_template_id": "uuid or null",
    "max_thinking_tokens": 4096
  },

  // NEW: Agent settings
  "agent_settings": {
    "default_agent_id": "uuid or null",
    "auto_delegate": false,
    "show_agent_indicator": true
  },

  // NEW: Focus mode settings
  "focus_settings": {
    "default_focus_mode": "quick",
    "remember_last_focus": true,
    "web_search_provider": "tavily"
  }
}
```

---

## PART 5: Implementation Priority

### Phase 1: Core Thinking System (Week 1)
1. Create thinking_traces table
2. Update chat handler to support thinking_enabled
3. Modify SSE stream to include thinking events
4. Update LLM services to handle thinking mode

### Phase 2: Basic Agents & Commands (Week 2)
1. Create custom_agents and custom_commands tables
2. Pre-populate system agents and commands
3. Create CRUD APIs for agents and commands
4. Integrate agent selection into chat handler

### Phase 3: Focus Mode & Web Search (Week 3)
1. Integrate web search provider (Tavily recommended)
2. Create focus_configurations table
3. Update chat handler for focus mode context loading
4. Cache and display search results

### Phase 4: Polish & Integration (Week 4)
1. Reasoning templates
2. Agent delegation
3. Command workflows
4. Settings UI integration

---

## Environment Variables Needed

```env
# Web Search
TAVILY_API_KEY=tvly-xxxxx
# OR
SERPER_API_KEY=xxxxx
# OR
BRAVE_SEARCH_API_KEY=xxxxx

# Optional: For advanced thinking
ANTHROPIC_EXTENDED_THINKING=true
```

---

## Testing Checklist

- [ ] Thinking shows in real-time during stream
- [ ] Thinking can be enabled/disabled per message
- [ ] Thinking traces are saved to database
- [ ] Custom agents can be created/edited/deleted
- [ ] Custom commands work with /slash trigger
- [ ] System agents cannot be deleted
- [ ] Web search returns results in focus mode
- [ ] Search results appear in context
- [ ] Focus mode loads correct contexts
- [ ] Settings persist across sessions

---

## Questions for Clarification

1. **Web Search Provider**: Which provider should we use? (Tavily recommended for AI use cases)
2. **Thinking Models**: Should we prioritize Claude extended thinking or DeepSeek-R1 style?
3. **Agent Delegation**: How complex should the delegation logic be?
4. **Command Workflows**: Do we need multi-step command workflows now or later?

---

## Reference: Recent Changes to Build On

The following was just merged to `main-dev`:

1. **Artifact Versioning** - Version history and restore for artifacts
2. **KB Sync** - Sync artifacts to Knowledge Base contexts
3. **Provider Inference** - Model name → provider routing in `services/llm.go`
4. **Settings Persistence** - Proper loading of model_settings

Use these patterns as reference for the new features.
