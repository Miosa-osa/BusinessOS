# Extended Thinking (Chain of Thought) System - Integration Guide

## 📚 Overview

The Extended Thinking system provides Chain of Thought (COT) reasoning capabilities, allowing the AI to show its reasoning process step-by-step. This document explains how the system is integrated and how to use it.

## 🏗️ Architecture

### Backend Components (Go)
Located in `desktop/backend-go/internal/`

**Database Tables:**
- `thinking_traces` - Stores thinking process traces
- `thinking_steps` - Individual reasoning steps
- `reasoning_templates` - Custom reasoning templates
- `thinking_settings` - User configuration

**API Endpoints:**
```
GET    /api/thinking/traces/:conversation_id    - Get traces for conversation
GET    /api/thinking/trace/:message_id          - Get trace for message
DELETE /api/thinking/traces/:conversation_id    - Delete traces

GET    /api/reasoning/templates                 - List templates
GET    /api/reasoning/templates/:id             - Get template
POST   /api/reasoning/templates                 - Create template
PUT    /api/reasoning/templates/:id             - Update template
DELETE /api/reasoning/templates/:id             - Delete template
POST   /api/reasoning/templates/:id/default     - Set as default

GET    /api/thinking/settings                   - Get settings
PUT    /api/thinking/settings                   - Update settings
```

### Frontend Components (SvelteKit)
Located in `frontend/src/lib/`

**API Client:** `lib/api/thinking/`
- `thinking.ts` - 11 API functions
- `types.ts` - TypeScript interfaces
- `index.ts` - Barrel exports

**Store:** `lib/stores/thinking.ts`
- 17 methods for managing state
- Caching with Map structures
- Derived stores for computed values

**UI Components:**
- `lib/components/chat/ThinkingPanel.svelte` - Display thinking traces in chat
- `routes/(app)/settings/ai/thinking/+page.svelte` - Settings page
- `routes/(app)/settings/ai/templates/+page.svelte` - Template management

## 🎯 How It Works

### 1. Enable Extended Thinking

Go to `/settings/ai/thinking` and configure:

```typescript
{
  enabled: true,              // Allow AI to use thinking
  show_in_ui: true,          // Show thinking panel in chat
  save_traces: true,         // Save to database
  max_tokens: 4096,          // Max tokens for thinking
  default_template_id: null  // Optional template
}
```

### 2. Thinking Flow

When a user sends a message with thinking enabled:

```
User Message
    ↓
Backend receives request
    ↓
AI starts thinking (if complex query)
    ↓
Backend emits SSE events:
  - { step: 'analyzing', content: '...', completed: false }
  - { step: 'planning', content: '...', completed: false }
  - { step: 'reasoning', content: '...', completed: false }
  - { step: 'concluding', content: '...', completed: true }
    ↓
Frontend receives events
    ↓
ThinkingPanel displays in real-time
    ↓
Trace saved to database (if enabled)
    ↓
Final response shown
```

### 3. Frontend Integration

**In Chat Component:**

```svelte
<script>
  import { thinking } from '$lib/stores/thinking';
  import ThinkingPanel from '$lib/components/chat/ThinkingPanel.svelte';

  // Listen for thinking events from SSE
  function handleThinkingEvent(event) {
    if (event.step) {
      thinking.updateThinkingStep({
        type: event.step,
        content: event.content
      });
    }

    if (event.completed) {
      thinking.completeThinking(trace);
    }
  }
</script>

<!-- In message display -->
{#if message.thinking_trace}
  <ThinkingPanel
    trace={message.thinking_trace}
    isStreaming={false}
    bind:isExpanded={expanded}
  />
{/if}
```

**ThinkingPanel Component:**

The panel shows:
- 💡 Lightbulb icon
- Collapsible header with metadata (tokens, duration, model)
- Step-by-step reasoning with colored badges:
  - 🔵 Exploration (blue)
  - 🟣 Analysis (purple)
  - 🟢 Synthesis (green)
  - 🟡 Conclusion (amber)
  - 🟢 Verification (teal)
- Streaming cursor when active

### 4. Using the Store

**Load Settings:**
```typescript
import { thinking } from '$lib/stores/thinking';

await thinking.loadSettings();
```

**Update Settings:**
```typescript
await thinking.updateSettings({
  enabled: true,
  show_in_ui: true,
  save_traces: true,
  max_tokens: 8192,
  default_template_id: null
});
```

**Get Trace for Message:**
```typescript
const trace = await thinking.getTraceForMessage(conversationId, messageId);
```

**Create Template:**
```typescript
await thinking.createTemplate({
  name: "Analytical Thinking",
  description: "Deep analysis for complex problems",
  steps: [
    { order: 0, type: 'exploration', prompt: 'Understand the problem...' },
    { order: 1, type: 'analysis', prompt: 'Analyze components...' },
    { order: 2, type: 'conclusion', prompt: 'Synthesize solution...' }
  ]
});
```

## 📊 Data Flow

### Settings Update Flow:
```
UI (Toggle/Input)
    ↓
thinking.updateSettings(data)
    ↓
PUT /api/thinking/settings
    ↓
Database Update
    ↓
Store State Updated
    ↓
UI Re-renders
```

### Template Management Flow:
```
UI (Create/Edit Form)
    ↓
thinking.createTemplate(data)
    ↓
POST /api/reasoning/templates
    ↓
Database Insert
    ↓
Store Cache Updated
    ↓
Templates List Refreshed
```

