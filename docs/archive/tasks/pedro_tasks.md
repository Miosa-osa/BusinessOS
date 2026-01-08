# Pedro's Implementation Plan - Memory, Context & Intelligence System

**Created:** December 31, 2025
**Branch:** `pedro-dev`
**Status:** Planning Complete
**Total Tasks:** 12

---

## Overview

Este documento detalha as tarefas exclusivas do Pedro para implementar o sistema de Memory, Context & Intelligence do BusinessOS.

### Objetivo

Transformar o BusinessOS em um sistema verdadeiramente inteligente e personalizado que:
- Armazena e recupera memórias episódicas
- Gerencia contexto hierárquico (Node → Project → Profile → Documents)
- Aprende continuamente sobre o usuário
- Produz outputs personalizados (não genéricos)

---

## Task Summary

| ID | Descrição | Tipo | Status |
|----|-----------|------|--------|
| P1.4 | Memory API Endpoints | Handler | Pending |
| P2.3 | Context System Tables | Migration | Pending |
| P2.5 | ContextService | Service | Pending |
| P3.3 | Seed Output Styles | Migration | Pending |
| P4.2 | Chat History Context | Migration | Pending |
| P4.3 | Voice Notes Context | Migration | Pending |
| P4.4 | Documents Tables | Migration | Pending |
| P4.5 | Context Profile Items | Migration | Pending |
| P4.8 | ProjectContextService | Service | Pending |
| P5.2 | Context Tree API | Handler | Pending |
| P6.2 | Learning System Tables | Migration | Pending |
| P8.2 | Application Profiles | Migration | Pending |

---

## Detailed Tasks

### P1.4: Memory API Endpoints

**File:** `internal/handlers/memory.go`

**Endpoints to create:**

```go
// Memory CRUD
GET    /api/memories                    // List memories (filters: type, project, node)
POST   /api/memories                    // Create memory
GET    /api/memories/:id                // Get specific memory
PUT    /api/memories/:id                // Update memory
DELETE /api/memories/:id                // Delete memory
POST   /api/memories/:id/pin            // Pin/unpin memory

// Memory Search & Retrieval
POST   /api/memories/search             // Semantic search
POST   /api/memories/relevant           // Get relevant to context
GET    /api/memories/project/:projectId // Memories for project
GET    /api/memories/node/:nodeId       // Memories for node

// User Facts
GET    /api/user-facts                  // All user facts
PUT    /api/user-facts/:key             // Update fact
DELETE /api/user-facts/:key             // Delete fact

// Analytics
GET    /api/memories/stats              // Statistics
```

**Dependencies:** Migration 016 (memories tables) must exist first.

---

### P2.3: Context System Migration

**File:** `internal/database/migrations/017_context_system.sql`

