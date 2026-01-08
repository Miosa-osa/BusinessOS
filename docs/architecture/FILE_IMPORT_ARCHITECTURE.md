# File Import Architecture

## Overview

BusinessOS supports importing conversation history and data from external AI platforms. Unlike OAuth-based integrations, file imports are manual uploads that don't require API credentials or ongoing sync.

**Supported Providers:**
- ChatGPT (JSON export)
- Claude (JSON export)
- Perplexity (JSON export)
- Gemini (JSON export)
- Granola (Meeting notes JSON)

---

## Data Flow Architecture

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                           FILE IMPORT PIPELINE                              │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│   ┌─────────────┐     ┌─────────────┐     ┌─────────────┐     ┌──────────┐ │
│   │   Upload    │────▶│   Parser    │────▶│  Processor  │────▶│  Store   │ │
│   │   (.json)   │     │  (provider  │     │  (normalize │     │ (DB +    │ │
│   │             │     │   specific) │     │   + enrich) │     │  vector) │ │
│   └─────────────┘     └─────────────┘     └─────────────┘     └──────────┘ │
│                                                                     │       │
│                                                                     ▼       │
│   ┌─────────────────────────────────────────────────────────────────────┐   │
│   │                         KNOWLEDGE GRAPH                             │   │
│   │  ┌────────────┐  ┌────────────┐  ┌────────────┐  ┌────────────┐    │   │
│   │  │  Memories  │  │  Contexts  │  │   Facts    │  │  Entities  │    │   │
│   │  └────────────┘  └────────────┘  └────────────┘  └────────────┘    │   │
│   └─────────────────────────────────────────────────────────────────────┘   │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## Database Schema

### Core Import Tables