### Trace Display Flow:
```
Chat receives message
    ↓
Check if message has thinking_trace
    ↓
Render ThinkingPanel component
    ↓
User clicks to expand/collapse
    ↓
Show steps with colored badges
    ↓
Display metadata (tokens, duration, model)
```

## 🎨 UI Components

### ThinkingPanel Props:
```typescript
interface Props {
  trace: ThinkingTrace | APIThinkingTrace;
  isStreaming?: boolean;
  isExpanded?: boolean;  // $bindable
}
```

### ThinkingTrace Interface:
```typescript
interface ThinkingTrace {
  id: string;
  user_id: string;
  conversation_id: string;
  message_id: string;
  thinking_content: string;
  thinking_type?: ThinkingType;
  step_number?: number;
  started_at: string;
  completed_at?: string;
  duration_ms?: number;
  thinking_tokens?: number;
  model_used?: string;
  reasoning_template_id?: string;
  metadata?: Record<string, unknown>;
  created_at: string;
}
```

## 🔧 Configuration

### Environment Variables (Backend):
```bash
# None required - uses existing DB connection
```

### Default Settings:
```typescript
{
  enabled: true,
  show_in_ui: false,  // Don't show by default
  save_traces: true,
  max_tokens: 4096,
  default_template_id: null
}
```

## 🧪 Testing

### Manual Testing:

1. **Enable Thinking:**
   - Go to `/settings/ai/thinking`
   - Toggle "Enable Extended Thinking"
   - Toggle "Show Thinking by Default"
   - Click "Save Settings"

2. **Send Complex Query:**
   - Go to chat
   - Send: "Explain quantum computing in detail"
   - Watch for ThinkingPanel to appear
   - Click to expand and see steps

3. **Create Template:**
   - Go to `/settings/ai/templates`
   - Click "Create Template"
   - Fill in name, description, steps
   - Save and set as default

4. **Verify Trace Saved:**
   - Check database: `SELECT * FROM thinking_traces`
   - Should see trace with steps

### Integration Testing:

```typescript
// Test thinking flow
test('thinking flow', async () => {
  // Enable thinking
  await thinking.loadSettings();
  await thinking.updateSettings({ enabled: true, show_in_ui: true });

  // Send message
  const response = await sendMessage("Complex query");

  // Check trace
  expect(response.thinking_trace).toBeDefined();
  expect(response.thinking_trace.steps).toHaveLength(greaterThan(0));
});
```

## 📝 Logs

The system logs to console in development:

```
[COT] Thinking event: {step: 'analyzing', content: '...', completed: false}
[COT] Thinking started (preserved search): no
[COT] Thinking completed: {step: 1}
```

Monitor these logs to debug thinking flow.

## 🚀 Deployment

### Database Migration:

Ensure migrations are run:
```sql
-- Already in migrations/
-- thinking_traces table
-- reasoning_templates table
-- thinking_settings table
```

### Backend Deployment:

No special configuration needed. Endpoints are automatically available.

### Frontend Build:

```bash
cd frontend
npm run build
```

ThinkingPanel is code-split and lazy-loaded.

## 🔍 Troubleshooting

### Thinking Not Showing:

1. Check settings: `enabled: true` and `show_in_ui: true`
2. Check console for `[COT]` logs
3. Verify backend is emitting SSE events
4. Check ThinkingPanel component is imported in chat

### Traces Not Saving:

1. Check `save_traces: true` in settings
2. Verify database tables exist
3. Check backend logs for errors
4. Confirm API endpoints responding

### Templates Not Loading:

1. Check `/api/reasoning/templates` returns data
2. Verify store.loadTemplates() is called
3. Check console for errors
4. Refresh page to clear cache

## 📚 API Usage Examples

### Get Settings:
```bash
curl http://localhost:8080/api/thinking/settings \
  -H "Authorization: Bearer $TOKEN"
```

### Update Settings:
```bash
curl -X PUT http://localhost:8080/api/thinking/settings \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "enabled": true,
    "show_in_ui": true,
    "save_traces": true,
    "max_tokens": 8192,
    "default_template_id": null
  }'
```

### Create Template:
```bash
curl -X POST http://localhost:8080/api/reasoning/templates \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Analytical Thinking",
    "description": "Deep analysis",
    "steps": [
      {"order": 0, "type": "exploration", "prompt": "Understand..."},
      {"order": 1, "type": "analysis", "prompt": "Analyze..."},
      {"order": 2, "type": "conclusion", "prompt": "Conclude..."}
    ]
  }'
```

### Get Trace:
```bash
curl http://localhost:8080/api/thinking/trace/:message_id \
  -H "Authorization: Bearer $TOKEN"
```

## 🎯 Best Practices

1. **Enable thinking for complex queries only** - Simple queries don't need it
2. **Save traces for debugging** - Helps understand AI reasoning
3. **Create domain-specific templates** - Guide AI for specific tasks
4. **Monitor token usage** - Thinking uses additional tokens
5. **Use collapsible UI** - Don't overwhelm users with thinking details

## 🔗 Related Files

- Backend: `desktop/backend-go/internal/handlers/thinking.go`
- Frontend Store: `frontend/src/lib/stores/thinking.ts`
- UI Component: `frontend/src/lib/components/chat/ThinkingPanel.svelte`
- Settings Page: `frontend/src/routes/(app)/settings/ai/thinking/+page.svelte`
- Templates Page: `frontend/src/routes/(app)/settings/ai/templates/+page.svelte`
- API Types: `frontend/src/lib/api/thinking/types.ts`

---

**Last Updated:** 2026-01-09
**Status:** ✅ Fully Implemented and Tested