```sql
-- ================================================
-- CONTEXT PROFILES
-- ================================================
CREATE TABLE context_profiles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(255) NOT NULL,

    -- What this profile describes
    entity_type VARCHAR(50) NOT NULL,         -- 'project', 'node', 'application'
    entity_id UUID NOT NULL,

    -- Profile content
    name VARCHAR(255) NOT NULL,
    description TEXT,

    -- Tree structure (JSON representation)
    context_tree JSONB NOT NULL DEFAULT '{}',

    -- Summary for quick loading
    summary TEXT,
    key_facts TEXT[],

    -- Statistics
    total_contexts INTEGER DEFAULT 0,
    total_memories INTEGER DEFAULT 0,
    total_artifacts INTEGER DEFAULT 0,
    total_tasks INTEGER DEFAULT 0,

    -- Embedding for semantic search
    embedding vector(1536),

    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),

    UNIQUE(user_id, entity_type, entity_id)
);

CREATE INDEX idx_context_profiles_user ON context_profiles(user_id);
CREATE INDEX idx_context_profiles_entity ON context_profiles(entity_type, entity_id);

-- ================================================
-- CONTEXT LOADING RULES
-- ================================================
CREATE TABLE context_loading_rules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(255) NOT NULL,

    -- Rule definition
    name VARCHAR(255) NOT NULL,
    description TEXT,

    -- When to apply
    trigger_type VARCHAR(50) NOT NULL,        -- 'project_select', 'node_select', 'agent_type', 'keyword', 'always'
    trigger_value VARCHAR(255),

    -- What to load
    load_memories BOOLEAN DEFAULT TRUE,
    memory_types TEXT[],
    memory_limit INTEGER DEFAULT 10,

    load_contexts BOOLEAN DEFAULT TRUE,
    context_categories TEXT[],
    context_limit INTEGER DEFAULT 5,

    load_artifacts BOOLEAN DEFAULT FALSE,
    artifact_types TEXT[],
    artifact_limit INTEGER DEFAULT 3,

    load_recent_conversations BOOLEAN DEFAULT TRUE,
    conversation_limit INTEGER DEFAULT 3,

    -- Priority
    priority INTEGER DEFAULT 0,

    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_context_rules_user ON context_loading_rules(user_id);
CREATE INDEX idx_context_rules_trigger ON context_loading_rules(trigger_type, trigger_value);

-- ================================================
-- AGENT CONTEXT SESSIONS
-- ================================================
CREATE TABLE agent_context_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(255) NOT NULL,
    conversation_id UUID NOT NULL,

    -- Agent info
    agent_type VARCHAR(100) NOT NULL,
    agent_id UUID,

    -- Context window tracking
    max_context_tokens INTEGER NOT NULL,
    used_context_tokens INTEGER DEFAULT 0,
    available_tokens INTEGER,

    -- What's loaded
    loaded_memories UUID[] DEFAULT '{}',
    loaded_contexts UUID[] DEFAULT '{}',
    loaded_artifacts UUID[] DEFAULT '{}',

    -- System prompt
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

-- ================================================
-- CONTEXT RETRIEVAL LOG
-- ================================================
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

---

### P2.5: ContextService Implementation

**File:** `internal/services/context.go`

```go
package services

import (
    "context"
    "github.com/google/uuid"
    "github.com/jackc/pgx/v5/pgxpool"
)

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
    TreePath       []string  `json:"tree_path"`
    TokenEstimate  int       `json:"token_estimate"`
}

type contextServiceImpl struct {
    pool *pgxpool.Pool
}

func NewContextService(pool *pgxpool.Pool) ContextService {
    return &contextServiceImpl{pool: pool}
}

// Implementation methods...
```

---

### P3.3: Seed Output Styles

**File:** `internal/database/migrations/018_output_styles.sql`

```sql
-- Seed 8 pre-defined output styles
INSERT INTO output_styles (name, display_name, description, icon, style_type, use_headers, use_bullets, use_numbered_lists, use_paragraphs, use_code_blocks, use_tables, verbosity, tone, style_instructions, block_mapping, is_system, sort_order)
VALUES
    -- 1. Conversational
    ('conversational', 'Conversational', 'Natural, flowing conversation', 'message-circle', 'prose',
     FALSE, FALSE, FALSE, TRUE, FALSE, FALSE,
     'balanced', 'friendly',
     'Respond naturally in flowing paragraphs. Be warm and conversational. Avoid bullet points.',
     '{"paragraph": "text", "emphasis": "text"}',
     TRUE, 1),

    -- 2. Professional
    ('professional', 'Professional', 'Clear, structured business communication', 'briefcase', 'structured',
     TRUE, TRUE, FALSE, TRUE, FALSE, FALSE,
     'balanced', 'professional',
     'Use clear structure with headers and bullet points. Be professional and actionable.',
     '{"h2": "heading", "bullet": "bullet_list", "paragraph": "text"}',
     TRUE, 2),

    -- 3. Technical
    ('technical', 'Technical', 'Precise technical documentation with code', 'code', 'code',
     TRUE, TRUE, FALSE, TRUE, TRUE, TRUE,
     'detailed', 'technical',
     'Include code examples. Use precise terminology. Structure with headers and tables.',
     '{"h2": "heading", "code": "code", "table": "table", "bullet": "bullet_list"}',
     TRUE, 3),

    -- 4. Executive Summary
    ('executive', 'Executive Summary', 'Brief, high-level summaries', 'zap', 'structured',
     FALSE, TRUE, FALSE, TRUE, FALSE, FALSE,
     'concise', 'formal',
     'Lead with key takeaway. Use 3-5 bullets max. Under 200 words. Focus on business impact.',
     '{"bullet": "bullet_list", "paragraph": "text"}',
     TRUE, 4),

    -- 5. Detailed Analysis
    ('detailed', 'Detailed Analysis', 'Comprehensive, in-depth analysis', 'file-text', 'structured',
     TRUE, TRUE, TRUE, TRUE, FALSE, TRUE,
     'comprehensive', 'professional',
     'Provide thorough analysis with sections. Include considerations and trade-offs.',
     '{"h2": "heading", "h3": "subheading", "numbered": "numbered_list", "bullet": "bullet_list", "table": "table"}',
     TRUE, 5),

    -- 6. Creative
    ('creative', 'Creative', 'Engaging, narrative style', 'sparkles', 'prose',
     FALSE, FALSE, FALSE, TRUE, FALSE, FALSE,
     'detailed', 'casual',
     'Use narrative techniques, metaphors, vivid descriptions. Make content memorable.',
     '{"paragraph": "text", "blockquote": "quote"}',
     TRUE, 6),

    -- 7. Tutorial
    ('tutorial', 'Step-by-Step', 'Clear tutorial with numbered steps', 'list-ordered', 'structured',
     TRUE, FALSE, TRUE, TRUE, TRUE, FALSE,
     'detailed', 'friendly',
     'Use numbered steps. Include code examples with explanations. Add tips for important points.',
     '{"h2": "heading", "numbered": "numbered_list", "code": "code", "note": "callout"}',
     TRUE, 7),

    -- 8. Q&A
    ('qa', 'Q&A', 'Direct question-and-answer format', 'help-circle', 'mixed',
     FALSE, FALSE, FALSE, TRUE, FALSE, FALSE,
     'concise', 'friendly',
     'Answer directly and concisely. Start with direct answer, then brief context if needed.',
     '{"paragraph": "text"}',
     TRUE, 8)

