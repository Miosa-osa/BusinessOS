# Backend Development Tasks for Pedro - V2: Memory, Context & Intelligence System

**Created:** December 31, 2025
**Priority:** CRITICAL
**Context:** Building on Agent V2 (COT, Custom Agents, Focus Modes, Web Search) - Now implementing the intelligence layer

---

## Executive Summary

This document outlines the comprehensive implementation of BusinessOS's **Memory, Context, and Intelligence System**. This is the core "brain" of the platform that makes it truly intelligent and personalized.

### Roberto's Vision (Direct Quotes)

> "I need to make sure the system has the episodic memory and stores deep context when the users use the system based off of different actions or whatever."

> "We need to have a place in the knowledge base right in a database where we're able to view the memory as well, because it's not just artifacts and stuff that are going to be created or generated or whatever inside of the knowledge base. There'll be memories too, so we need to add that."

> "Our system must contain these memories and know when to pull from them."

> "The system knows when to go look in a knowledge base or find a particular document that's related to what we're doing."

> "The AI agent system in the background is able to have a context system that is managed based off the flow of the information from the hierarchies."

> "We need to ensure that this is actually working through the main orchestrator agent and all the other agents have access to this type of memory depending on what's relevant for them."

> "The system will be technically self-learning. So we need to implement the self-learning aspect into this as well so that it knows and grows about the user."

> "I don't want generic ass outputs like they have to be real ass outputs."

---
Nick
## PART 1: Episodic Memory System

### 1.1 Core Requirements

Roberto's exact requirements:

> "I need to make sure the system has the episodic memory and stores deep context when the users use the system based off of different actions or whatever."

> "We need to have a place in the knowledge base right in a database where we're able to view the memory as well, because it's not just artifacts and stuff that are going to be created or generated or whatever inside of the knowledge base. There'll be memories too, so we need to add that."

> "And our system must contain these memories and know when to pull from them."
Nick / Pedro 
### 1.2 Database Schema

```sql
-- ===== EPISODIC MEMORY SYSTEM =====

-- Core memory entries - stores individual memories
CREATE TABLE memories (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(255) NOT NULL,

    -- Memory Identity
    title VARCHAR(255) NOT NULL,
    summary TEXT NOT NULL,                    -- Brief summary of the memory
    content TEXT NOT NULL,                    -- Full memory content

    -- Memory Type
    memory_type VARCHAR(50) NOT NULL,         -- 'fact', 'preference', 'decision', 'pattern', 'insight', 'interaction', 'learning'
    category VARCHAR(100),                    -- User-defined category

    -- Source Tracking (where did this memory come from?)
    source_type VARCHAR(50) NOT NULL,         -- 'conversation', 'voice_note', 'document', 'task', 'project', 'manual', 'inferred'
    source_id UUID,                           -- Reference to source (conversation_id, voice_note_id, etc.)
    source_context TEXT,                      -- Additional context about the source

    -- Hierarchy Links
    project_id UUID REFERENCES projects(id) ON DELETE SET NULL,
    node_id UUID REFERENCES nodes(id) ON DELETE SET NULL,

    -- Relevance & Retrieval
    importance_score DECIMAL(3,2) DEFAULT 0.5,  -- 0.0 to 1.0
    access_count INTEGER DEFAULT 0,             -- How many times this memory was accessed
    last_accessed_at TIMESTAMPTZ,

    -- Embeddings for Semantic Search
    embedding vector(1536),                   -- OpenAI ada-002 embedding
    embedding_model VARCHAR(100),             -- Which model generated the embedding

    -- Memory Lifecycle
    is_active BOOLEAN DEFAULT TRUE,
    is_pinned BOOLEAN DEFAULT FALSE,          -- User pinned this memory
    expires_at TIMESTAMPTZ,                   -- Optional expiration

    -- Metadata
    tags TEXT[] DEFAULT '{}',
    metadata JSONB DEFAULT '{}',

    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_memories_user ON memories(user_id);
CREATE INDEX idx_memories_type ON memories(memory_type);
CREATE INDEX idx_memories_project ON memories(project_id);
CREATE INDEX idx_memories_node ON memories(node_id);
CREATE INDEX idx_memories_importance ON memories(importance_score DESC);
CREATE INDEX idx_memories_embedding ON memories USING ivfflat (embedding vector_cosine_ops);

-- Memory associations - links memories to other entities
CREATE TABLE memory_associations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    memory_id UUID NOT NULL REFERENCES memories(id) ON DELETE CASCADE,

    -- What is this memory associated with?
    entity_type VARCHAR(50) NOT NULL,         -- 'project', 'node', 'task', 'client', 'artifact', 'context', 'conversation'
    entity_id UUID NOT NULL,

    -- Association strength
    relevance_score DECIMAL(3,2) DEFAULT 0.5,
    association_type VARCHAR(50),             -- 'about', 'created_from', 'related_to', 'derived_from'

    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_memory_assoc_memory ON memory_associations(memory_id);
CREATE INDEX idx_memory_assoc_entity ON memory_associations(entity_type, entity_id);

-- Memory access log - track when and how memories are accessed
CREATE TABLE memory_access_log (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    memory_id UUID NOT NULL REFERENCES memories(id) ON DELETE CASCADE,
    user_id VARCHAR(255) NOT NULL,

    -- Access context
    access_type VARCHAR(50) NOT NULL,         -- 'agent_retrieval', 'user_view', 'search_result', 'auto_inject'
    accessing_agent VARCHAR(100),             -- Which agent accessed this memory
    conversation_id UUID,

    -- What triggered the access?
    trigger_query TEXT,                       -- The query/context that triggered retrieval
    relevance_score DECIMAL(3,2),             -- How relevant was this memory to the query

    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_memory_access_memory ON memory_access_log(memory_id);
CREATE INDEX idx_memory_access_time ON memory_access_log(created_at DESC);

-- User preferences and facts - quick-access important user info
CREATE TABLE user_facts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(255) NOT NULL,

    -- Fact details
    fact_key VARCHAR(255) NOT NULL,           -- e.g., 'preferred_name', 'timezone', 'communication_style'
    fact_value TEXT NOT NULL,
    fact_type VARCHAR(50) NOT NULL,           -- 'preference', 'fact', 'style', 'context'

    -- Source
    source_memory_id UUID REFERENCES memories(id) ON DELETE SET NULL,
    confidence_score DECIMAL(3,2) DEFAULT 1.0,

    -- Metadata
    is_active BOOLEAN DEFAULT TRUE,
    last_confirmed_at TIMESTAMPTZ,

    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),

    UNIQUE(user_id, fact_key)
);

CREATE INDEX idx_user_facts_user ON user_facts(user_id);
CREATE INDEX idx_user_facts_type ON user_facts(fact_type);
```

### 1.3 Memory Types

| Type | Description | Example |
|------|-------------|---------|
| `fact` | Factual information about user/business | "User's company has 50 employees" |
| `preference` | User preferences and styles | "User prefers bullet points for summaries" |
| `decision` | Decisions made during conversations | "Decided to use PostgreSQL for the project" |
| `pattern` | Observed patterns in behavior | "User typically works on tasks in morning" |
| `insight` | AI-generated insights | "Project X has recurring deadline issues" |
| `interaction` | Important conversation moments | "Discussed pricing strategy with client" |
| `learning` | Things the system learned | "User corrected AI on terminology X" |
  Nick / Pedro
### 1.4 API Endpoints

```
# Memory CRUD
GET    /api/memories                    # List user's memories (with filters)
POST   /api/memories                    # Create new memory
GET    /api/memories/:id                # Get specific memory
PUT    /api/memories/:id                # Update memory
DELETE /api/memories/:id                # Delete memory
POST   /api/memories/:id/pin            # Pin/unpin memory

# Memory Search & Retrieval
POST   /api/memories/search             # Semantic search memories
POST   /api/memories/relevant           # Get memories relevant to context
GET    /api/memories/project/:projectId # Get memories for project
GET    /api/memories/node/:nodeId       # Get memories for node

# User Facts
GET    /api/user-facts                  # Get all user facts
PUT    /api/user-facts/:key             # Update a user fact
DELETE /api/user-facts/:key             # Delete a user fact

# Memory Analytics
GET    /api/memories/stats              # Memory statistics
GET    /api/memories/recent             # Recently accessed memories
GET    /api/memories/frequent           # Most accessed memories
```
Pedro
### 1.5 Memory Service

```go
// services/memory.go

type MemoryService interface {
    // CRUD
    CreateMemory(ctx context.Context, userID string, input CreateMemoryInput) (*Memory, error)
    GetMemory(ctx context.Context, userID string, memoryID uuid.UUID) (*Memory, error)
    UpdateMemory(ctx context.Context, userID string, memoryID uuid.UUID, input UpdateMemoryInput) (*Memory, error)
    DeleteMemory(ctx context.Context, userID string, memoryID uuid.UUID) error

    // Search & Retrieval
    SemanticSearch(ctx context.Context, userID string, query string, opts SearchOptions) ([]Memory, error)
    GetRelevantMemories(ctx context.Context, userID string, context MemoryContext) ([]Memory, error)
    GetMemoriesForProject(ctx context.Context, userID string, projectID uuid.UUID) ([]Memory, error)
    GetMemoriesForNode(ctx context.Context, userID string, nodeID uuid.UUID) ([]Memory, error)

    // Auto-extraction
    ExtractMemoriesFromConversation(ctx context.Context, userID string, conversationID uuid.UUID) ([]Memory, error)
    ExtractMemoriesFromVoiceNote(ctx context.Context, userID string, voiceNoteID uuid.UUID) ([]Memory, error)

    // Learning
    RecordMemoryAccess(ctx context.Context, memoryID uuid.UUID, accessType string, context AccessContext) error
    UpdateMemoryImportance(ctx context.Context, memoryID uuid.UUID) error

    // User Facts
    GetUserFacts(ctx context.Context, userID string) ([]UserFact, error)
    SetUserFact(ctx context.Context, userID string, key string, value string, factType string) error
}

type CreateMemoryInput struct {
    Title       string
    Summary     string
    Content     string
    MemoryType  string
    Category    string
    SourceType  string
    SourceID    *uuid.UUID
    ProjectID   *uuid.UUID
    NodeID      *uuid.UUID
    Tags        []string
    Metadata    map[string]interface{}
}

type SearchOptions struct {
    MemoryTypes []string
    ProjectID   *uuid.UUID
    NodeID      *uuid.UUID
    MinScore    float64
    MaxResults  int
    IncludeExpired bool
}

type MemoryContext struct {
    Query         string
    ConversationID *uuid.UUID
    ProjectID     *uuid.UUID
    NodeID        *uuid.UUID
    AgentType     string
    CurrentTopic  string
}
```