```sql
-- Import jobs tracking
CREATE TABLE file_imports (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    provider VARCHAR(50) NOT NULL, -- 'chatgpt', 'claude', 'perplexity', 'gemini', 'granola'
    filename VARCHAR(255) NOT NULL,
    file_size_bytes BIGINT NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending', -- 'pending', 'processing', 'completed', 'failed'

    -- Statistics
    total_conversations INT DEFAULT 0,
    total_messages INT DEFAULT 0,
    processed_conversations INT DEFAULT 0,
    processed_messages INT DEFAULT 0,

    -- Metadata
    import_options JSONB DEFAULT '{}', -- User-selected import options
    error_message TEXT,
    started_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),

    CONSTRAINT valid_provider CHECK (provider IN ('chatgpt', 'claude', 'perplexity', 'gemini', 'granola')),
    CONSTRAINT valid_status CHECK (status IN ('pending', 'processing', 'completed', 'failed'))
);

CREATE INDEX idx_file_imports_user ON file_imports(user_id);
CREATE INDEX idx_file_imports_status ON file_imports(status);

-- Imported conversations
CREATE TABLE imported_conversations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    import_id UUID NOT NULL REFERENCES file_imports(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,

    -- Source identification
    provider VARCHAR(50) NOT NULL,
    external_id VARCHAR(255), -- Original conversation ID from provider

    -- Content
    title VARCHAR(500),
    summary TEXT, -- AI-generated summary

    -- Timestamps (from source)
    started_at TIMESTAMPTZ,
    ended_at TIMESTAMPTZ,

    -- Classification
    topics TEXT[], -- Extracted topics
    entities TEXT[], -- Extracted entities (people, companies, etc.)
    sentiment VARCHAR(20), -- Overall conversation sentiment
    category VARCHAR(100), -- Business, personal, technical, etc.

    -- Integration links
    linked_project_id UUID REFERENCES projects(id),
    linked_client_id UUID REFERENCES clients(id),
    linked_context_id UUID REFERENCES contexts(id),

    -- Metadata
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),

    UNIQUE(import_id, external_id)
);

CREATE INDEX idx_imported_conversations_user ON imported_conversations(user_id);
CREATE INDEX idx_imported_conversations_provider ON imported_conversations(provider);
CREATE INDEX idx_imported_conversations_topics ON imported_conversations USING GIN(topics);

-- Imported messages
CREATE TABLE imported_messages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    conversation_id UUID NOT NULL REFERENCES imported_conversations(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,

    -- Message content
    role VARCHAR(20) NOT NULL, -- 'user', 'assistant', 'system'
    content TEXT NOT NULL,

    -- Ordering
    sequence_number INT NOT NULL,

    -- Timestamps
    created_at_source TIMESTAMPTZ, -- Original timestamp from provider

    -- Extracted data
    code_blocks TEXT[], -- Extracted code snippets
    urls TEXT[], -- Extracted URLs
    entities JSONB DEFAULT '[]', -- [{type: 'person', value: 'John'}]

    -- Embeddings for semantic search
    embedding vector(1536), -- OpenAI embedding

    -- Metadata
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_imported_messages_conversation ON imported_messages(conversation_id);
CREATE INDEX idx_imported_messages_role ON imported_messages(role);
CREATE INDEX idx_imported_messages_embedding ON imported_messages USING ivfflat (embedding vector_cosine_ops);

-- Knowledge extraction (facts, insights, decisions)
CREATE TABLE imported_knowledge (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    conversation_id UUID REFERENCES imported_conversations(id) ON DELETE SET NULL,
    message_id UUID REFERENCES imported_messages(id) ON DELETE SET NULL,

    -- Classification
    knowledge_type VARCHAR(50) NOT NULL, -- 'fact', 'decision', 'insight', 'preference', 'goal', 'task'

    -- Content
    title VARCHAR(500),
    content TEXT NOT NULL,
    confidence FLOAT DEFAULT 1.0, -- AI confidence score

    -- Categorization
    category VARCHAR(100),
    tags TEXT[],

    -- Relationships
    related_entities JSONB DEFAULT '[]', -- [{type: 'person', value: 'John', id: 'uuid'}]

    -- Integration
    linked_memory_id UUID REFERENCES memories(id),
    linked_context_id UUID REFERENCES contexts(id),

    -- Embeddings
    embedding vector(1536),

    -- Metadata
    source_provider VARCHAR(50),
    extracted_at TIMESTAMPTZ DEFAULT NOW(),
    verified BOOLEAN DEFAULT FALSE, -- User verified this is accurate

    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_imported_knowledge_user ON imported_knowledge(user_id);
CREATE INDEX idx_imported_knowledge_type ON imported_knowledge(knowledge_type);
CREATE INDEX idx_imported_knowledge_tags ON imported_knowledge USING GIN(tags);
CREATE INDEX idx_imported_knowledge_embedding ON imported_knowledge USING ivfflat (embedding vector_cosine_ops);
```

---

## Provider-Specific Parsers

### ChatGPT Export Format

**File Structure:**
```json
{
  "title": "conversations.json",
  "conversations": [
    {
      "id": "abc123",
      "title": "Python debugging help",
      "create_time": 1699900000,
      "update_time": 1699900500,
      "mapping": {
        "message-id-1": {
          "id": "message-id-1",
          "message": {
            "author": {"role": "user"},
            "content": {"parts": ["How do I fix this error?"]},
            "create_time": 1699900000
          }
        },
        "message-id-2": {
          "id": "message-id-2",
          "parent": "message-id-1",
          "message": {
            "author": {"role": "assistant"},
            "content": {"parts": ["Here's how to fix it..."]},
            "create_time": 1699900100
          }
        }
      }
    }
  ]
}
```

**Parser Implementation:**
```go
type ChatGPTParser struct{}

func (p *ChatGPTParser) Parse(data []byte) (*ParsedImport, error) {
    var export ChatGPTExport
    if err := json.Unmarshal(data, &export); err != nil {
        return nil, fmt.Errorf("invalid ChatGPT export format: %w", err)
    }

    result := &ParsedImport{
        Provider: "chatgpt",
        Conversations: make([]ParsedConversation, 0, len(export.Conversations)),
    }

    for _, conv := range export.Conversations {
        parsed := ParsedConversation{
            ExternalID: conv.ID,
            Title:      conv.Title,
            StartedAt:  time.Unix(int64(conv.CreateTime), 0),
            Messages:   p.extractMessages(conv.Mapping),
        }
        result.Conversations = append(result.Conversations, parsed)
    }

    return result, nil
}

func (p *ChatGPTParser) extractMessages(mapping map[string]ChatGPTMessage) []ParsedMessage {
    // Build message tree and flatten in order
    // Handle branching conversations (take main thread)
}
```