ON CONFLICT (name) DO UPDATE SET
    display_name = EXCLUDED.display_name,
    description = EXCLUDED.description,
    style_instructions = EXCLUDED.style_instructions,
    updated_at = NOW();
```

---

### P4.2: Chat History Context Schema

**File:** `internal/database/migrations/020_context_integration.sql` (Part 1)

```sql
-- ================================================
-- CHAT HISTORY AS CONTEXT
-- ================================================

-- Add context columns to conversations
ALTER TABLE conversations ADD COLUMN IF NOT EXISTS is_context_source BOOLEAN DEFAULT TRUE;
ALTER TABLE conversations ADD COLUMN IF NOT EXISTS extracted_memories UUID[] DEFAULT '{}';
ALTER TABLE conversations ADD COLUMN IF NOT EXISTS summary TEXT;
ALTER TABLE conversations ADD COLUMN IF NOT EXISTS key_topics TEXT[] DEFAULT '{}';
ALTER TABLE conversations ADD COLUMN IF NOT EXISTS embedding vector(1536);

CREATE INDEX IF NOT EXISTS idx_conversations_context ON conversations(user_id, is_context_source);
CREATE INDEX IF NOT EXISTS idx_conversations_embedding ON conversations USING ivfflat (embedding vector_cosine_ops);

-- Conversation summaries for efficient loading
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

    -- Embedding
    embedding vector(1536),

    -- Timestamps
    summarized_at TIMESTAMPTZ DEFAULT NOW(),
    message_count INTEGER,
    time_range TSTZRANGE
);

CREATE INDEX idx_conv_summaries_conv ON conversation_summaries(conversation_id);
CREATE INDEX idx_conv_summaries_user ON conversation_summaries(user_id);
```

---

### P4.3: Voice Notes Context Schema

**File:** `internal/database/migrations/020_context_integration.sql` (Part 2)

```sql
-- ================================================
-- VOICE NOTES CONTEXT
-- ================================================

ALTER TABLE voice_notes ADD COLUMN IF NOT EXISTS project_id UUID REFERENCES projects(id) ON DELETE SET NULL;
ALTER TABLE voice_notes ADD COLUMN IF NOT EXISTS node_id UUID REFERENCES nodes(id) ON DELETE SET NULL;
ALTER TABLE voice_notes ADD COLUMN IF NOT EXISTS is_context_source BOOLEAN DEFAULT TRUE;
ALTER TABLE voice_notes ADD COLUMN IF NOT EXISTS extracted_memories UUID[] DEFAULT '{}';
ALTER TABLE voice_notes ADD COLUMN IF NOT EXISTS embedding vector(1536);