---

## PART 2: Intelligent Context System

### 2.1 Core Requirements

Roberto's exact requirements:

> "We need our system to be where the context of everything, the system knows when to go look in a knowledge base or find a particular document that's related to what we're doing or something like that."

> "The AI agent system in the background is able to have a context system that is managed based off the flow of the information from the hierarchies."

> "We need to ensure that this is actually working through the main orchestrator agent and all the other agents have access to this type of memory depending on what's relevant for them."

> "They need to have some sort of tool to I guess do some sort of search like a tree search or something like this. Where I can search through the titles and then based off the title or based off different titles, it can always just select one based off that title, then it'll pull that document, read that document, add that document into its context, blah blah, blah."

> "And then these agents, we need to make sure we track their context windows as well."

### 2.2 Context Hierarchy

**IMPORTANT**: The hierarchy starts at the Node level (top) because that's where the most comprehensive context lives.

```
NODE (Top Level - Most Comprehensive Context)
 └── PROJECTS (Each node can contain multiple projects)
      └── CONTEXT PROFILES (Organized groupings of related documents)
           └── DOCUMENTS (Individual items within a profile)
                ├── PDF files (uploaded)
                ├── Markdown files (uploaded or created)
                ├── Standard Operating Procedures (SOPs)
                ├── Frameworks and templates
                └── Any other document type
           └── ARTIFACTS (Generated content)
           └── MEMORIES (Episodic memories)
      └── CHAT HISTORY (Conversations within project context)
      └── VOICE NOTES (Transcribed and linked to project)
      └── TASKS (Project tasks with their context)
```

**Key Points:**
1. **Node is the TOP** - Contains the most comprehensive context
2. **Projects have individual context profiles** - Each profile groups related documents
3. **Context Profiles contain multiple document types** - Not just text, but PDFs, markdown, SOPs, frameworks
4. **Users can upload files** - The system handles PDFs, markdown, and other document formats
5. **Chat history is context** - Conversation history is just as important as voice notes
nick / pedro
### 2.3 Database Schema

```sql
-- ===== CONTEXT MANAGEMENT SYSTEM =====

-- Context profiles - describes the full context tree for an entity
CREATE TABLE context_profiles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(255) NOT NULL,

    -- What this profile describes
    entity_type VARCHAR(50) NOT NULL,         -- 'project', 'node', 'application'
    entity_id UUID NOT NULL,

    -- Profile content
    name VARCHAR(255) NOT NULL,
    description TEXT,

    -- Tree structure (JSON representation of context tree)
    context_tree JSONB NOT NULL DEFAULT '{}',

    -- Summary for quick loading
    summary TEXT,                             -- AI-generated summary of this context
    key_facts TEXT[],                         -- Important facts extracted

    -- Statistics
    total_contexts INTEGER DEFAULT 0,
    total_memories INTEGER DEFAULT 0,
    total_artifacts INTEGER DEFAULT 0,
    total_tasks INTEGER DEFAULT 0,

    -- Embedding for profile matching
    embedding vector(1536),

    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),

    UNIQUE(user_id, entity_type, entity_id)
);

CREATE INDEX idx_context_profiles_user ON context_profiles(user_id);
CREATE INDEX idx_context_profiles_entity ON context_profiles(entity_type, entity_id);

-- Context loading rules - defines what context to load automatically
CREATE TABLE context_loading_rules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(255) NOT NULL,

    -- Rule definition
    name VARCHAR(255) NOT NULL,
    description TEXT,

    -- When to apply this rule
    trigger_type VARCHAR(50) NOT NULL,        -- 'project_select', 'node_select', 'agent_type', 'keyword', 'always'
    trigger_value VARCHAR(255),               -- The specific trigger (project_id, agent name, keyword)

    -- What to load
    load_memories BOOLEAN DEFAULT TRUE,
    memory_types TEXT[],                      -- Which memory types to load
    memory_limit INTEGER DEFAULT 10,

    load_contexts BOOLEAN DEFAULT TRUE,
    context_categories TEXT[],                -- Which KB categories to load
    context_limit INTEGER DEFAULT 5,

    load_artifacts BOOLEAN DEFAULT FALSE,
    artifact_types TEXT[],
    artifact_limit INTEGER DEFAULT 3,

    load_recent_conversations BOOLEAN DEFAULT TRUE,
    conversation_limit INTEGER DEFAULT 3,

    -- Priority (higher = loaded first)
    priority INTEGER DEFAULT 0,

    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_context_rules_user ON context_loading_rules(user_id);
CREATE INDEX idx_context_rules_trigger ON context_loading_rules(trigger_type, trigger_value);

-- Agent context sessions - tracks context for each agent session
CREATE TABLE agent_context_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(255) NOT NULL,
    conversation_id UUID NOT NULL,

    -- Agent info
    agent_type VARCHAR(100) NOT NULL,
    agent_id UUID,                            -- If using custom agent

    -- Context window tracking
    max_context_tokens INTEGER NOT NULL,      -- Model's max context
    used_context_tokens INTEGER DEFAULT 0,
    available_tokens INTEGER,

    -- What's loaded in context
    loaded_memories UUID[] DEFAULT '{}',
    loaded_contexts UUID[] DEFAULT '{}',
    loaded_artifacts UUID[] DEFAULT '{}',

    -- System prompt components
    base_system_prompt TEXT,
    injected_context TEXT,
    total_system_prompt_tokens INTEGER,

    -- Session state
    project_id UUID,
    node_id UUID,
    focus_mode VARCHAR(50),

    -- Timestamps
    started_at TIMESTAMPTZ DEFAULT NOW(),
    last_activity_at TIMESTAMPTZ DEFAULT NOW(),
    ended_at TIMESTAMPTZ
);

CREATE INDEX idx_agent_sessions_user ON agent_context_sessions(user_id);
CREATE INDEX idx_agent_sessions_conversation ON agent_context_sessions(conversation_id);

-- Context retrieval log - tracks what context was retrieved and used
CREATE TABLE context_retrieval_log (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id UUID REFERENCES agent_context_sessions(id) ON DELETE CASCADE,
    user_id VARCHAR(255) NOT NULL,

    -- What was retrieved
    retrieval_type VARCHAR(50) NOT NULL,      -- 'memory', 'context', 'artifact', 'document'
    item_id UUID NOT NULL,
    item_title VARCHAR(255),

    -- How it was retrieved
    retrieval_method VARCHAR(50) NOT NULL,    -- 'semantic_search', 'tree_search', 'rule_based', 'manual'
    query_used TEXT,
    relevance_score DECIMAL(3,2),

    -- Token tracking
    token_count INTEGER,
    was_truncated BOOLEAN DEFAULT FALSE,

    -- Usage
    was_used_in_response BOOLEAN DEFAULT FALSE,

    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_context_retrieval_session ON context_retrieval_log(session_id);
CREATE INDEX idx_context_retrieval_item ON context_retrieval_log(item_id);
```
Pedro
### 2.4 Tree Search Tool for Agents

Roberto's requirement:

> "They need to have some sort of tool to I guess do some sort of search like a tree search or something like this. Where I can search through the titles and then based off the title or based off different titles, it can always just select one based off that title, then it'll pull that document, read that document, add that document into its context."

```go
// tools/context_tools.go

// TreeSearchTool allows agents to search and retrieve context
type TreeSearchTool struct {
    Name        string
    Description string
}

func NewTreeSearchTool() *TreeSearchTool {
    return &TreeSearchTool{
        Name: "tree_search",
        Description: `Search through the user's knowledge base, memories, and documents.

        Use this tool to:
        1. Search by title/name to find specific documents
        2. Search by content/keywords to find relevant information
        3. Browse the context tree to understand available information

        Parameters:
        - query: Search query (title, keywords, or topic)
        - search_type: 'title' | 'content' | 'semantic' | 'browse'
        - entity_types: Array of types to search ['memories', 'contexts', 'artifacts', 'tasks']
        - project_scope: Limit to specific project (optional)
        - node_scope: Limit to specific node (optional)
        - max_results: Maximum results to return (default: 10)

        Returns:
        - results: Array of matching items with id, title, type, relevance_score
        - tree_path: Where in the hierarchy each result lives
        `,
    }
}

type TreeSearchParams struct {
    Query        string   `json:"query"`
    SearchType   string   `json:"search_type"`    // 'title', 'content', 'semantic', 'browse'
    EntityTypes  []string `json:"entity_types"`   // 'memories', 'contexts', 'artifacts', 'tasks'
    ProjectScope *string  `json:"project_scope"`
    NodeScope    *string  `json:"node_scope"`
    MaxResults   int      `json:"max_results"`
}

type TreeSearchResult struct {
    ID             uuid.UUID `json:"id"`
    Title          string    `json:"title"`
    Type           string    `json:"type"`
    Summary        string    `json:"summary"`
    RelevanceScore float64   `json:"relevance_score"`
    TreePath       []string  `json:"tree_path"`    // e.g., ["Project X", "Phase 1", "Research"]
    TokenEstimate  int       `json:"token_estimate"`
}

// LoadContextTool allows agents to load specific items into their context
type LoadContextTool struct {
    Name        string
    Description string
}

func NewLoadContextTool() *LoadContextTool {
    return &LoadContextTool{
        Name: "load_context",
        Description: `Load a specific document, memory, or artifact into the current context.

        Use this after tree_search to actually load and read the content.

        Parameters:
        - item_id: UUID of the item to load
        - item_type: Type of item ('memory', 'context', 'artifact', 'task')
        - include_related: Also load related items (default: false)

        Returns:
        - content: Full content of the item
        - metadata: Additional information about the item
        - related_items: If include_related=true, list of related items
        - tokens_used: How many tokens this added to context
        `,
    }
}

// BrowseTreeTool allows agents to browse the context hierarchy
type BrowseTreeTool struct {
    Name        string
    Description string
}

func NewBrowseTreeTool() *BrowseTreeTool {
    return &BrowseTreeTool{
        Name: "browse_tree",
        Description: `Browse the user's context tree to understand available information.

        Use this to:
        1. See the overall structure of the user's knowledge
        2. Navigate to specific branches
        3. Understand what information is available before searching

        Parameters:
        - path: Path to browse (e.g., "projects/PROJECT_ID/nodes" or "/" for root)
        - depth: How deep to show (default: 2)

        Returns:
        - tree: Hierarchical structure of items at this path
        - total_items: Count of items at each level
        `,
    }
}
```

