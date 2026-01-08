# P0: Memories & User Facts

> **Priority:** P0 - Critical for Beta
> **Backend Status:** Complete (16 endpoints total)
> **Frontend Status:** Not Started
> **Estimated Effort:** 1-2 sprints

---

## Overview

Memories and User Facts enable AI personalization. The system learns from user interactions and stores:
- **Memories**: Episodic memories from conversations (facts, decisions, learnings)
- **User Facts**: Extracted patterns and preferences about the user

This makes the AI feel "intelligent" and personalized - a key differentiator.

---

## Backend API Endpoints

### Memories (Episodic)
| Method | Endpoint | Purpose |
|--------|----------|---------|
| GET | `/api/memories` | List memories with filters |
| POST | `/api/memories` | Create memory manually |
| GET | `/api/memories/stats` | Memory statistics |
| POST | `/api/memories/search` | Search memories |
| POST | `/api/memories/relevant` | Get contextually relevant memories |
| GET | `/api/memories/project/:projectId` | Get project-scoped memories |
| GET | `/api/memories/node/:nodeId` | Get node-scoped memories |
| GET | `/api/memories/:id` | Get specific memory |
| PUT | `/api/memories/:id` | Update memory |
| DELETE | `/api/memories/:id` | Delete memory |
| POST | `/api/memories/:id/pin` | Pin important memory |

### User Facts (Extracted Preferences)
| Method | Endpoint | Purpose |
|--------|----------|---------|
| GET | `/api/user-facts` | List all user facts |
| PUT | `/api/user-facts/:key` | Update a fact |
| POST | `/api/user-facts/:key/confirm` | Confirm AI-extracted fact |
| POST | `/api/user-facts/:key/reject` | Reject incorrect fact |
| DELETE | `/api/user-facts/:key` | Delete fact |

---

## Data Models

### Memory
```typescript
interface Memory {
  id: string;
  user_id: string;
  workspace_id: string;

  // Content
  content: string;          // The actual memory text
  memory_type: MemoryType;
  source: MemorySource;

  // Context
  conversation_id?: string;
  message_id?: string;
  project_id?: string;
  node_id?: string;

  // Metadata
  importance: number;       // 0-1 score
  is_pinned: boolean;
  access_count: number;
  last_accessed_at?: string;

  // Embedding
  embedding?: number[];     // For semantic search

  created_at: string;
  updated_at: string;
}

type MemoryType =
  | 'fact'        // A learned fact about user/domain
  | 'decision'    // A decision that was made
  | 'preference'  // A user preference
  | 'learning'    // Something learned
  | 'context'     // Contextual information
  | 'task'        // Task-related memory
  | 'goal';       // Goal or objective

type MemorySource =
  | 'conversation'  // Extracted from chat
  | 'voice_note'    // Extracted from voice
  | 'manual'        // User created
  | 'system';       // System generated
```

### User Fact
```typescript
interface UserFact {
  key: string;              // e.g., "preferred_language"
  value: string;            // e.g., "TypeScript"
  category: FactCategory;
  confidence: number;       // 0-1, how sure AI is
  source: string;           // Where it was learned
  is_confirmed: boolean;    // User confirmed accuracy
  created_at: string;
  updated_at: string;
}

type FactCategory =
  | 'preference'     // User preferences
  | 'expertise'      // Areas of expertise
  | 'workflow'       // Work patterns
  | 'communication'  // Communication style
  | 'personal'       // Personal info
  | 'business';      // Business-related
```

---

## Frontend Implementation Tasks

### Phase 1: Memory Viewer

#### 1.1 Memories Page
**File:** `src/routes/(app)/memories/+page.svelte`

- [ ] List view of all memories
- [ ] Filter by: type, source, date range, project, node
- [ ] Search memories (semantic search)
- [ ] Sort by: importance, date, access count
- [ ] Pinned memories section at top

#### 1.2 Memory Card Component
**File:** `src/lib/components/memories/MemoryCard.svelte`

```svelte
<div class="memory-card">
  <Badge>{memory.memory_type}</Badge>
  <p>{memory.content}</p>
  <div class="meta">
    <span>From: {memory.source}</span>
    <span>Importance: {formatPercent(memory.importance)}</span>
    {#if memory.project_id}
      <Link to="/projects/{memory.project_id}">View Project</Link>
    {/if}
  </div>
  <div class="actions">
    <Button variant="ghost" on:click={() => pin(memory)}>
      {memory.is_pinned ? 'Unpin' : 'Pin'}
    </Button>
    <Button variant="ghost" on:click={() => edit(memory)}>Edit</Button>
    <Button variant="ghost" color="red" on:click={() => delete(memory)}>Delete</Button>
  </div>
</div>
```

#### 1.3 Memory Stats Dashboard
**File:** `src/lib/components/memories/MemoryStats.svelte`

- [ ] Total memories count
- [ ] Breakdown by type (pie chart)
- [ ] Breakdown by source
- [ ] Most accessed memories
- [ ] Recent memories timeline

#### 1.4 Create Memory Modal
**File:** `src/lib/components/memories/CreateMemoryModal.svelte`

- [ ] Content textarea
- [ ] Type selector
- [ ] Project/Node association (optional)
- [ ] Importance slider

### Phase 2: User Facts Management

#### 2.1 User Facts Page
**File:** `src/routes/(app)/settings/facts/+page.svelte`

- [ ] Grouped list of facts by category
- [ ] Each fact shows: key, value, confidence, confirmed status
- [ ] Actions: Confirm, Reject, Edit, Delete