CREATE INDEX IF NOT EXISTS idx_voice_notes_project ON voice_notes(project_id);
CREATE INDEX IF NOT EXISTS idx_voice_notes_node ON voice_notes(node_id);
CREATE INDEX IF NOT EXISTS idx_voice_notes_embedding ON voice_notes USING ivfflat (embedding vector_cosine_ops);
```

---

### P4.4: Documents Tables

**File:** `internal/database/migrations/019_documents.sql`

```sql
-- ================================================
-- UPLOADED DOCUMENTS
-- ================================================
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
    storage_path VARCHAR(1000) NOT NULL,
    storage_provider VARCHAR(50) DEFAULT 'local',

    -- Extracted content
    extracted_text TEXT,
    page_count INTEGER,
    word_count INTEGER,

    -- Context links
    context_profile_id UUID REFERENCES context_profiles(id) ON DELETE SET NULL,
    project_id UUID REFERENCES projects(id) ON DELETE SET NULL,
    node_id UUID REFERENCES nodes(id) ON DELETE SET NULL,

    -- Categorization
    document_type VARCHAR(100),              -- 'sop', 'framework', 'template', 'reference', 'report'
    category VARCHAR(100),
    tags TEXT[] DEFAULT '{}',

    -- Semantic search
    embedding vector(1536),

    -- Processing
    processing_status VARCHAR(50) DEFAULT 'pending',
    processed_at TIMESTAMPTZ,

    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_uploaded_docs_user ON uploaded_documents(user_id);
CREATE INDEX idx_uploaded_docs_profile ON uploaded_documents(context_profile_id);
CREATE INDEX idx_uploaded_docs_project ON uploaded_documents(project_id);
CREATE INDEX idx_uploaded_docs_type ON uploaded_documents(document_type);
CREATE INDEX idx_uploaded_docs_embedding ON uploaded_documents USING ivfflat (embedding vector_cosine_ops);

-- ================================================
-- DOCUMENT CHUNKS
-- ================================================
CREATE TABLE document_chunks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    document_id UUID NOT NULL REFERENCES uploaded_documents(id) ON DELETE CASCADE,

    -- Chunk info
    chunk_index INTEGER NOT NULL,
    content TEXT NOT NULL,
    token_count INTEGER,

    -- Position
    page_number INTEGER,
    section_title VARCHAR(255),

    -- Embedding
    embedding vector(1536),

    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_doc_chunks_document ON document_chunks(document_id);
CREATE INDEX idx_doc_chunks_embedding ON document_chunks USING ivfflat (embedding vector_cosine_ops);
```

---

### P4.5: Context Profile Items

**File:** `internal/database/migrations/020_context_integration.sql` (Part 3)

```sql
-- ================================================
-- CONTEXT PROFILE ITEMS
-- ================================================

-- Update context_profiles
ALTER TABLE context_profiles ADD COLUMN IF NOT EXISTS document_types TEXT[] DEFAULT '{}';
ALTER TABLE context_profiles ADD COLUMN IF NOT EXISTS total_documents INTEGER DEFAULT 0;
ALTER TABLE context_profiles ADD COLUMN IF NOT EXISTS total_file_size_bytes BIGINT DEFAULT 0;

-- Links various content types to profiles
CREATE TABLE context_profile_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    context_profile_id UUID NOT NULL REFERENCES context_profiles(id) ON DELETE CASCADE,
    user_id VARCHAR(255) NOT NULL,

    -- Item reference
    item_type VARCHAR(50) NOT NULL,           -- 'document', 'artifact', 'memory', 'conversation', 'voice_note'
    item_id UUID NOT NULL,

    -- Display
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

---

### P4.8: ProjectContextService

**File:** `internal/services/project_context.go`