### 2.5 Context Service

```go
// services/context.go

type ContextService interface {
    // Context Building
    BuildContextForAgent(ctx context.Context, userID string, input ContextBuildInput) (*AgentContext, error)
    GetContextProfile(ctx context.Context, userID string, entityType string, entityID uuid.UUID) (*ContextProfile, error)
    UpdateContextProfile(ctx context.Context, userID string, entityType string, entityID uuid.UUID) error

    // Tree Operations
    GetContextTree(ctx context.Context, userID string, projectID *uuid.UUID, nodeID *uuid.UUID) (*ContextTree, error)
    SearchTree(ctx context.Context, userID string, params TreeSearchParams) ([]TreeSearchResult, error)
    LoadContextItem(ctx context.Context, userID string, itemID uuid.UUID, itemType string) (*ContextItem, error)

    // Context Window Management
    CreateContextSession(ctx context.Context, input CreateSessionInput) (*AgentContextSession, error)
    UpdateSessionContext(ctx context.Context, sessionID uuid.UUID, items []ContextItem) error
    GetSessionTokenUsage(ctx context.Context, sessionID uuid.UUID) (*TokenUsage, error)

    // Rules
    GetLoadingRules(ctx context.Context, userID string, triggerType string, triggerValue string) ([]ContextLoadingRule, error)
    ApplyLoadingRules(ctx context.Context, userID string, session *AgentContextSession) error
}

type ContextBuildInput struct {
    UserID         string
    ProjectID      *uuid.UUID
    NodeID         *uuid.UUID
    ConversationID *uuid.UUID
    AgentType      string
    FocusMode      string
    CurrentQuery   string
    MaxTokens      int
}

type AgentContext struct {
    SystemPromptAddition string
    LoadedMemories       []Memory
    LoadedContexts       []Context
    LoadedArtifacts      []Artifact
    RecentConversations  []ConversationSummary
    UserFacts            []UserFact
    TotalTokens          int
    TokenBreakdown       map[string]int
}

type ContextTree struct {
    RootNode     *TreeNode
    TotalItems   int
    LastUpdated  time.Time
}

type TreeNode struct {
    ID       uuid.UUID   `json:"id"`
    Type     string      `json:"type"`      // 'project', 'node', 'category', 'item'
    Name     string      `json:"name"`
    ItemCount int        `json:"item_count"`
    Children []*TreeNode `json:"children,omitempty"`
}

type TokenUsage struct {
    MaxTokens       int
    UsedTokens      int
    AvailableTokens int
    Breakdown       map[string]int // memory: X, context: Y, etc.
}
```
Pedro
### 2.6 Context Tracking for Agents

Roberto's requirement:

> "And then these agents, we need to make sure we track their context windows as well. So the whole flow of all this is going to be very important."

```go
// services/context_tracker.go

type ContextTracker struct {
    pool     *pgxpool.Pool
    sessions map[uuid.UUID]*TrackedSession
    mu       sync.RWMutex
}

type TrackedSession struct {
    SessionID       uuid.UUID
    AgentType       string
    MaxTokens       int
    UsedTokens      int
    LoadedItems     map[uuid.UUID]LoadedItem
    TokenHistory    []TokenSnapshot
}

type LoadedItem struct {
    ItemID     uuid.UUID
    ItemType   string
    Title      string
    TokenCount int
    LoadedAt   time.Time
}

type TokenSnapshot struct {
    Timestamp  time.Time
    UsedTokens int
    Action     string    // 'loaded', 'unloaded', 'response'
    ItemID     *uuid.UUID
}

func (t *ContextTracker) TrackLoad(sessionID uuid.UUID, item ContextItem) error {
    t.mu.Lock()
    defer t.mu.Unlock()

    session := t.sessions[sessionID]
    if session == nil {
        return errors.New("session not found")
    }

    // Check if we have room
    if session.UsedTokens + item.TokenCount > session.MaxTokens {
        return ErrContextWindowFull
    }

    session.LoadedItems[item.ID] = LoadedItem{
        ItemID:     item.ID,
        ItemType:   item.Type,
        Title:      item.Title,
        TokenCount: item.TokenCount,
        LoadedAt:   time.Now(),
    }
    session.UsedTokens += item.TokenCount
    session.TokenHistory = append(session.TokenHistory, TokenSnapshot{
        Timestamp:  time.Now(),
        UsedTokens: session.UsedTokens,
        Action:     "loaded",
        ItemID:     &item.ID,
    })

    return nil
}

func (t *ContextTracker) GetAvailableTokens(sessionID uuid.UUID) int {
    t.mu.RLock()
    defer t.mu.RUnlock()

    session := t.sessions[sessionID]
    if session == nil {
        return 0
    }
    return session.MaxTokens - session.UsedTokens
}

func (t *ContextTracker) UnloadOldestIfNeeded(sessionID uuid.UUID, requiredTokens int) error {
    // Unload oldest items to make room for new context
    // Implements LRU-style eviction
}
```

---

## PART 3: Output Styles & Block System

### 3.1 Core Requirements

Roberto's exact requirements:

> "If you look at the ChatGPT interface, we need to also make sure we have the output styles and stuff because it can't always output in the same way as if it was like a database. Like checklists or some shit like that, or like bullet points, because sometimes right now that's the way it outputs with bullet points and stuff."

> "When sometimes I needed to output in really good formats like paragraph forms or styles because we have a block system that the documents and the text and everything go inside of."

> "So I'd ask to make sure it relates with the blocks as well, and also if there's anything we need to add, then we add those."
nick/ Pedro
### 3.2 Database Schema

```sql
-- ===== OUTPUT STYLES SYSTEM =====

-- Output style templates
CREATE TABLE output_styles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    -- Style Identity
    name VARCHAR(100) NOT NULL UNIQUE,
    display_name VARCHAR(100) NOT NULL,
    description TEXT,
    icon VARCHAR(50),

    -- Style Configuration
    style_type VARCHAR(50) NOT NULL,          -- 'prose', 'structured', 'code', 'mixed'

    -- Formatting Rules
    use_headers BOOLEAN DEFAULT TRUE,
    use_bullets BOOLEAN DEFAULT TRUE,
    use_numbered_lists BOOLEAN DEFAULT FALSE,
    use_paragraphs BOOLEAN DEFAULT TRUE,
    use_code_blocks BOOLEAN DEFAULT FALSE,
    use_tables BOOLEAN DEFAULT FALSE,
    use_blockquotes BOOLEAN DEFAULT FALSE,

    -- Length & Density
    verbosity VARCHAR(20) DEFAULT 'balanced',  -- 'concise', 'balanced', 'detailed', 'comprehensive'
    max_paragraphs INTEGER,
    max_bullets_per_section INTEGER,

    -- Tone
    tone VARCHAR(50) DEFAULT 'professional',   -- 'casual', 'professional', 'formal', 'friendly', 'technical'

    -- System Prompt Addition
    style_instructions TEXT NOT NULL,          -- Instructions to add to system prompt

    -- Block Mapping (how to convert output to blocks)
    block_mapping JSONB DEFAULT '{}',          -- Maps markdown elements to block types

    is_system BOOLEAN DEFAULT FALSE,
    is_active BOOLEAN DEFAULT TRUE,
    sort_order INTEGER DEFAULT 0,

    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- User output style preferences
CREATE TABLE user_output_preferences (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(255) NOT NULL,

    -- Default style
    default_style_id UUID REFERENCES output_styles(id),

    -- Context-specific overrides
    style_overrides JSONB DEFAULT '{}',       -- {"focus_mode:deep": "style_id", "agent:analyst": "style_id"}

    -- Custom instructions
    custom_instructions TEXT,                  -- User's own style preferences

    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),

    UNIQUE(user_id)
);

CREATE INDEX idx_user_output_prefs ON user_output_preferences(user_id);
```
Nick
### 3.3 Pre-seeded Output Styles