#### 2.2 Fact Card Component
**File:** `src/lib/components/facts/FactCard.svelte`

```svelte
<div class="fact-card" class:unconfirmed={!fact.is_confirmed}>
  <div class="header">
    <span class="key">{humanize(fact.key)}</span>
    <Badge>{fact.category}</Badge>
  </div>
  <p class="value">{fact.value}</p>
  <div class="meta">
    <span>Confidence: {formatPercent(fact.confidence)}</span>
    <span>Learned: {formatDate(fact.created_at)}</span>
  </div>
  {#if !fact.is_confirmed}
    <div class="confirm-prompt">
      <p>Is this accurate?</p>
      <Button on:click={() => confirm(fact.key)}>Yes</Button>
      <Button variant="outline" on:click={() => reject(fact.key)}>No</Button>
    </div>
  {/if}
  <Button variant="ghost" on:click={() => edit(fact)}>Edit</Button>
</div>
```

#### 2.3 Facts Review Flow
- [ ] Show unconfirmed facts prominently
- [ ] Inline confirm/reject buttons
- [ ] Toast notifications on action

### Phase 3: Chat Integration

#### 3.1 Memory Panel in Chat
**File:** Update `src/routes/(app)/chat/+page.svelte`

- [ ] Collapsible "Relevant Memories" panel
- [ ] Shows memories the AI used for context
- [ ] Click memory to see full content

#### 3.2 "Remember This" Action
- [ ] Button to manually create memory from message
- [ ] Context menu option on messages

#### 3.3 Memory Indicators
- [ ] Show when AI remembers something from past
- [ ] "I remember you mentioned..." style responses

### Phase 4: Project/Node Context

#### 4.1 Project Memory Tab
**File:** `src/routes/(app)/projects/[id]/+page.svelte`

- [ ] Add "Memories" tab to project detail
- [ ] List memories associated with project
- [ ] Create project-scoped memory

#### 4.2 Node Memory Tab
**File:** `src/routes/(app)/nodes/[id]/+page.svelte`

- [ ] Add "Memories" tab to node detail
- [ ] List memories associated with node
- [ ] Create node-scoped memory

### Phase 5: API Client

#### 5.1 Memories API
**File:** `src/lib/api/memories/memories.ts`

```typescript
export async function getMemories(filters?: MemoryFilters): Promise<Memory[]>
export async function createMemory(data: CreateMemoryInput): Promise<Memory>
export async function updateMemory(id: string, data: UpdateMemoryInput): Promise<Memory>
export async function deleteMemory(id: string): Promise<void>
export async function searchMemories(query: string): Promise<Memory[]>
export async function getRelevantMemories(context: string): Promise<Memory[]>
export async function pinMemory(id: string): Promise<void>
export async function unpinMemory(id: string): Promise<void>
export async function getMemoryStats(): Promise<MemoryStats>
export async function getProjectMemories(projectId: string): Promise<Memory[]>
export async function getNodeMemories(nodeId: string): Promise<Memory[]>
```

#### 5.2 User Facts API
**File:** `src/lib/api/userFacts/userFacts.ts`

```typescript
export async function getUserFacts(): Promise<UserFact[]>
export async function updateFact(key: string, value: string): Promise<UserFact>
export async function confirmFact(key: string): Promise<void>
export async function rejectFact(key: string): Promise<void>
export async function deleteFact(key: string): Promise<void>
```

#### 5.3 Memories Store
**File:** `src/lib/stores/memories.ts`

```typescript
interface MemoriesStore {
  memories: Memory[];
  pinnedMemories: Memory[];
  stats: MemoryStats | null;
  isLoading: boolean;

  // Actions
  loadMemories(filters?: MemoryFilters): Promise<void>;
  createMemory(data: CreateMemoryInput): Promise<void>;
  updateMemory(id: string, data: UpdateMemoryInput): Promise<void>;
  deleteMemory(id: string): Promise<void>;
  search(query: string): Promise<Memory[]>;
  pinMemory(id: string): Promise<void>;
}
```

---

## UI/UX Requirements

### Memory Visualization
- Use color coding for memory types
- Show importance with visual indicator (bar, stars, etc.)
- Pinned memories have star/pin icon

### User Facts UX
- Unconfirmed facts should be visually distinct
- Easy confirm/reject flow (swipe on mobile?)
- Grouping by category for easier scanning

### Privacy Considerations
- Clear explanation of what's being remembered
- Easy delete all data option
- Export memories option

---

## Testing Requirements

- [ ] Unit tests for memories store
- [ ] Unit tests for facts store
- [ ] Component tests for MemoryCard, FactCard
- [ ] E2E: Create memory flow
- [ ] E2E: Confirm/reject fact flow
- [ ] E2E: Search memories

---

## Linear Issues to Create

1. **[MEM-001]** Create Memories page with list view
2. **[MEM-002]** Build MemoryCard component
3. **[MEM-003]** Implement memory search
4. **[MEM-004]** Add Memory Stats dashboard
5. **[MEM-005]** Create User Facts settings page
6. **[MEM-006]** Build fact confirmation flow
7. **[MEM-007]** Integrate memory panel in chat
8. **[MEM-008]** Add memories to Project/Node pages
9. **[MEM-009]** API client and store implementation
10. **[MEM-010]** E2E tests

---

## Dependencies

- Pedro's Conversation Intelligence (extracts memories)
- Pedro's Learning module (extracts facts)

## Blockers

- None identified

---

## Notes

- This ties closely with Pedro's Learning & Personalization work
- Consider memory decay over time (less important = eventually delete)
- Privacy-first: users should have full control