### Claude Export Format

**File Structure:**
```json
{
  "version": "1.0",
  "conversations": [
    {
      "uuid": "conv-uuid-123",
      "name": "API design discussion",
      "created_at": "2024-01-15T10:30:00Z",
      "updated_at": "2024-01-15T11:45:00Z",
      "chat_messages": [
        {
          "uuid": "msg-uuid-1",
          "sender": "human",
          "text": "Help me design an API",
          "created_at": "2024-01-15T10:30:00Z"
        },
        {
          "uuid": "msg-uuid-2",
          "sender": "assistant",
          "text": "I'd recommend RESTful design...",
          "created_at": "2024-01-15T10:31:00Z"
        }
      ]
    }
  ]
}
```

**Parser Implementation:**
```go
type ClaudeParser struct{}

func (p *ClaudeParser) Parse(data []byte) (*ParsedImport, error) {
    var export ClaudeExport
    if err := json.Unmarshal(data, &export); err != nil {
        return nil, fmt.Errorf("invalid Claude export format: %w", err)
    }

    result := &ParsedImport{
        Provider: "claude",
        Conversations: make([]ParsedConversation, 0, len(export.Conversations)),
    }

    for _, conv := range export.Conversations {
        messages := make([]ParsedMessage, 0, len(conv.ChatMessages))
        for i, msg := range conv.ChatMessages {
            role := "user"
            if msg.Sender == "assistant" {
                role = "assistant"
            }
            messages = append(messages, ParsedMessage{
                Role:           role,
                Content:        msg.Text,
                SequenceNumber: i + 1,
                CreatedAt:      msg.CreatedAt,
            })
        }

        result.Conversations = append(result.Conversations, ParsedConversation{
            ExternalID: conv.UUID,
            Title:      conv.Name,
            StartedAt:  conv.CreatedAt,
            EndedAt:    conv.UpdatedAt,
            Messages:   messages,
        })
    }

    return result, nil
}
```

### Perplexity Export Format

**File Structure:**
```json
{
  "threads": [
    {
      "id": "thread-123",
      "title": "Research on quantum computing",
      "created_at": "2024-01-10T09:00:00Z",
      "messages": [
        {
          "role": "user",
          "content": "What are the latest advances in quantum computing?",
          "timestamp": "2024-01-10T09:00:00Z"
        },
        {
          "role": "assistant",
          "content": "Based on recent research...",
          "citations": [
            {"url": "https://example.com/paper1", "title": "Quantum Advances 2024"}
          ],
          "timestamp": "2024-01-10T09:00:30Z"
        }
      ]
    }
  ]
}
```

**Parser Implementation:**
```go
type PerplexityParser struct{}

func (p *PerplexityParser) Parse(data []byte) (*ParsedImport, error) {
    var export PerplexityExport
    // Similar pattern, also extract citations as metadata
    // Citations are valuable for knowledge graph
}
```

### Granola Meeting Notes Format

**File Structure:**
```json
{
  "meetings": [
    {
      "id": "meeting-123",
      "title": "Weekly Standup",
      "date": "2024-01-15",
      "start_time": "09:00",
      "end_time": "09:30",
      "participants": ["John", "Jane", "Bob"],
      "transcript": [
        {"speaker": "John", "text": "Let's start with updates..."},
        {"speaker": "Jane", "text": "I completed the API work..."}
      ],
      "summary": "Team discussed sprint progress...",
      "action_items": [
        {"assignee": "Bob", "task": "Review PR #123"}
      ]
    }
  ]
}
```

**Parser Implementation:**
```go
type GranolaParser struct{}

func (p *GranolaParser) Parse(data []byte) (*ParsedImport, error) {
    // Parse meeting format
    // Extract action items as potential tasks
    // Link to calendar events if possible
    // Create knowledge entries for decisions made
}
```