```sql
INSERT INTO output_styles (name, display_name, description, icon, style_type, use_headers, use_bullets, use_numbered_lists, use_paragraphs, use_code_blocks, use_tables, verbosity, tone, style_instructions, block_mapping, is_system, sort_order)
VALUES
    -- 1. Conversational (ChatGPT-like casual)
    ('conversational', 'Conversational', 'Natural, flowing conversation like talking to a friend', 'message-circle', 'prose',
     FALSE, FALSE, FALSE, TRUE, FALSE, FALSE,
     'balanced', 'friendly',
     'Respond in a natural, conversational way. Use flowing paragraphs instead of bullet points or lists. Write as if having a friendly conversation. Avoid formal structure - just communicate naturally and clearly. Use "I" and "you" freely. Keep responses warm and approachable.',
     '{"paragraph": "text", "emphasis": "text"}',
     TRUE, 1),

    -- 2. Professional (Business communication)
    ('professional', 'Professional', 'Clear, structured business communication', 'briefcase', 'structured',
     TRUE, TRUE, FALSE, TRUE, FALSE, FALSE,
     'balanced', 'professional',
     'Respond in a professional business style. Use clear structure with headers for main sections. Use bullet points for lists of items or key points. Keep language formal but accessible. Be direct and actionable.',
     '{"h2": "heading", "bullet": "bullet_list", "paragraph": "text"}',
     TRUE, 2),

    -- 3. Technical (Developer documentation)
    ('technical', 'Technical', 'Precise technical documentation with code examples', 'code', 'code',
     TRUE, TRUE, FALSE, TRUE, TRUE, TRUE,
     'detailed', 'technical',
     'Respond in a technical documentation style. Include code examples where relevant. Use precise technical terminology. Structure with clear headers. Use tables for comparisons or specifications. Include type definitions and API signatures when applicable.',
     '{"h2": "heading", "code": "code", "table": "table", "bullet": "bullet_list"}',
     TRUE, 3),

    -- 4. Executive Summary (Brief, high-level)
    ('executive', 'Executive Summary', 'Brief, high-level summaries for quick consumption', 'zap', 'structured',
     FALSE, TRUE, FALSE, TRUE, FALSE, FALSE,
     'concise', 'formal',
     'Provide brief, executive-level summaries. Lead with the key takeaway or recommendation. Use 3-5 bullet points maximum for supporting details. Avoid technical jargon. Focus on business impact and actionable insights. Keep total response under 200 words.',
     '{"bullet": "bullet_list", "paragraph": "text"}',
     TRUE, 4),

    -- 5. Detailed Analysis (Comprehensive reports)
    ('detailed', 'Detailed Analysis', 'Comprehensive, in-depth analysis with full context', 'file-text', 'structured',
     TRUE, TRUE, TRUE, TRUE, FALSE, TRUE,
     'comprehensive', 'professional',
     'Provide comprehensive, detailed analysis. Use clear section headers. Include numbered lists for sequential steps or rankings. Use tables for data comparisons. Provide context and background. Include considerations, trade-offs, and recommendations. Be thorough but organized.',
     '{"h2": "heading", "h3": "subheading", "numbered": "numbered_list", "bullet": "bullet_list", "table": "table", "paragraph": "text"}',
     TRUE, 5),

    -- 6. Creative (Engaging, story-like)
    ('creative', 'Creative', 'Engaging, narrative style for creative content', 'sparkles', 'prose',
     FALSE, FALSE, FALSE, TRUE, FALSE, FALSE,
     'detailed', 'casual',
     'Write in an engaging, creative style. Use narrative techniques - metaphors, vivid descriptions, storytelling. Create flow and rhythm in the writing. Engage the reader emotionally. Avoid dry, factual presentation. Make the content memorable and enjoyable to read.',
     '{"paragraph": "text", "blockquote": "quote"}',
     TRUE, 6),

    -- 7. Step-by-Step (Tutorial/how-to)
    ('tutorial', 'Step-by-Step', 'Clear tutorial style with numbered steps', 'list-ordered', 'structured',
     TRUE, FALSE, TRUE, TRUE, TRUE, FALSE,
     'detailed', 'friendly',
     'Write in a clear tutorial style. Use numbered steps for any process or procedure. Include code examples with explanations. Add tips or notes for important points. Anticipate common questions or issues. Make sure each step is clear and actionable.',
     '{"h2": "heading", "numbered": "numbered_list", "code": "code", "note": "callout", "paragraph": "text"}',
     TRUE, 7),

    -- 8. Q&A (Direct answers)
    ('qa', 'Q&A', 'Direct question-and-answer format', 'help-circle', 'mixed',
     FALSE, FALSE, FALSE, TRUE, FALSE, FALSE,
     'concise', 'friendly',
     'Answer directly and concisely. Start with the direct answer to the question. Then provide brief supporting context if needed. Avoid unnecessary preamble. If multiple questions, address each clearly. Use simple, clear language.',
     '{"paragraph": "text"}',
     TRUE, 8)

ON CONFLICT (name) DO UPDATE SET
    display_name = EXCLUDED.display_name,
    description = EXCLUDED.description,
    style_instructions = EXCLUDED.style_instructions,
    block_mapping = EXCLUDED.block_mapping,
    updated_at = NOW();
```
Pedro
### 3.4 Block Type Mapping

```go
// services/block_mapper.go

type BlockMapper struct {
    mappings map[string]BlockMapping
}

type BlockMapping struct {
    MarkdownElement string   // 'h1', 'h2', 'bullet', 'numbered', 'code', 'paragraph', 'table', 'blockquote'
    BlockType       string   // 'heading', 'text', 'bullet_list', 'numbered_list', 'code', 'table', 'quote', 'callout'
    Transform       func(content string) Block
}

type Block struct {
    ID         string                 `json:"id"`
    Type       string                 `json:"type"`
    Content    string                 `json:"content"`
    Properties map[string]interface{} `json:"properties,omitempty"`
    Children   []Block                `json:"children,omitempty"`
}

// ConvertResponseToBlocks converts AI response to block format
func (m *BlockMapper) ConvertResponseToBlocks(response string, style OutputStyle) ([]Block, error) {
    // Parse markdown
    // Map to block types based on style.BlockMapping
    // Return structured blocks
}

// Available block types:
// - heading (h1, h2, h3)
// - subheading
// - text (paragraph)
// - bullet_list
// - numbered_list
// - checklist
// - code
// - table
// - quote
// - callout (info, warning, tip)
// - divider
// - image
// - enickmbed
```
nick / pedro
### 3.5 API Endpoints

```
# Output Styles
GET    /api/output-styles                     # List all styles
GET    /api/output-styles/:id                 # Get specific style
POST   /api/output-styles                     # Create custom style (if allowed)

# User Preferences
GET    /api/user/output-preferences           # Get user's preferences
PUT    /api/user/output-preferences           # Update preferences
PUT    /api/user/output-preferences/default   # Set default style

# Block Conversion
POST   /api/blocks/convert                    # Convert text to blocks
```

---

## PART 4: Chat History, Voice Notes & Project Context Integration

### 4.1 Core Requirements

Roberto's exact requirements:

> "It's not just voice notes. It's also the chat history that you have with the system. The voice notes and the chat history might not always be the same."

> "I need to make sure that the voice thing is also saved. The voice notes are saved as well, those are part of the context depending on the project."

> "Because remember you have to select a project to use the system that's something important that you know is important here that's already there."

> "And then when you have a project, you can always select a node. You don't always have to select a node but you can."

> "And so the node system context and everything the database is to be set up perfectly where when that node is selected, that particular context is loaded into that model or into that conversation for example with the system."

**Critical Clarification**: Both chat history AND voice notes are context sources. They may contain different information and both need to be searchable and loadable into agent context.
Pedro
### 4.2 Chat History as Context Source

```sql
-- Add context source flags to conversations table (extend existing)
ALTER TABLE conversations ADD COLUMN IF NOT EXISTS is_context_source BOOLEAN DEFAULT TRUE;
ALTER TABLE conversations ADD COLUMN IF NOT EXISTS extracted_memories UUID[] DEFAULT '{}';
ALTER TABLE conversations ADD COLUMN IF NOT EXISTS summary TEXT;                    -- AI-generated summary
ALTER TABLE conversations ADD COLUMN IF NOT EXISTS key_topics TEXT[] DEFAULT '{}'; -- Extracted topics
ALTER TABLE conversations ADD COLUMN IF NOT EXISTS embedding vector(1536);         -- For semantic search

CREATE INDEX IF NOT EXISTS idx_conversations_context ON conversations(user_id, is_context_source);
CREATE INDEX IF NOT EXISTS idx_conversations_embedding ON conversations USING ivfflat (embedding vector_cosine_ops);

-- Conversation summaries for efficient context loading
CREATE TABLE conversation_summaries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    conversation_id UUID NOT NULL REFERENCES conversations(id) ON DELETE CASCADE,
    user_id VARCHAR(255) NOT NULL,

    -- Summary content
    summary TEXT NOT NULL,
    key_points TEXT[] DEFAULT '{}',
    decisions_made TEXT[] DEFAULT '{}',
    action_items TEXT[] DEFAULT '{}',

    -- Topics and entities
    topics TEXT[] DEFAULT '{}',
    mentioned_entities JSONB DEFAULT '{}',  -- {"projects": [], "clients": [], "tasks": []}

    -- Embedding for semantic search
    embedding vector(1536),

    -- Timestamps
    summarized_at TIMESTAMPTZ DEFAULT NOW(),
    message_count INTEGER,
    time_range TSTZRANGE             -- Start to end of conversation
);

CREATE INDEX idx_conv_summaries_conv ON conversation_summaries(conversation_id);
CREATE INDEX idx_conv_summaries_user ON conversation_summaries(user_id);
```
nick / Pedro
### 4.3 Voice Note Context Integration

```sql
-- Add context linking to voice_notes table (extend existing)
ALTER TABLE voice_notes ADD COLUMN IF NOT EXISTS project_id UUID REFERENCES projects(id) ON DELETE SET NULL;
ALTER TABLE voice_notes ADD COLUMN IF NOT EXISTS node_id UUID REFERENCES nodes(id) ON DELETE SET NULL;
ALTER TABLE voice_notes ADD COLUMN IF NOT EXISTS is_context_source BOOLEAN DEFAULT TRUE;
ALTER TABLE voice_notes ADD COLUMN IF NOT EXISTS extracted_memories UUID[] DEFAULT '{}';
ALTER TABLE voice_notes ADD COLUMN IF NOT EXISTS embedding vector(1536);

CREATE INDEX IF NOT EXISTS idx_voice_notes_project ON voice_notes(project_id);
CREATE INDEX IF NOT EXISTS idx_voice_notes_node ON voice_notes(node_id);
```
Pedro
### 4.4 File Upload & Document Management

