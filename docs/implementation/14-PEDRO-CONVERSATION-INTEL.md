# Pedro Tasks: Conversation Intelligence

> **Priority:** P1 - High Value (Pedro's Work)
> **Backend Status:** Complete (6 endpoints)
> **Frontend Status:** Not Started
> **Owner:** Pedro
> **Estimated Effort:** 1 sprint

---

## Overview

Conversation Intelligence analyzes conversations to extract insights, action items, key decisions, and memories. This turns conversations into actionable knowledge.

---

## Backend API Endpoints

| Method | Endpoint | Purpose |
|--------|----------|---------|
| POST | `/api/intelligence/analyze` | Analyze conversation |
| GET | `/api/intelligence/conversations/:id` | Get stored analysis |
| GET | `/api/intelligence/conversations/search` | Search analyses |
| POST | `/api/intelligence/extract/conversation` | Extract memories from conversation |
| POST | `/api/intelligence/extract/voice-note` | Extract memories from voice note |
| GET | `/api/intelligence/memories` | Get all extracted memories |

---

## Data Models

```typescript
interface ConversationAnalysis {
  id: string;
  conversation_id: string;

  // Analysis Results
  summary: string;
  topics: Topic[];
  sentiment: SentimentAnalysis;
  action_items: ActionItem[];
  key_decisions: Decision[];
  questions_raised: string[];
  follow_ups: string[];

  // Metadata
  message_count: number;
  participant_count: number;
  duration_estimate?: string;

  analyzed_at: string;
}

interface Topic {
  name: string;
  relevance: number;  // 0-1
  mentions: number;
}

interface SentimentAnalysis {
  overall: 'positive' | 'neutral' | 'negative';
  score: number;  // -1 to 1
  by_section?: SectionSentiment[];
}

interface ActionItem {
  description: string;
  assignee?: string;
  due_date?: string;
  priority: 'high' | 'medium' | 'low';
  status: 'pending' | 'done';
  source_message_id: string;
}

interface Decision {
  description: string;
  rationale?: string;
  made_by?: string;
  source_message_id: string;
}

interface ExtractedMemory {
  id: string;
  conversation_id?: string;
  voice_note_id?: string;

  content: string;
  memory_type: 'fact' | 'decision' | 'preference' | 'learning' | 'action';
  confidence: number;
  source_text: string;

  created_at: string;
}
```

---

## Frontend Implementation Tasks

### Phase 1: Conversation Analysis View

#### 1.1 Analysis Panel in Chat
**File:** `src/lib/components/chat/AnalysisPanel.svelte`

- [ ] "Analyze Conversation" button in chat header
- [ ] Collapsible analysis panel
- [ ] Summary section
- [ ] Topics with relevance bars
- [ ] Sentiment indicator

#### 1.2 Action Items Section
**File:** `src/lib/components/intelligence/ActionItems.svelte`

- [ ] List of extracted action items
- [ ] Priority badges
- [ ] Assignee display
- [ ] "Create Task" button
- [ ] Mark as done

#### 1.3 Key Decisions Section
**File:** `src/lib/components/intelligence/KeyDecisions.svelte`

- [ ] List of decisions made
- [ ] Rationale display
- [ ] Link to source message
- [ ] "Save to Memories" button

#### 1.4 Questions & Follow-ups
- [ ] Questions raised during conversation
- [ ] Suggested follow-ups
- [ ] "Ask this" button to continue conversation

### Phase 2: Analysis History

#### 2.1 Intelligence Dashboard
**File:** `src/routes/(app)/intelligence/+page.svelte`

- [ ] List of analyzed conversations
- [ ] Search analyses
- [ ] Filter by: date, topic, sentiment
- [ ] Quick view summaries

#### 2.2 Analysis Detail Page
**File:** `src/routes/(app)/intelligence/[id]/+page.svelte`

- [ ] Full analysis display
- [ ] Link to original conversation
- [ ] Export analysis
- [ ] Re-analyze button

### Phase 3: Memory Extraction

#### 3.1 Memory Extraction UI
**File:** `src/lib/components/intelligence/MemoryExtractor.svelte`

- [ ] "Extract Memories" button
- [ ] Preview extracted memories
- [ ] Edit before saving
- [ ] Bulk save/dismiss

#### 3.2 Voice Note Memory Extraction
- [ ] Same interface for voice notes
- [ ] Show transcript alongside
- [ ] Extract from spoken content

### Phase 4: Auto-Analysis

#### 4.1 Auto-Analyze Settings
- [ ] Toggle: Auto-analyze after conversation ends
- [ ] Minimum message count threshold
- [ ] Notification when analysis ready

### Phase 5: API Client

#### 5.1 Intelligence API
**File:** `src/lib/api/intelligence/intelligence.ts`

```typescript
export async function analyzeConversation(
  conversationId: string
): Promise<ConversationAnalysis>

export async function getConversationAnalysis(
  conversationId: string
): Promise<ConversationAnalysis | null>

export async function searchAnalyses(
  query: string,
  filters?: AnalysisFilters
): Promise<ConversationAnalysis[]>

export async function extractConversationMemories(
  conversationId: string,
  options?: ExtractionOptions
): Promise<ExtractedMemory[]>

export async function extractVoiceNoteMemories(
  voiceNoteId: string,
  options?: ExtractionOptions
): Promise<ExtractedMemory[]>

export async function getExtractedMemories(
  filters?: MemoryFilters
): Promise<ExtractedMemory[]>
```

#### 5.2 Intelligence Store
**File:** `src/lib/stores/intelligence.ts`

```typescript
interface IntelligenceStore {
  currentAnalysis: ConversationAnalysis | null;
  analyses: ConversationAnalysis[];
  extractedMemories: ExtractedMemory[];
  isAnalyzing: boolean;

  analyzeConversation(id: string): Promise<void>;
  loadAnalysis(id: string): Promise<void>;
  extractMemories(conversationId: string): Promise<ExtractedMemory[]>;
  searchAnalyses(query: string): Promise<void>;
}
```

---

## UI/UX Requirements

### Analysis UX
- Show "Analyzing..." with progress
- Non-blocking (can continue chatting)
- Subtle notification when done
- Easy to dismiss/minimize

### Action Items UX
- One-click task creation
- Clear priority indication
- Checkable without navigation

### Memory Extraction UX
- Preview before saving
- Edit capability
- Confidence indicator
- Easy dismiss for irrelevant

---

## Testing Requirements

- [ ] Unit tests for intelligence store
- [ ] Component tests for analysis panel
- [ ] E2E: Trigger analysis
- [ ] E2E: Extract and save memories
- [ ] E2E: Create task from action item

---

## Linear Issues to Create

1. **[INTEL-001]** Add analysis trigger to chat
2. **[INTEL-002]** Build analysis panel component
3. **[INTEL-003]** Create action items section
4. **[INTEL-004]** Add key decisions section
5. **[INTEL-005]** Build intelligence dashboard
6. **[INTEL-006]** Implement memory extraction UI
7. **[INTEL-007]** Add voice note memory extraction
8. **[INTEL-008]** Create auto-analyze settings
9. **[INTEL-009]** API client and store
10. **[INTEL-010]** E2E tests

---

## Dependencies

- Ties into Memories (03-MEMORIES-USER-FACTS.md)
- Uses Tasks for action item creation

## Blockers

- None identified

---

## Notes

- This adds massive value to every conversation
- Consider analysis cost (LLM calls)
- Cache analyses to avoid re-computing
- Could generate weekly/monthly summaries