---

## Common Data Models

```go
// Normalized import structure (all providers convert to this)
type ParsedImport struct {
    Provider      string
    Conversations []ParsedConversation
    Metadata      map[string]interface{}
}

type ParsedConversation struct {
    ExternalID  string
    Title       string
    StartedAt   time.Time
    EndedAt     time.Time
    Messages    []ParsedMessage
    Metadata    map[string]interface{}
}

type ParsedMessage struct {
    Role           string // "user", "assistant", "system"
    Content        string
    SequenceNumber int
    CreatedAt      time.Time
    CodeBlocks     []string
    URLs           []string
    Citations      []Citation
    Metadata       map[string]interface{}
}

type Citation struct {
    URL   string
    Title string
}
```

---

## Processing Pipeline

### Stage 1: Upload & Validation

```go
type ImportService struct {
    db      *pgxpool.Pool
    storage *StorageClient // GCS or local
    parsers map[string]Parser
}

func (s *ImportService) StartImport(ctx context.Context, userID uuid.UUID, provider string, file io.Reader, filename string) (*FileImport, error) {
    // 1. Validate file size (max 100MB)
    // 2. Create import record (status: pending)
    // 3. Store file temporarily
    // 4. Queue processing job

    importRecord := &FileImport{
        ID:       uuid.New(),
        UserID:   userID,
        Provider: provider,
        Filename: filename,
        Status:   "pending",
    }

    // Save to DB
    // Queue async processing

    return importRecord, nil
}
```

### Stage 2: Parsing

```go
func (s *ImportService) ProcessImport(ctx context.Context, importID uuid.UUID) error {
    // 1. Load file from storage
    // 2. Get appropriate parser
    // 3. Parse to normalized format
    // 4. Update import status to "processing"

    parser := s.parsers[import.Provider]
    parsed, err := parser.Parse(fileData)
    if err != nil {
        return s.failImport(ctx, importID, err.Error())
    }

    // Continue to enrichment
    return s.EnrichAndStore(ctx, importID, parsed)
}
```

### Stage 3: Enrichment

```go
type EnrichmentService struct {
    llm     *AnthropicClient
    embedder *OpenAIEmbedder
}

func (s *EnrichmentService) EnrichConversation(ctx context.Context, conv *ParsedConversation) (*EnrichedConversation, error) {
    // 1. Generate summary using LLM
    summary, err := s.llm.Summarize(conv.Messages)

    // 2. Extract topics
    topics, err := s.llm.ExtractTopics(conv.Messages)

    // 3. Extract entities (people, companies, projects)
    entities, err := s.llm.ExtractEntities(conv.Messages)

    // 4. Classify category
    category, err := s.llm.Classify(conv.Messages)

    // 5. Extract knowledge (facts, decisions, insights)
    knowledge, err := s.llm.ExtractKnowledge(conv.Messages)

    return &EnrichedConversation{
        ParsedConversation: conv,
        Summary:   summary,
        Topics:    topics,
        Entities:  entities,
        Category:  category,
        Knowledge: knowledge,
    }, nil
}

func (s *EnrichmentService) GenerateEmbeddings(ctx context.Context, messages []ParsedMessage) ([][]float32, error) {
    // Generate embeddings for semantic search
    // Batch for efficiency (max 100 at a time)
}
```

### Stage 4: Storage & Integration

```go
func (s *ImportService) StoreEnriched(ctx context.Context, importID uuid.UUID, enriched *EnrichedConversation) error {
    tx, err := s.db.Begin(ctx)
    defer tx.Rollback(ctx)

    // 1. Store conversation
    convID, err := s.storeConversation(ctx, tx, importID, enriched)

    // 2. Store messages with embeddings
    for _, msg := range enriched.Messages {
        err = s.storeMessage(ctx, tx, convID, msg)
    }

    // 3. Store extracted knowledge
    for _, k := range enriched.Knowledge {
        err = s.storeKnowledge(ctx, tx, convID, k)

        // 4. Create memory entries for important knowledge
        if k.Type == "decision" || k.Type == "insight" {
            err = s.createMemory(ctx, tx, k)
        }
    }

    // 5. Link to existing entities if matches found
    err = s.linkEntities(ctx, tx, convID, enriched.Entities)

    return tx.Commit(ctx)
}
```