```go
package services

import (
    "context"
    "github.com/google/uuid"
    "github.com/jackc/pgx/v5/pgxpool"
)

type ProjectContextService struct {
    pool           *pgxpool.Pool
    memoryService  MemoryService
    contextService ContextService
}

func NewProjectContextService(pool *pgxpool.Pool, ms MemoryService, cs ContextService) *ProjectContextService {
    return &ProjectContextService{
        pool:           pool,
        memoryService:  ms,
        contextService: cs,
    }
}

type ProjectContext struct {
    Project       *Project
    Profile       *ContextProfile
    Memories      []Memory
    Contexts      []Context
    VoiceNotes    []VoiceNote
    Conversations []ConversationSummary
}

type NodeContext struct {
    Node          *Node
    Ancestors     []*Node
    Profile       *ContextProfile
    Memories      []Memory
    ParentContext *ProjectContext
}

// LoadProjectContext loads all relevant context when a project is selected
func (s *ProjectContextService) LoadProjectContext(ctx context.Context, userID string, projectID uuid.UUID) (*ProjectContext, error) {
    // 1. Get project details
    // 2. Get project's context profile
    // 3. Load memories associated with project
    // 4. Load KB contexts linked to project
    // 5. Get recent voice notes for project
    // 6. Get recent conversations in project
    return nil, nil // TODO: implement
}

// LoadNodeContext loads context when a specific node is selected
func (s *ProjectContextService) LoadNodeContext(ctx context.Context, userID string, nodeID uuid.UUID) (*NodeContext, error) {
    // 1. Get node details and its ancestors
    // 2. Get node's context profile
    // 3. Load memories for this specific node
    // 4. Inherit relevant parent context
    return nil, nil // TODO: implement
}
```

---

### P5.2: Context Tree API

**File:** `internal/handlers/context.go` (partial)

```go
// Context Tree Endpoints

// GET /api/context-tree
func (h *Handlers) GetContextTree(w http.ResponseWriter, r *http.Request) {
    // Return full user context tree
}

// GET /api/context-tree/project/:projectId
func (h *Handlers) GetProjectContextTree(w http.ResponseWriter, r *http.Request) {
    // Return tree for specific project
}

// GET /api/context-tree/node/:nodeId
func (h *Handlers) GetNodeContextTree(w http.ResponseWriter, r *http.Request) {
    // Return tree for specific node
}

// GET /api/context-tree/stats
func (h *Handlers) GetContextTreeStats(w http.ResponseWriter, r *http.Request) {
    // Return tree statistics
}

// Response format
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
    ItemCount   int                `json:"item_count"`
    TokenCount  int                `json:"token_count"`
    Children    []*ContextTreeNode `json:"children,omitempty"`
}

type TreeStatistics struct {
    TotalProjects   int            `json:"total_projects"`
    TotalNodes      int            `json:"total_nodes"`
    TotalMemories   int            `json:"total_memories"`
    TotalContexts   int            `json:"total_contexts"`
    TotalArtifacts  int            `json:"total_artifacts"`
    TotalVoiceNotes int            `json:"total_voice_notes"`
    TotalDocuments  int            `json:"total_documents"`
    ByType          map[string]int `json:"by_type"`
    TotalTokens     int            `json:"total_tokens"`
}
```

---

### P6.2: Learning System Tables

**File:** `internal/database/migrations/021_learning_system.sql`