Roberto's clarification:
> "I need to make sure it's not just text where people can upload files and stuff"

```sql
-- ===== FILE UPLOAD & DOCUMENT MANAGEMENT =====

-- Uploaded documents (PDFs, markdown, etc.)
CREATE TABLE uploaded_documents (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(255) NOT NULL,

    -- Document identity
    filename VARCHAR(500) NOT NULL,
    original_filename VARCHAR(500) NOT NULL,
    display_name VARCHAR(255),
    description TEXT,

    -- File info
    file_type VARCHAR(50) NOT NULL,          -- 'pdf', 'markdown', 'docx', 'txt', 'image'
    mime_type VARCHAR(255) NOT NULL,
    file_size_bytes BIGINT NOT NULL,

    -- Storage
    storage_path VARCHAR(1000) NOT NULL,     -- Path in storage (GCS, S3, local)
    storage_provider VARCHAR(50) DEFAULT 'local',  -- 'local', 'gcs', 's3'

    -- Extracted content
    extracted_text TEXT,                      -- Full text extraction from PDF/DOCX
    page_count INTEGER,                       -- For PDFs
    word_count INTEGER,

    -- Context profile link
    context_profile_id UUID REFERENCES context_profiles(id) ON DELETE SET NULL,
    project_id UUID REFERENCES projects(id) ON DELETE SET NULL,
    node_id UUID REFERENCES nodes(id) ON DELETE SET NULL,

    -- Categorization
    document_type VARCHAR(100),               -- 'sop', 'framework', 'template', 'reference', 'report'
    category VARCHAR(100),
    tags TEXT[] DEFAULT '{}',

    -- Semantic search
    embedding vector(1536),

    -- Processing status
    processing_status VARCHAR(50) DEFAULT 'pending',  -- 'pending', 'processing', 'completed', 'failed'
    processed_at TIMESTAMPTZ,

    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_uploaded_docs_user ON uploaded_documents(user_id);
CREATE INDEX idx_uploaded_docs_profile ON uploaded_documents(context_profile_id);
CREATE INDEX idx_uploaded_docs_project ON uploaded_documents(project_id);
CREATE INDEX idx_uploaded_docs_type ON uploaded_documents(document_type);
CREATE INDEX idx_uploaded_docs_embedding ON uploaded_documents USING ivfflat (embedding vector_cosine_ops);

-- Document chunks for large documents (split for better retrieval)
CREATE TABLE document_chunks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    document_id UUID NOT NULL REFERENCES uploaded_documents(id) ON DELETE CASCADE,

    -- Chunk info
    chunk_index INTEGER NOT NULL,
    content TEXT NOT NULL,
    token_count INTEGER,

    -- Position in original
    page_number INTEGER,                      -- For PDFs
    section_title VARCHAR(255),               -- If extractable

    -- Embedding for semantic search
    embedding vector(1536),

    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_doc_chunks_document ON document_chunks(document_id);
CREATE INDEX idx_doc_chunks_embedding ON document_chunks USING ivfflat (embedding vector_cosine_ops);
```
Pedro
### 4.5 Context Profiles with Multiple Document Types

```sql
-- Update context_profiles to handle document types better
ALTER TABLE context_profiles ADD COLUMN IF NOT EXISTS document_types TEXT[] DEFAULT '{}';  -- ['pdf', 'markdown', 'sop']
ALTER TABLE context_profiles ADD COLUMN IF NOT EXISTS total_documents INTEGER DEFAULT 0;
ALTER TABLE context_profiles ADD COLUMN IF NOT EXISTS total_file_size_bytes BIGINT DEFAULT 0;

-- Context profile items - links various content types to profiles
CREATE TABLE context_profile_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    context_profile_id UUID NOT NULL REFERENCES context_profiles(id) ON DELETE CASCADE,
    user_id VARCHAR(255) NOT NULL,

    -- Item reference
    item_type VARCHAR(50) NOT NULL,           -- 'document', 'artifact', 'memory', 'conversation', 'voice_note'
    item_id UUID NOT NULL,

    -- Display info
    display_name VARCHAR(255),
    description TEXT,

    -- Ordering
    sort_order INTEGER DEFAULT 0,
    is_pinned BOOLEAN DEFAULT FALSE,

    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_profile_items_profile ON context_profile_items(context_profile_id);
CREATE INDEX idx_profile_items_type ON context_profile_items(item_type, item_id);
```
Pedro
### 4.6 Document Processing Service

```go
// services/document_processor.go

type DocumentProcessor interface {
    // Upload & Processing
    UploadDocument(ctx context.Context, userID string, file io.Reader, metadata DocumentMetadata) (*UploadedDocument, error)
    ProcessDocument(ctx context.Context, docID uuid.UUID) error

    // Text extraction
    ExtractTextFromPDF(ctx context.Context, filePath string) (string, []PageContent, error)
    ExtractTextFromDocx(ctx context.Context, filePath string) (string, error)
    ExtractTextFromMarkdown(ctx context.Context, filePath string) (string, error)

    // Chunking for retrieval
    ChunkDocument(ctx context.Context, docID uuid.UUID, chunkSize int) ([]DocumentChunk, error)

    // Search
    SearchDocuments(ctx context.Context, userID string, query string, opts SearchOpts) ([]DocumentSearchResult, error)
    GetDocumentContent(ctx context.Context, docID uuid.UUID) (*DocumentContent, error)
}

type DocumentMetadata struct {
    Filename        string
    DisplayName     string
    Description     string
    DocumentType    string    // 'sop', 'framework', 'template', etc.
    Category        string
    Tags            []string
    ContextProfileID *uuid.UUID
    ProjectID       *uuid.UUID
    NodeID          *uuid.UUID
}

type DocumentContent struct {
    Document     *UploadedDocument
    FullText     string
    Chunks       []DocumentChunk
    TokenCount   int
}

type PageContent struct {
    PageNumber int
    Content    string
    TokenCount int
}
```
nick / pedro
### 4.7 File Upload API Endpoints

```
# Document Upload & Management
POST   /api/documents/upload              # Upload a document
GET    /api/documents                     # List user's documents
GET    /api/documents/:id                 # Get document details
GET    /api/documents/:id/content         # Get document content/text
DELETE /api/documents/:id                 # Delete document

# Document Search
POST   /api/documents/search              # Semantic search documents

# Context Profile Document Management
POST   /api/context-profiles/:id/documents    # Add document to profile
GET    /api/context-profiles/:id/documents    # List documents in profile
DELETE /api/context-profiles/:id/documents/:docId  # Remove from profile
```
Nick / Pedro
### 4.8 Project/Node Context Loading

```go
// services/project_context.go

type ProjectContextService struct {
    pool            *pgxpool.Pool
    memoryService   MemoryService
    contextService  ContextService
}

// LoadProjectContext loads all relevant context when a project is selected
func (s *ProjectContextService) LoadProjectContext(ctx context.Context, userID string, projectID uuid.UUID) (*ProjectContext, error) {
    // 1. Get project details
    project, _ := s.getProject(ctx, userID, projectID)

    // 2. Get project's context profile
    profile, _ := s.contextService.GetContextProfile(ctx, userID, "project", projectID)

    // 3. Load memories associated with project
    memories, _ := s.memoryService.GetMemoriesForProject(ctx, userID, projectID)

    // 4. Load KB contexts linked to project
    contexts, _ := s.loadProjectContexts(ctx, userID, projectID)

    // 5. Get recent voice notes for project
    voiceNotes, _ := s.getProjectVoiceNotes(ctx, userID, projectID)

    // 6. Get recent conversations in project
    conversations, _ := s.getRecentProjectConversations(ctx, userID, projectID)

    return &ProjectContext{
        Project:       project,
        Profile:       profile,
        Memories:      memories,
        Contexts:      contexts,
        VoiceNotes:    voiceNotes,
        Conversations: conversations,
    }, nil
}

// LoadNodeContext loads context when a specific node is selected
func (s *ProjectContextService) LoadNodeContext(ctx context.Context, userID string, nodeID uuid.UUID) (*NodeContext, error) {
    // 1. Get node details and its ancestors
    node, _ := s.getNodeWithAncestors(ctx, userID, nodeID)

    // 2. Get node's context profile
    profile, _ := s.contextService.GetContextProfile(ctx, userID, "node", nodeID)

    // 3. Load memories for this specific node
    memories, _ := s.memoryService.GetMemoriesForNode(ctx, userID, nodeID)

    // 4. Inherit relevant parent context
    parentContext, _ := s.loadParentNodeContext(ctx, userID, node.ParentID)

    return &NodeContext{
        Node:          node,
        Ancestors:     node.Ancestors,
        Profile:       profile,
        Memories:      memories,
        ParentContext: parentContext,
    }, nil
}
```
Pedro
### 4.9 Conversation Context Injection

When a conversation starts or context changes:

```go
// handlers/chat_v2.go - Updated context injection

func (h *Handlers) injectContextIntoConversation(
    ctx context.Context,
    userID string,
    projectID *uuid.UUID,
    nodeID *uuid.UUID,
    agentType string,
    focusMode string,
) (*InjectedContext, error) {

    contextBuilder := services.NewContextBuilder(h.pool)

    input := services.ContextBuildInput{
        UserID:    userID,
        ProjectID: projectID,
        NodeID:    nodeID,
        AgentType: agentType,
        FocusMode: focusMode,
        MaxTokens: getModelMaxTokens(agentType),
    }

    // Build context based on project/node selection
    agentContext, err := contextBuilder.BuildContextForAgent(ctx, userID, input)
    if err != nil {
        return nil, err
    }

    // Format for system prompt injection
    contextPrompt := formatContextForInjection(agentContext)

    return &InjectedContext{
        SystemPromptAddition: contextPrompt,
        LoadedItems:          agentContext.LoadedMemories,
        TotalTokens:          agentContext.TotalTokens,
    }, nil
}
```