---

## Knowledge Extraction Prompts

### Summary Generation

```
Analyze this conversation and provide:
1. A concise summary (2-3 sentences)
2. Key topics discussed (list of 3-5 topics)
3. Any decisions made
4. Any action items or tasks mentioned

Conversation:
{messages}

Respond in JSON format:
{
  "summary": "...",
  "topics": ["topic1", "topic2"],
  "decisions": [{"decision": "...", "context": "..."}],
  "action_items": [{"task": "...", "assignee": "..."}]
}
```

### Entity Extraction

```
Extract all named entities from this conversation:
- People (names, roles)
- Companies/Organizations
- Projects
- Technologies/Tools
- Locations
- Dates/Deadlines

Conversation:
{messages}

Respond in JSON:
{
  "entities": [
    {"type": "person", "value": "John Smith", "context": "developer"},
    {"type": "company", "value": "Acme Corp", "context": "client"}
  ]
}
```

### Knowledge Extraction

```
Extract valuable knowledge from this conversation:
- Facts (verified information)
- Decisions (choices made)
- Insights (conclusions/realizations)
- Preferences (user preferences)
- Goals (objectives mentioned)

For each, rate confidence (0-1) and provide source quote.

Conversation:
{messages}

Respond in JSON:
{
  "knowledge": [
    {
      "type": "decision",
      "content": "Use PostgreSQL for the database",
      "confidence": 0.95,
      "source_quote": "Let's go with PostgreSQL for this project"
    }
  ]
}
```

---

## API Endpoints

### Import Management

```
POST   /api/imports/upload
       Body: multipart/form-data with file + provider
       Response: { import_id, status: "pending" }

GET    /api/imports
       Query: ?status=completed&provider=chatgpt&limit=20
       Response: { imports: [...], total: 50 }

GET    /api/imports/:id
       Response: { id, status, stats, conversations: [...] }

GET    /api/imports/:id/progress
       Response: { processed: 45, total: 100, current_conversation: "..." }

DELETE /api/imports/:id
       Deletes import and all associated data

POST   /api/imports/:id/retry
       Retry failed import
```

### Imported Data Access

```
GET    /api/imports/conversations
       Query: ?provider=all&search=api+design&from=2024-01-01
       Response: { conversations: [...], total: 200 }

GET    /api/imports/conversations/:id
       Response: { conversation with messages }

GET    /api/imports/conversations/:id/knowledge
       Response: { facts, decisions, insights extracted }

POST   /api/imports/conversations/:id/link
       Body: { project_id: "...", client_id: "..." }
       Link conversation to project/client
```

### Knowledge Search

```
GET    /api/imports/search
       Query: ?q=database+decision&type=decision
       Semantic search across all imported knowledge

GET    /api/imports/knowledge
       Query: ?type=fact&verified=true&tags=architecture
       Filter extracted knowledge
```

---

## Integration with Existing Modules

### Memories Module

Extracted knowledge can become memories:

```go
func (s *ImportService) createMemoryFromKnowledge(ctx context.Context, k *ExtractedKnowledge) error {
    memory := &Memory{
        UserID:      k.UserID,
        Title:       k.Title,
        Content:     k.Content,
        Type:        mapKnowledgeTypeToMemoryType(k.Type),
        Source:      "import:" + k.SourceProvider,
        SourceID:    k.ConversationID.String(),
        Tags:        k.Tags,
        Embedding:   k.Embedding,
        ImportedAt:  time.Now(),
    }

    return s.memoryService.Create(ctx, memory)
}
```

### Contexts Module

Conversations can create context nodes:

