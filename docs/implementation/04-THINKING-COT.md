# P1: Thinking / Chain-of-Thought

> **Priority:** P1 - High Value
> **Backend Status:** Complete (13 endpoints)
> **Frontend Status:** Not Started
> **Estimated Effort:** 1 sprint

---

## Overview

Chain-of-Thought (COT) allows users to see the AI's reasoning process. This increases trust and helps users understand complex responses. Backend supports thinking traces, reasoning templates, and thinking settings.

---

## Backend API Endpoints

### Thinking Traces
| Method | Endpoint | Purpose |
|--------|----------|---------|
| GET | `/api/thinking/traces/:conversationId` | List thinking traces for conversation |
| GET | `/api/thinking/trace/:messageId` | Get thinking trace for specific message |
| DELETE | `/api/thinking/traces/:conversationId` | Delete thinking traces |

### Reasoning Templates
| Method | Endpoint | Purpose |
|--------|----------|---------|
| GET | `/api/thinking/templates` | List reasoning templates |
| POST | `/api/thinking/templates` | Create reasoning template |
| GET | `/api/thinking/templates/:id` | Get template |
| PUT | `/api/thinking/templates/:id` | Update template |
| DELETE | `/api/thinking/templates/:id` | Delete template |
| POST | `/api/thinking/templates/:id/default` | Set as default template |

### Thinking Settings
| Method | Endpoint | Purpose |
|--------|----------|---------|
| GET | `/api/thinking/settings` | Get thinking settings |
| PUT | `/api/thinking/settings` | Update thinking settings |

---

## Data Models

### Thinking Trace
```typescript
interface ThinkingTrace {
  id: string;
  message_id: string;
  conversation_id: string;

  // Reasoning steps
  steps: ThinkingStep[];

  // Metadata
  model_used: string;
  template_id?: string;
  duration_ms: number;
  token_count: number;

  created_at: string;
}

interface ThinkingStep {
  step_number: number;
  step_type: StepType;
  content: string;
  duration_ms: number;
}

type StepType =
  | 'understand'      // Understanding the question
  | 'analyze'         // Analyzing context
  | 'plan'            // Planning approach
  | 'reason'          // Core reasoning
  | 'evaluate'        // Evaluating options
  | 'conclude'        // Drawing conclusion
  | 'verify';         // Verifying answer
```

### Reasoning Template
```typescript
interface ReasoningTemplate {
  id: string;
  name: string;
  description: string;
  steps: TemplateStep[];
  is_default: boolean;
  created_at: string;
}

interface TemplateStep {
  name: string;
  prompt: string;
  required: boolean;
}
```

### Thinking Settings
```typescript
interface ThinkingSettings {
  enabled: boolean;           // Show thinking in UI
  show_by_default: boolean;   // Auto-expand thinking
  default_template_id?: string;
  save_traces: boolean;       // Persist traces
  max_steps: number;          // Limit reasoning steps
}
```

---

## Frontend Implementation Tasks

### Phase 1: Thinking Display in Chat

#### 1.1 ThinkingPanel Component
**File:** `src/lib/components/chat/ThinkingPanel.svelte`

```svelte
<script lang="ts">
  export let trace: ThinkingTrace;
  export let isExpanded = false;
</script>

<div class="thinking-panel">
  <button class="toggle" on:click={() => isExpanded = !isExpanded}>
    <BrainIcon />
    <span>Thinking... ({trace.steps.length} steps)</span>
    <ChevronIcon direction={isExpanded ? 'up' : 'down'} />
  </button>

  {#if isExpanded}
    <div class="steps" transition:slide>
      {#each trace.steps as step}
        <div class="step">
          <Badge>{step.step_type}</Badge>
          <p>{step.content}</p>
          <span class="duration">{step.duration_ms}ms</span>
        </div>
      {/each}
    </div>
    <div class="meta">
      <span>Model: {trace.model_used}</span>
      <span>Total: {trace.duration_ms}ms</span>
      <span>Tokens: {trace.token_count}</span>
    </div>
  {/if}
</div>
```

#### 1.2 Integrate into Message Component
**File:** Update `src/lib/components/chat/Message.svelte`

- [ ] Add ThinkingPanel above AI message content
- [ ] Show when thinking trace exists for message
- [ ] Animate thinking indicator during streaming