```sql
-- ================================================
-- LEARNING EVENTS
-- ================================================
CREATE TABLE learning_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(255) NOT NULL,

    -- What was learned
    learning_type VARCHAR(50) NOT NULL,       -- 'correction', 'preference', 'pattern', 'feedback', 'behavior'
    learning_content TEXT NOT NULL,

    -- Source
    source_type VARCHAR(50) NOT NULL,         -- 'explicit_feedback', 'implicit_behavior', 'correction', 'conversation'
    source_id UUID,
    source_context TEXT,

    -- Confidence
    confidence_score DECIMAL(3,2) DEFAULT 0.5,
    times_applied INTEGER DEFAULT 0,
    last_applied_at TIMESTAMPTZ,

    -- Results
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

-- ================================================
-- USER BEHAVIOR PATTERNS
-- ================================================
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
    evidence_ids UUID[],

    -- Confidence
    confidence_score DECIMAL(3,2) DEFAULT 0.5,

    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),

    UNIQUE(user_id, pattern_type, pattern_key)
);

CREATE INDEX idx_behavior_patterns_user ON user_behavior_patterns(user_id);

-- ================================================
-- FEEDBACK LOG
-- ================================================
CREATE TABLE feedback_log (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(255) NOT NULL,

    -- Target
    target_type VARCHAR(50) NOT NULL,         -- 'message', 'artifact', 'memory', 'suggestion'
    target_id UUID NOT NULL,

    -- Feedback
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

---

### P8.2: Application Profiles Table

**File:** `internal/database/migrations/022_application_profiles.sql`

```sql
-- ================================================
-- APPLICATION PROFILES (for future IDE integration)
-- ================================================
CREATE TABLE application_profiles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(255) NOT NULL,

    -- Application Identity
    name VARCHAR(255) NOT NULL,
    description TEXT,
    app_type VARCHAR(100),                    -- 'web_app', 'mobile_app', 'api', 'library', 'platform'

    -- Tech Stack
    tech_stack JSONB DEFAULT '{}',            -- {"frontend": "svelte", "backend": "go", "database": "postgres"}

    -- Structure
    structure_tree JSONB NOT NULL DEFAULT '{}',

    -- Components Registry
    components JSONB DEFAULT '[]',            -- [{name, path, description, props}]

    -- Modules Registry
    modules JSONB DEFAULT '[]',               -- [{name, path, description, exports}]

    -- Icons/Assets
    icons JSONB DEFAULT '[]',

    -- API Endpoints
    api_endpoints JSONB DEFAULT '[]',

    -- Database Schema
    database_schema JSONB DEFAULT '{}',

    -- Conventions
    conventions JSONB DEFAULT '{}',

    -- Integration Points
    integration_points JSONB DEFAULT '[]',

    -- Embedding
    embedding vector(1536),

    -- Sync
    last_synced_at TIMESTAMPTZ,
    sync_source VARCHAR(255),

    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_app_profiles_user ON application_profiles(user_id);
CREATE INDEX idx_app_profiles_type ON application_profiles(app_type);
```

---

## Files to Create

```
desktop/backend-go/
├── internal/database/migrations/
│   ├── 017_context_system.sql        # P2.3
│   ├── 018_output_styles.sql         # P3.3 (seed only, table created by Nick)
│   ├── 019_documents.sql             # P4.4
│   ├── 020_context_integration.sql   # P4.2, P4.3, P4.5
│   ├── 021_learning_system.sql       # P6.2
│   └── 022_application_profiles.sql  # P8.2
├── internal/handlers/
│   ├── memory.go                     # P1.4
│   └── context.go                    # P5.2
└── internal/services/
    ├── context.go                    # P2.5
    └── project_context.go            # P4.8
```

---

## Execution Order

```
1. P2.3  → Migration 017 (context system tables)
2. P4.4  → Migration 019 (documents tables)
3. P4.2  → Migration 020 part 1 (chat history)
4. P4.3  → Migration 020 part 2 (voice notes)
5. P4.5  → Migration 020 part 3 (context profile items)
6. P6.2  → Migration 021 (learning system)
7. P8.2  → Migration 022 (application profiles)
8. P3.3  → Migration 018 seed (output styles)
9. P2.5  → Service: ContextService
10. P4.8 → Service: ProjectContextService
11. P1.4 → Handler: Memory endpoints
12. P5.2 → Handler: Context Tree endpoints
```

---

## Environment Variables

```env
# Embeddings (required for semantic search)
OPENAI_API_KEY=sk-xxx
EMBEDDING_MODEL=text-embedding-ada-002
EMBEDDING_DIMENSIONS=1536

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

## Testing Checklist

- [ ] Migration 017 runs without errors
- [ ] Migration 019 runs without errors
- [ ] Migration 020 runs without errors
- [ ] Migration 021 runs without errors
- [ ] Migration 022 runs without errors
- [ ] ContextService.GetContextTree returns valid tree
- [ ] ContextService.SearchTree finds relevant items
- [ ] ProjectContextService.LoadProjectContext loads all context types
- [ ] Memory endpoints work (CRUD + search)
- [ ] Context Tree API returns proper structure
- [ ] Semantic search with embeddings works

---

## Notes

- All tables use `vector(1536)` for OpenAI ada-002 embeddings
- pgvector extension must be enabled: `CREATE EXTENSION IF NOT EXISTS vector;`
- Indexes use `ivfflat` for approximate nearest neighbor search
- Token tracking is essential for context window management

---

**Document Version:** 1.0.0
**Last Updated:** December 31, 2025