```go
func (s *ImportService) createContextFromConversation(ctx context.Context, conv *ImportedConversation) error {
    context := &Context{
        UserID:      conv.UserID,
        Title:       conv.Title,
        Summary:     conv.Summary,
        Type:        "imported_conversation",
        SourceType:  conv.Provider,
        SourceID:    conv.ExternalID,
        Topics:      conv.Topics,
        Entities:    conv.Entities,
        LinkedItems: []LinkedItem{
            {Type: "imported_conversation", ID: conv.ID},
        },
    }

    return s.contextService.Create(ctx, context)
}
```

### AI Chat Integration

The AI can reference imported knowledge:

```go
func (s *ChatService) GetRelevantContext(ctx context.Context, userID uuid.UUID, query string) ([]ContextItem, error) {
    // Search imported knowledge via embeddings
    knowledge, err := s.importService.SemanticSearch(ctx, userID, query, 5)

    // Include in context for AI
    items := make([]ContextItem, 0)
    for _, k := range knowledge {
        items = append(items, ContextItem{
            Type:    "imported_knowledge",
            Content: fmt.Sprintf("[From %s conversation]: %s", k.Provider, k.Content),
            Source:  k.ConversationTitle,
        })
    }

    return items, nil
}
```

---

## Configuration Options

### User Import Settings

```typescript
interface ImportSettings {
  // Privacy
  excludePatterns: string[];      // Regex patterns to exclude
  anonymizeNames: boolean;        // Replace names with placeholders

  // Processing
  generateSummaries: boolean;     // AI summarization (costs tokens)
  extractKnowledge: boolean;      // Extract facts/decisions
  createMemories: boolean;        // Auto-create memories
  createContexts: boolean;        // Auto-create context nodes

  // Linking
  autoLinkProjects: boolean;      // Try to link to projects
  autoLinkClients: boolean;       // Try to link to clients

  // Retention
  retentionDays: number;          // Auto-delete after N days (0 = forever)
}
```

### Per-Import Options

```typescript
interface ImportOptions {
  provider: 'chatgpt' | 'claude' | 'perplexity' | 'gemini' | 'granola';

  // What to import
  dateRange?: {
    from: Date;
    to: Date;
  };
  titleFilter?: string;           // Only import matching titles

  // Processing overrides
  skipSummaries?: boolean;
  skipKnowledge?: boolean;

  // Destination
  targetContextId?: string;       // Add to specific context
  tags?: string[];                // Apply tags to all imported items
}
```

---

## Frontend UI Components

### Import Page (`/settings/imports`)

```svelte
<script lang="ts">
  import { importStore } from '$lib/stores/importStore';

  let selectedProvider = '';
  let file: File | null = null;
  let importing = false;

  async function handleImport() {
    if (!file || !selectedProvider) return;

    importing = true;
    try {
      const result = await importStore.startImport(selectedProvider, file);
      // Show progress modal
    } finally {
      importing = false;
    }
  }
</script>

<div class="imports-page">
  <h1>Import Conversations</h1>

  <div class="provider-grid">
    {#each providers as provider}
      <button
        class="provider-card"
        class:selected={selectedProvider === provider.id}
        on:click={() => selectedProvider = provider.id}
      >
        <img src={provider.icon} alt={provider.name} />
        <span>{provider.name}</span>
      </button>
    {/each}
  </div>

  {#if selectedProvider}
    <div class="upload-area">
      <FileDropzone bind:file accept=".json" />

      <ImportOptions provider={selectedProvider} />

      <button on:click={handleImport} disabled={!file || importing}>
        {importing ? 'Importing...' : 'Start Import'}
      </button>
    </div>
  {/if}

  <ImportHistory />
</div>
```

### Import Progress Component

```svelte
<script lang="ts">
  export let importId: string;

  let progress = $state({ processed: 0, total: 0, status: 'pending' });

  $effect(() => {
    const interval = setInterval(async () => {
      progress = await fetchProgress(importId);
      if (progress.status === 'completed' || progress.status === 'failed') {
        clearInterval(interval);
      }
    }, 1000);

    return () => clearInterval(interval);
  });
</script>

<div class="import-progress">
  <div class="progress-bar">
    <div
      class="progress-fill"
      style="width: {(progress.processed / progress.total) * 100}%"
    />
  </div>

  <div class="stats">
    <span>{progress.processed} / {progress.total} conversations</span>
    <span>{progress.status}</span>
  </div>
</div>
```