---

## PART 5: Context Tree Visualization

### 5.1 Core Requirements

Roberto's exact requirements:

> "So we need all of that type of stuff stored where we create a context profile on the application where it's like a tree in a way."

> "We need to make sure there's a way for the users to see that tree of how their context might be organized based off of let's just say if it's the tasks right or let's just say if it's the module for their projects right. They should be able to see the tree breakdown and all that stuff in some sort of way."
nick / pedro
### 5.2 API Endpoints for Tree Visualization

```
# Context Tree API
GET    /api/context-tree                          # Get user's full context tree
GET    /api/context-tree/project/:projectId       # Get tree for specific project
GET    /api/context-tree/node/:nodeId             # Get tree for specific node

# Tree Statistics
GET    /api/context-tree/stats                    # Overall stats
GET    /api/context-tree/stats/project/:projectId # Project-specific stats
```
pedro
### 5.3 Tree Response Format

```go
type ContextTreeResponse struct {
    Tree       *ContextTreeNode `json:"tree"`
    Statistics TreeStatistics   `json:"statistics"`
    LastUpdate time.Time        `json:"last_update"`
}

type ContextTreeNode struct {
    ID          uuid.UUID          `json:"id"`
    Type        string             `json:"type"`       // 'root', 'project', 'node', 'category', 'item'
    Name        string             `json:"name"`
    Description string             `json:"description,omitempty"`
    Icon        string             `json:"icon,omitempty"`
    ItemCount   int                `json:"item_count"` // Number of items at this level
    TokenCount  int                `json:"token_count"` // Total tokens if all loaded
    Children    []*ContextTreeNode `json:"children,omitempty"`
    Metadata    map[string]any     `json:"metadata,omitempty"`
}

type TreeStatistics struct {
    TotalProjects   int            `json:"total_projects"`
    TotalNodes      int            `json:"total_nodes"`
    TotalMemories   int            `json:"total_memories"`
    TotalContexts   int            `json:"total_contexts"`
    TotalArtifacts  int            `json:"total_artifacts"`
    TotalVoiceNotes int            `json:"total_voice_notes"`
    ByType          map[string]int `json:"by_type"`
    TotalTokens     int            `json:"total_tokens"`
}
```

### 5.4 Example Tree Structure

**Note**: The tree starts at Node level (most comprehensive), then Projects, then Context Profiles, then individual items.

```json
{
  "tree": {
    "id": "node-main",
    "type": "node",
    "name": "MIOSA Platform",
    "description": "Main development node",
    "item_count": 256,
    "children": [
      {
        "id": "proj-businessos",
        "type": "project",
        "name": "BusinessOS Development",
        "icon": "folder",
        "item_count": 89,
        "children": [
          {
            "id": "profile-frontend",
            "type": "context_profile",
            "name": "Frontend Development",
            "description": "All frontend-related context",
            "item_count": 34,
            "children": [
              {
                "id": "doc-svelte-guide",
                "type": "document",
                "name": "SvelteKit Best Practices.pdf",
                "icon": "file-pdf",
                "document_type": "reference"
              },
              {
                "id": "doc-component-sop",
                "type": "document",
                "name": "Component Creation SOP.md",
                "icon": "file-text",
                "document_type": "sop"
              },
              {
                "id": "artifact-1",
                "type": "artifact",
                "name": "Chat Component Spec",
                "icon": "file-code"
              }
            ]
          },
          {
            "id": "profile-backend",
            "type": "context_profile",
            "name": "Backend Development",
            "description": "Go backend context",
            "item_count": 42,
            "children": [
              {
                "id": "doc-go-patterns",
                "type": "document",
                "name": "Go Patterns Guide.pdf",
                "icon": "file-pdf",
                "document_type": "reference"
              }
            ]
          },
          {
            "id": "cat-chat-history",
            "type": "category",
            "name": "Chat History",
            "icon": "message-square",
            "item_count": 28
          },
          {
            "id": "cat-voice-notes",
            "type": "category",
            "name": "Voice Notes",
            "icon": "mic",
            "item_count": 15
          },
          {
            "id": "cat-memories",
            "type": "category",
            "name": "Memories",
            "icon": "brain",
            "item_count": 45
          }
        ]
      }
    ]
  },
  "statistics": {
    "total_nodes": 5,
    "total_projects": 12,
    "total_context_profiles": 34,
    "total_documents": 89,
    "total_memories": 156,
    "total_artifacts": 78,
    "total_conversations": 234,
    "total_voice_notes": 45,
    "total_tokens": 1245000,
    "by_document_type": {
      "pdf": 23,
      "markdown": 45,
      "sop": 12,
      "framework": 5,
      "template": 4
    }
  }
}
```

---

## PART 6: Self-Learning System

### 6.1 Core Requirements

Roberto's exact requirements:

> "And then all of that we need to have a way for the agents to be able to perfectly sequentially get that data. It's going to be like some sort of semantic search and so the system will be technically self-learning."

> "So we need to implement the self-learning aspect into this as well so that it knows and grows about the user."

> "I don't want generic ass outputs like they have to be real ass outputs."
nick / pedro
### 6.2 Database Schema

```sql
-- ===== SELF-LEARNING SYSTEM =====

-- Learning events - tracks what the system learns
CREATE TABLE learning_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(255) NOT NULL,

    -- What was learned
    learning_type VARCHAR(50) NOT NULL,       -- 'correction', 'preference', 'pattern', 'feedback', 'behavior'
    learning_content TEXT NOT NULL,

    -- Source of learning
    source_type VARCHAR(50) NOT NULL,         -- 'explicit_feedback', 'implicit_behavior', 'correction', 'conversation'
    source_id UUID,
    source_context TEXT,

    -- Confidence & Application
    confidence_score DECIMAL(3,2) DEFAULT 0.5,
    times_applied INTEGER DEFAULT 0,
    last_applied_at TIMESTAMPTZ,

    -- Whether it resulted in a memory or fact
    created_memory_id UUID REFERENCES memories(id),
    created_fact_key VARCHAR(255),

    -- Validation
    was_validated BOOLEAN DEFAULT FALSE,
    validated_at TIMESTAMPTZ,
    validation_result VARCHAR(50),            -- 'confirmed', 'rejected', 'modified'

    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_learning_events_user ON learning_events(user_id);
CREATE INDEX idx_learning_events_type ON learning_events(learning_type);

-- User behavior patterns - observed patterns in user behavior
CREATE TABLE user_behavior_patterns (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(255) NOT NULL,

    -- Pattern details
    pattern_type VARCHAR(100) NOT NULL,       -- 'time_preference', 'topic_interest', 'communication_style', 'tool_usage'
    pattern_key VARCHAR(255) NOT NULL,
    pattern_value TEXT NOT NULL,

    -- Evidence
    observation_count INTEGER DEFAULT 1,
    first_observed_at TIMESTAMPTZ DEFAULT NOW(),
    last_observed_at TIMESTAMPTZ DEFAULT NOW(),
    evidence_ids UUID[],                      -- References to supporting data

    -- Confidence
    confidence_score DECIMAL(3,2) DEFAULT 0.5,

    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),

    UNIQUE(user_id, pattern_type, pattern_key)
);

CREATE INDEX idx_behavior_patterns_user ON user_behavior_patterns(user_id);

-- Feedback log - explicit user feedback on AI outputs
CREATE TABLE feedback_log (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(255) NOT NULL,

    -- What feedback is about
    target_type VARCHAR(50) NOT NULL,         -- 'message', 'artifact', 'memory', 'suggestion'
    target_id UUID NOT NULL,

    -- Feedback content
    feedback_type VARCHAR(50) NOT NULL,       -- 'thumbs_up', 'thumbs_down', 'correction', 'comment'
    feedback_value TEXT,

    -- Context
    conversation_id UUID,
    agent_type VARCHAR(100),

    -- Processing
    was_processed BOOLEAN DEFAULT FALSE,
    resulting_learning_id UUID REFERENCES learning_events(id),

    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_feedback_log_user ON feedback_log(user_id);
CREATE INDEX idx_feedback_log_target ON feedback_log(target_type, target_id);
```
Pedro
### 6.3 Learning Service

```go
// services/learning.go

type LearningService interface {
    // Learning from interactions
    LearnFromConversation(ctx context.Context, userID string, conversationID uuid.UUID) error
    LearnFromFeedback(ctx context.Context, userID string, feedback FeedbackInput) error
    LearnFromCorrection(ctx context.Context, userID string, original string, correction string, context string) error

    // Pattern detection
    DetectBehaviorPatterns(ctx context.Context, userID string) ([]BehaviorPattern, error)
    UpdatePatternConfidence(ctx context.Context, patternID uuid.UUID, observation bool) error

    // Applying learnings
    GetUserLearnings(ctx context.Context, userID string) (*UserLearnings, error)
    GetRelevantLearnings(ctx context.Context, userID string, context LearningContext) ([]Learning, error)
    ApplyLearningsToPrompt(ctx context.Context, userID string, basePrompt string) (string, error)

    // Personalization
    GetPersonalizationProfile(ctx context.Context, userID string) (*PersonalizationProfile, error)
    UpdatePersonalizationProfile(ctx context.Context, userID string) error
}

type UserLearnings struct {
    Facts              []UserFact
    Preferences        []Learning
    CommunicationStyle *CommunicationStyle
    BehaviorPatterns   []BehaviorPattern
    RecentCorrections  []Learning
}

type PersonalizationProfile struct {
    UserID             string
    PreferredTone      string              // 'formal', 'casual', 'professional', 'friendly'
    PreferredVerbosity string              // 'concise', 'balanced', 'detailed'
    PreferredFormat    string              // 'prose', 'bullets', 'structured'
    ExpertiseAreas     []string            // Topics user knows well
    LearningAreas      []string            // Topics user is learning
    CommonTopics       []string            // Frequently discussed topics
    Timezone           string
    WorkingHours       *WorkingHours
    LastUpdated        time.Time
}

type CommunicationStyle struct {
    Formality         float64             // 0 = casual, 1 = formal
    Technicality      float64             // 0 = simple, 1 = technical
    DetailLevel       float64             // 0 = brief, 1 = comprehensive
    PrefersBullets    bool
    PrefersExamples   bool
    PrefersAnalogies  bool
}
```
nick / pedro
### 6.4 Automatic Learning Triggers