#### 1.3 Real-time Thinking Display
- [ ] During streaming, show steps as they happen
- [ ] Animate step appearance
- [ ] Show "Thinking..." with pulsing indicator

### Phase 2: Thinking Settings

#### 2.1 Thinking Settings Page
**File:** `src/routes/(app)/settings/ai/thinking/+page.svelte`

- [ ] Toggle: Enable thinking display
- [ ] Toggle: Show thinking by default (auto-expand)
- [ ] Toggle: Save thinking traces
- [ ] Max steps slider
- [ ] Default template selector

#### 2.2 Inline Toggle in Chat
- [ ] Quick toggle in chat header to show/hide thinking
- [ ] Keyboard shortcut: `Cmd+Shift+T`

### Phase 3: Reasoning Templates

#### 3.1 Templates Management Page
**File:** `src/routes/(app)/settings/ai/templates/+page.svelte`

- [ ] List all reasoning templates
- [ ] Create new template button
- [ ] Edit/delete template
- [ ] Set default template

#### 3.2 Template Editor
**File:** `src/lib/components/thinking/TemplateEditor.svelte`

- [ ] Name and description inputs
- [ ] Steps list (drag to reorder)
- [ ] Add/remove steps
- [ ] Step prompt editor
- [ ] Preview with sample question

#### 3.3 Built-in Templates
Display these preset templates:
- **Analytical**: Understand → Analyze → Evaluate → Conclude
- **Creative**: Explore → Ideate → Refine → Present
- **Problem Solving**: Define → Research → Hypothesize → Test → Solve
- **Step-by-Step**: Break down → Sequence → Execute → Verify

### Phase 4: API Client

#### 4.1 Thinking API
**File:** `src/lib/api/thinking/thinking.ts`

```typescript
export async function getThinkingTraces(conversationId: string): Promise<ThinkingTrace[]>
export async function getThinkingTrace(messageId: string): Promise<ThinkingTrace | null>
export async function deleteThinkingTraces(conversationId: string): Promise<void>

export async function getReasoningTemplates(): Promise<ReasoningTemplate[]>
export async function createReasoningTemplate(data: CreateTemplateInput): Promise<ReasoningTemplate>
export async function updateReasoningTemplate(id: string, data: UpdateTemplateInput): Promise<ReasoningTemplate>
export async function deleteReasoningTemplate(id: string): Promise<void>
export async function setDefaultTemplate(id: string): Promise<void>

export async function getThinkingSettings(): Promise<ThinkingSettings>
export async function updateThinkingSettings(data: Partial<ThinkingSettings>): Promise<ThinkingSettings>
```

#### 4.2 Thinking Store
**File:** `src/lib/stores/thinking.ts`

```typescript
interface ThinkingStore {
  settings: ThinkingSettings;
  templates: ReasoningTemplate[];
  currentTrace: ThinkingTrace | null;
  isLoading: boolean;

  // Actions
  loadSettings(): Promise<void>;
  updateSettings(data: Partial<ThinkingSettings>): Promise<void>;
  loadTemplates(): Promise<void>;
  setShowThinking(show: boolean): void;
}
```

---

## UI/UX Requirements

### Thinking Panel Design
- Collapsible by default (unless setting enabled)
- Subtle background color to distinguish from response
- Step-by-step appearance during streaming
- Visual progress indicator for steps

### Accessibility
- Screen reader announces "AI is thinking"
- Thinking content is focusable
- Keyboard navigation for expand/collapse

### Performance
- Don't block message display for thinking
- Lazy load thinking traces
- Cache traces per conversation

---

## Testing Requirements

- [ ] Unit tests for thinking store
- [ ] Component tests for ThinkingPanel
- [ ] E2E: Toggle thinking display
- [ ] E2E: Create reasoning template
- [ ] E2E: View thinking in chat

---

## Linear Issues to Create

1. **[COT-001]** Create ThinkingPanel component
2. **[COT-002]** Integrate thinking into Message component
3. **[COT-003]** Add real-time thinking during streaming
4. **[COT-004]** Build Thinking Settings page
5. **[COT-005]** Create Templates management page
6. **[COT-006]** Build Template Editor
7. **[COT-007]** API client and store implementation
8. **[COT-008]** E2E tests

---

## Dependencies

- Chat streaming must support thinking events

## Blockers

- None identified

---

## Notes

- This feature increases trust in AI responses
- Consider "lite" mode that just shows step names without content
- Could add "explain this step" feature in future