---

## Security Considerations

### Data Handling

1. **Encryption at Rest**: All imported data encrypted in database
2. **Encryption in Transit**: HTTPS only
3. **Temporary File Cleanup**: Delete uploaded files after processing
4. **User Isolation**: Strict user_id checks on all queries

### Content Filtering

```go
func (s *ImportService) filterSensitiveContent(content string, settings *ImportSettings) string {
    // Remove credit card numbers
    content = creditCardRegex.ReplaceAllString(content, "[REDACTED]")

    // Remove SSN patterns
    content = ssnRegex.ReplaceAllString(content, "[REDACTED]")

    // Apply user exclusion patterns
    for _, pattern := range settings.ExcludePatterns {
        re := regexp.MustCompile(pattern)
        content = re.ReplaceAllString(content, "[EXCLUDED]")
    }

    // Anonymize names if requested
    if settings.AnonymizeNames {
        content = s.anonymizer.Anonymize(content)
    }

    return content
}
```

### Access Control

```go
func (h *ImportHandler) GetImport(w http.ResponseWriter, r *http.Request) {
    userID := auth.GetUserID(r.Context())
    importID := chi.URLParam(r, "id")

    imp, err := h.service.GetImport(r.Context(), importID)
    if err != nil {
        http.Error(w, "Not found", 404)
        return
    }

    // Strict ownership check
    if imp.UserID != userID {
        http.Error(w, "Forbidden", 403)
        return
    }

    json.NewEncoder(w).Encode(imp)
}
```

---

## Implementation Priority

### Phase 1: Core Infrastructure
1. Database migrations
2. File upload endpoint
3. ChatGPT parser (most common)
4. Basic storage (no enrichment)
5. List/view imported conversations

### Phase 2: Enrichment
1. Claude parser
2. Summary generation
3. Topic extraction
4. Embedding generation
5. Semantic search

### Phase 3: Knowledge Extraction
1. Perplexity parser (with citations)
2. Entity extraction
3. Knowledge extraction (facts, decisions)
4. Memory integration
5. Context integration

### Phase 4: Advanced Features
1. Granola parser (meetings)
2. Gemini parser
3. Auto-linking to projects/clients
4. Bulk operations
5. Export functionality

---

## File Structure

```
backend-go/
├── internal/
│   ├── handlers/
│   │   └── imports.go              # HTTP handlers
│   ├── services/
│   │   ├── imports.go              # Core import service
│   │   ├── import_parsers.go       # Parser implementations
│   │   └── import_enrichment.go    # AI enrichment
│   └── database/
│       └── queries/
│           └── imports.sql          # SQL queries

frontend/
├── src/
│   ├── lib/
│   │   ├── api/
│   │   │   └── imports.ts          # API client
│   │   ├── stores/
│   │   │   └── importStore.ts      # State management
│   │   └── components/
│   │       └── imports/
│   │           ├── ImportPage.svelte
│   │           ├── ProviderSelector.svelte
│   │           ├── FileDropzone.svelte
│   │           ├── ImportProgress.svelte
│   │           └── ImportHistory.svelte
│   └── routes/
│       └── (app)/
│           └── settings/
│               └── imports/
│                   └── +page.svelte
```

---

## Summary

The File Import system enables users to bring their AI conversation history into BusinessOS, where it becomes searchable, linkable knowledge. Key features:

1. **Multi-Provider Support**: ChatGPT, Claude, Perplexity, Gemini, Granola
2. **AI Enrichment**: Automatic summarization, topic extraction, entity recognition
3. **Knowledge Extraction**: Facts, decisions, and insights become first-class data
4. **Deep Integration**: Links to memories, contexts, projects, and clients
5. **Semantic Search**: Vector embeddings enable finding relevant past conversations
6. **Privacy Controls**: Filtering, anonymization, and retention policies