```go
// services/learning_triggers.go

type LearningTrigger struct {
    service LearningService
}

// AfterConversation - called after each conversation turn
func (t *LearningTrigger) AfterConversation(ctx context.Context, userID string, message Message, response Message) {
    // Extract potential facts from conversation
    facts := t.extractFacts(message.Content, response.Content)
    for _, fact := range facts {
        t.service.LearnFromConversation(ctx, userID, fact)
    }

    // Detect communication preferences
    if len(response.Content) > 500 && !containsBullets(response.Content) {
        t.service.UpdatePatternConfidence(ctx, "prefers_prose", true)
    }
}

// OnFeedback - called when user gives explicit feedback
func (t *LearningTrigger) OnFeedback(ctx context.Context, userID string, feedback Feedback) {
    if feedback.Type == "thumbs_down" {
        // Learn what went wrong
        t.service.LearnFromFeedback(ctx, userID, FeedbackInput{
            Type:    "negative",
            Context: feedback.Context,
        })
    }
}

// OnCorrection - called when user corrects AI output
func (t *LearningTrigger) OnCorrection(ctx context.Context, userID string, original, corrected string) {
    t.service.LearnFromCorrection(ctx, userID, original, corrected, "")

    // If correction is about terminology, create a memory
    if isTerminologyCorrection(original, corrected) {
        // Create memory about correct terminology
    }
}
```

---
Nick
## PART 7: Orchestrator System Prompt Enhancement

### 7.1 Core Requirements

Roberto's exact requirements:

> "We need to improve the system prompt of the orchestrator dramatically because it has to be very precise with all of this stuff as well to know how to use particular things and blah blah blah."

> "And we need to make sure that all the context tracking is set up correctly."

### 7.2 Enhanced Orchestrator Prompt

Create/update: `internal/prompts/orchestrator_v2.go`

```go
package prompts

const OrchestratorV2SystemPrompt = `You are OSA (Operating System Agent), the intelligent orchestrator for BusinessOS.

## Your Core Identity

You are not a generic AI assistant. You are a personalized business partner who:
- KNOWS the user deeply through memories and learned patterns
- UNDERSTANDS the current context (project, node, conversation history)
- ADAPTS your communication style to the user's preferences
- RETRIEVES relevant information before responding
- LEARNS from every interaction

## Context Awareness

### Current Context
{{if .Project}}
**Active Project:** {{.Project.Name}}
{{.Project.Description}}
{{end}}

{{if .Node}}
**Current Node:** {{.Node.Name}}
Path: {{.Node.Path}}
{{end}}

### User Profile
{{if .UserProfile}}
**Communication Style:** {{.UserProfile.PreferredTone}} | {{.UserProfile.PreferredVerbosity}}
**Expertise:** {{join .UserProfile.ExpertiseAreas ", "}}
**Output Preference:** {{.UserProfile.PreferredFormat}}
{{end}}

### Active Memories
{{range .RelevantMemories}}
- [{{.Type}}] {{.Summary}}
{{end}}

## Your Tools

You have access to these tools to help you serve the user:

### Context Tools
- **tree_search**: Search through knowledge base, memories, and documents
  - Use this to find relevant information before answering
  - Search by title, content, or semantically

- **load_context**: Load a specific document or memory into your context
  - Use after finding relevant items with tree_search

- **browse_tree**: Browse the user's context hierarchy
  - Use to understand what information is available

### Memory Tools
- **save_memory**: Save important information as a memory
  - Use when user shares important facts, decisions, or preferences
  - Categorize appropriately (fact, preference, decision, etc.)

- **recall_memory**: Recall specific memories by topic
  - Use when you need to remember past discussions or decisions

## Response Guidelines

### BEFORE Responding
1. **Check Context**: Is there relevant information in the current project/node context?
2. **Search if Needed**: Use tree_search to find related documents or memories
3. **Load Relevant**: Use load_context to pull in specific information
4. **Consider History**: Review recent conversation and any corrections

### WHILE Responding
1. **Match Style**: Use the user's preferred communication style
2. **Be Specific**: Reference specific facts, not generic advice
3. **Stay Contextual**: Relate responses to the current project/node
4. **Show Sources**: Mention when using specific memories or documents

### AFTER Key Interactions
1. **Save Important Info**: If user shares facts, preferences, or decisions, save as memory
2. **Note Corrections**: If user corrects you, learn from it
3. **Update Patterns**: Note communication preferences

## Output Format

{{if eq .OutputStyle "conversational"}}
Write naturally in flowing paragraphs. Be warm and conversational. Avoid bullet points and formal structure.
{{else if eq .OutputStyle "professional"}}
Use clear structure with headers and bullet points. Be professional and actionable.
{{else if eq .OutputStyle "technical"}}
Include code examples and technical precision. Use proper formatting.
{{else if eq .OutputStyle "executive"}}
Be brief and high-level. Lead with recommendations. Maximum 200 words.
{{else if eq .OutputStyle "detailed"}}
Provide comprehensive analysis with sections, examples, and thorough coverage.
{{end}}

## Critical Rules

1. **Never be generic** - Use what you know about the user
2. **Always check context first** - Don't answer from pure training data when context exists
3. **Cite your sources** - When using memories or documents, mention them
4. **Learn actively** - Save important information, note preferences
5. **Adapt continuously** - Your responses should evolve as you learn
6. **Track your context window** - Be aware of what you've loaded

## Agent Delegation

When the task requires specialized expertise, you can delegate to:
{{range .AvailableAgents}}
- **@{{.Name}}**: {{.Description}}
{{end}}

Delegate when:
- Task requires deep expertise in a specific domain
- User explicitly mentions an agent
- Focus mode suggests a specialist

## Remember

You are not starting fresh. You have history with this user. You know their:
- Projects and what they're working on
- Preferences and how they like to communicate
- Past decisions and why they made them
- Patterns and how they typically work

Use this knowledge. Be the personalized business partner they need.
`
```

### 7.3 Context Injection for Orchestrator

```go
// prompts/composer_v2.go

type OrchestratorPromptInput struct {
    // Current context
    Project          *Project
    Node             *Node

    // User personalization
    UserProfile      *PersonalizationProfile
    UserFacts        []UserFact

    // Relevant memories
    RelevantMemories []Memory

    // Output style
    OutputStyle      string

    // Available agents for delegation
    AvailableAgents  []AgentInfo

    // Session tracking
    LoadedContexts   []ContextItem
    TokenBudget      int
    TokensUsed       int
}

func ComposeOrchestratorPrompt(input OrchestratorPromptInput) (string, error) {
    tmpl, err := template.New("orchestrator").Parse(OrchestratorV2SystemPrompt)
    if err != nil {
        return "", err
    }

    var buf bytes.Buffer
    if err := tmpl.Execute(&buf, input); err != nil {
        return "", err
    }

    return buf.String(), nil
}
```

---

## PART 8: Future Integration Preparation

### 8.1 Core Requirements

Roberto's exact requirements:

> "I need to make it so that we set this up perfectly where if we want to integrate in the future, now we're going to integrate our OS agent, our actual system for a coding model."

> "I need to have it be able to access the context of this business OS for example and understand where everything is, what the icons and components are, what are the modules that are developed inside of it."
nick / pedro
### 8.2 Application Context Profile Schema

```sql
-- ===== APPLICATION CONTEXT PROFILES =====

-- Store context profiles for applications/codebases
CREATE TABLE application_profiles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(255) NOT NULL,

    -- Application Identity
    name VARCHAR(255) NOT NULL,
    description TEXT,
    app_type VARCHAR(100),                    -- 'web_app', 'mobile_app', 'api', 'library', 'platform'

    -- Tech Stack
    tech_stack JSONB DEFAULT '{}',            -- {"frontend": "svelte", "backend": "go", "database": "postgres"}

    -- Structure (tree representation of the app)
    structure_tree JSONB NOT NULL DEFAULT '{}',

    -- Components Registry
    components JSONB DEFAULT '[]',            -- [{name, path, description, props}]

    -- Modules Registry
    modules JSONB DEFAULT '[]',               -- [{name, path, description, exports}]

    -- Icons/Assets Registry
    icons JSONB DEFAULT '[]',                 -- [{name, path, usage}]

    -- API Endpoints
    api_endpoints JSONB DEFAULT '[]',         -- [{method, path, description, params}]

    -- Database Schema Summary
    database_schema JSONB DEFAULT '{}',

    -- Conventions & Patterns
    conventions JSONB DEFAULT '{}',           -- {"naming": {}, "structure": {}, "patterns": []}

    -- Integration Points
    integration_points JSONB DEFAULT '[]',    -- External services, APIs

    -- Embeddings for semantic search
    embedding vector(1536),

    -- Sync
    last_synced_at TIMESTAMPTZ,
    sync_source VARCHAR(255),                 -- Git repo, file path, etc.

    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_app_profiles_user ON application_profiles(user_id);
CREATE INDEX idx_app_profiles_type ON application_profiles(app_type);
```
Pedro
### 8.3 Auto-Profile Generation

```go
// services/app_profiler.go

type AppProfiler interface {
    // Generate profile from codebase
    ProfileFromDirectory(ctx context.Context, userID string, path string) (*ApplicationProfile, error)
    ProfileFromGitRepo(ctx context.Context, userID string, repoURL string) (*ApplicationProfile, error)

    // Update profile
    UpdateProfile(ctx context.Context, profileID uuid.UUID) error
    SyncProfile(ctx context.Context, profileID uuid.UUID) error

    // Query profile
    GetComponents(ctx context.Context, profileID uuid.UUID, filter ComponentFilter) ([]Component, error)
    GetModules(ctx context.Context, profileID uuid.UUID) ([]Module, error)
    SearchProfile(ctx context.Context, profileID uuid.UUID, query string) ([]ProfileSearchResult, error)
}

type ApplicationProfile struct {
    ID              uuid.UUID
    Name            string
    Description     string
    TechStack       TechStack
    StructureTree   *FileTree
    Components      []Component
    Modules         []Module
    Icons           []Icon
    APIEndpoints    []APIEndpoint
    DatabaseSchema  *DatabaseSchema
    Conventions     Conventions
}

type Component struct {
    Name        string                 `json:"name"`
    Path        string                 `json:"path"`
    Description string                 `json:"description"`
    Props       []PropDefinition       `json:"props"`
    Events      []EventDefinition      `json:"events"`
    Slots       []SlotDefinition       `json:"slots"`
    Usage       []UsageExample         `json:"usage"`
}

type Module struct {
    Name        string                 `json:"name"`
    Path        string                 `json:"path"`
    Description string                 `json:"description"`
    Exports     []ExportDefinition     `json:"exports"`
    Dependencies []string              `json:"dependencies"`
}
```

---

## PART 9: Implementation Priority

### Phase 1: Memory Foundation (Week 1-2)
1. Create memory tables and migrations
2. Implement MemoryService with CRUD
3. Add semantic search with embeddings
4. Build memory extraction from conversations
5. API endpoints for memory management

### Phase 2: Context System (Week 2-3)
1. Create context tracking tables
2. Implement ContextService
3. Build tree search tool for agents
4. Implement context loading rules
5. Context window tracking

### Phase 3: Output Styles & Blocks (Week 3)
1. Create output styles tables
2. Implement block mapper
3. Pre-seed output styles
4. API for style preferences
5. Integration with chat handler

### Phase 4: Document & File Upload System (Week 3-4)
1. Create uploaded_documents and document_chunks tables
2. Implement DocumentProcessor service
3. PDF/DOCX/Markdown text extraction
4. Document chunking for retrieval
5. Semantic search on documents
6. File upload API endpoints

### Phase 5: Chat History & Voice Note Integration (Week 4)
1. Add conversation summary tables
2. Implement chat history as context source
3. Update voice notes with context links
4. Implement project context loading
5. Implement node context loading
6. Context injection in conversations

### Phase 6: Self-Learning (Week 5)
1. Create learning tables
2. Implement LearningService
3. Build learning triggers
4. Pattern detection
5. Personalization profile

### Phase 7: Orchestrator Enhancement (Week 5-6)
1. New orchestrator prompt template
2. Context-aware prompt composition
3. Tool registration for context tools
4. Testing and refinement

### Phase 8: Tree Visualization & API (Week 6)
1. Context tree API endpoints
2. Tree statistics
3. Profile generation
4. Context profile management UI integration

---

## PART 10: API Endpoint Summary

### Memory System
```
GET    /api/memories
POST   /api/memories
GET    /api/memories/:id
PUT    /api/memories/:id
DELETE /api/memories/:id
POST   /api/memories/:id/pin
POST   /api/memories/search
POST   /api/memories/relevant
GET    /api/memories/project/:projectId
GET    /api/memories/node/:nodeId
GET    /api/user-facts
PUT    /api/user-facts/:key
DELETE /api/user-facts/:key
GET    /api/memories/stats
```

### Context System
```
GET    /api/context-tree
GET    /api/context-tree/project/:projectId
GET    /api/context-tree/node/:nodeId
GET    /api/context-tree/stats
POST   /api/context/search
POST   /api/context/load
GET    /api/context/rules
POST   /api/context/rules
PUT    /api/context/rules/:id
DELETE /api/context/rules/:id
```

### Documents & File Uploads
```
POST   /api/documents/upload              # Upload a document (PDF, markdown, etc.)
GET    /api/documents                     # List user's documents
GET    /api/documents/:id                 # Get document details
GET    /api/documents/:id/content         # Get extracted text content
DELETE /api/documents/:id                 # Delete document
POST   /api/documents/search              # Semantic search documents
```

### Context Profiles
```
GET    /api/context-profiles                          # List user's context profiles
POST   /api/context-profiles                          # Create context profile
GET    /api/context-profiles/:id                      # Get profile details
PUT    /api/context-profiles/:id                      # Update profile
DELETE /api/context-profiles/:id                      # Delete profile
POST   /api/context-profiles/:id/documents            # Add document to profile
GET    /api/context-profiles/:id/documents            # List profile documents
DELETE /api/context-profiles/:id/documents/:docId     # Remove document from profile
GET    /api/context-profiles/:id/items                # List all items in profile
```

### Chat History as Context
```
GET    /api/conversations/:id/summary     # Get conversation summary
POST   /api/conversations/:id/summarize   # Generate/update summary
GET    /api/conversations/search          # Search conversations semantically
GET    /api/conversations/context         # Get conversations as context source
```

### Output Styles
```
GET    /api/output-styles
GET    /api/output-styles/:id
GET    /api/user/output-preferences
PUT    /api/user/output-preferences
POST   /api/blocks/convert
```

### Learning System
```
POST   /api/feedback
GET    /api/learning/patterns
GET    /api/learning/profile
PUT    /api/learning/profile
```

### Application Profiles
```
GET    /api/app-profiles
POST   /api/app-profiles
GET    /api/app-profiles/:id
PUT    /api/app-profiles/:id
POST   /api/app-profiles/:id/sync
GET    /api/app-profiles/:id/components
GET    /api/app-profiles/:id/modules
POST   /api/app-profiles/:id/search
```

---

## PART 11: Environment Variables

```env
# Embeddings
OPENAI_API_KEY=sk-xxx                         # For ada-002 embeddings
EMBEDDING_MODEL=text-embedding-ada-002
EMBEDDING_DIMENSIONS=1536

# Memory Settings
MEMORY_AUTO_EXTRACT=true
MEMORY_MAX_PER_CONVERSATION=10
MEMORY_IMPORTANCE_THRESHOLD=0.5

# Context Settings
CONTEXT_MAX_TOKENS=8000
CONTEXT_AUTO_LOAD=true
CONTEXT_TREE_SEARCH_LIMIT=20

# Learning Settings
LEARNING_ENABLED=true
LEARNING_AUTO_PATTERN_DETECTION=true
PATTERN_CONFIDENCE_THRESHOLD=0.7
```

---

## PART 12: Testing Checklist

### Memory System
- [ ] Memories can be created, read, updated, deleted
- [ ] Semantic search returns relevant memories
- [ ] Memories are properly linked to projects/nodes
- [ ] Memory extraction from conversations works
- [ ] User facts are stored and retrieved correctly

### Context System
- [ ] Context tree API returns proper structure
- [ ] Tree search finds relevant items
- [ ] Context loading based on project/node works
- [ ] Context window tracking is accurate
- [ ] Loading rules are applied correctly

### Documents & File Uploads
- [ ] PDF upload and text extraction works
- [ ] Markdown file upload works
- [ ] DOCX file upload works
- [ ] Document chunking produces correct chunks
- [ ] Semantic search on documents returns relevant results
- [ ] Documents link properly to context profiles
- [ ] Document deletion cascades chunks properly

### Context Profiles
- [ ] Context profiles can be created, updated, deleted
- [ ] Documents can be added/removed from profiles
- [ ] Profile items include all types (documents, memories, artifacts)
- [ ] Profile hierarchy reflects Node → Project → Profile → Documents

### Chat History as Context
- [ ] Conversation summaries are generated correctly
- [ ] Key points and decisions are extracted
- [ ] Semantic search on conversations works
- [ ] Chat history loads into agent context properly
- [ ] Summaries update as conversations continue

### Output Styles
- [ ] All pre-seeded styles work
- [ ] User preferences are saved and applied
- [ ] Block conversion produces correct output
- [ ] Style changes affect AI output appropriately

### Self-Learning
- [ ] Feedback is recorded and processed
- [ ] Patterns are detected over time
- [ ] Corrections result in learnings
- [ ] Personalization profile updates correctly

### Orchestrator
- [ ] Context is properly injected into prompts
- [ ] Tools work (tree_search, load_context)
- [ ] Delegation to specialists works
- [ ] User profile affects communication style

---

## PART 13: Questions for Clarification

1. **Embedding Provider**: Confirm OpenAI ada-002 or should we support alternatives (Cohere, local)?

2. **Memory Limits**: Max memories per user? Auto-archival policy?

3. **Context Token Budget**: Default token allocation for memories vs contexts vs artifacts?

4. **Learning Validation**: Should users confirm learned facts/patterns?

5. **Tree Visualization**: Frontend implementation or just API?

6. **Application Profiler**: Priority level? Should it auto-sync with git?

---

## Document Version

**Version:** 2.0.0
**Created:** December 31, 2025
**Author:** Roberto (Product Vision), Claude (Technical Specification)

---

**This document captures Roberto's complete vision for the Memory, Context, and Intelligence System. Every requirement has been preserved and translated into technical specifications. Implementation should follow the phased approach while maintaining all stated objectives.**
