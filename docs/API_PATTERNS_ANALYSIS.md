# BusinessOS Frontend API Patterns Analysis

## Overview
This document details the API function patterns used in BusinessOS frontend for AI/agent endpoints, extracted from `frontend/src/lib/api/ai/ai.ts`.

---

## 1. Function Naming Conventions

### Pattern Structure
All API functions follow a consistent **verb-noun** pattern:

```typescript
[action][Resource/Endpoint]()
```

### Naming Categories

#### Read Operations (GET)
- `get` prefix for fetch operations
- Suffix indicates resource type
- Examples:
  - `getAIProviders()` - Fetch all providers
  - `getLocalModels()` - Fetch local models
  - `getAISystemInfo()` - Fetch system information
  - `getAgentPrompts()` - Fetch all agents
  - `getAgentPrompt(id)` - Fetch specific agent
  - `getTools()` - Fetch available tools
  - `getCustomAgents()` - Fetch custom agents

**Pattern**: `get[ResourceName]()` or `get[ResourceName](id)`

#### Write Operations (POST/PUT/DELETE)
- `update` prefix for PUT operations
- `save` prefix for POST operations (persist data)
- `pull` prefix for async operations (long-running)
- Examples:
  - `updateAIProvider(provider)` - Change active provider (PUT)
  - `saveAPIKey(provider, apiKey)` - Store API key (POST)
  - `pullModel(model)` - Download model (POST, streaming)
  - `warmupModel(model)` - Initialize model (POST)
  - `executeTool(toolName, args)` - Run tool (POST)

**Patterns**:
- `update[Resource](id, data)` - Modification
- `save[Resource](data)` - Creation/Persistence
- `pull[Resource](id)` - Async download
- `warmup[Resource](id)` - Initialization

#### Method Mapping
```
GET              → get*()
POST (create)    → save*()
POST (action)    → verb*()  (warmup, pull, execute)
PUT (update)     → update*()
DELETE           → delete* (not shown in current code)
```

---

## 2. Error Handling Patterns

### Pattern 1: Generic request<T>() Wrapper
Used for standard RESTful calls:

```typescript
export async function getAIProviders() {
  return request<AIProvidersResponse>('/ai/providers');
}
```

**Error handling delegated to base request utility:**
- Automatically handles non-2xx status codes
- Throws Error with formatted message: `${errorMessage} (HTTP ${status})`
- Extracts error from response: `error.detail || error.message`
- Logs to console for debugging

### Pattern 2: Custom Error Handling for Streaming
Used for endpoints returning streams (like `pullModel`):

```typescript
export async function pullModel(model: string): Promise<ReadableStream<Uint8Array> | null> {
  const response = await fetch(`${getApiBaseUrl()}/ai/models/pull`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    credentials: 'include',
    body: JSON.stringify({ model })
  });

  if (!response.ok) {
    const error = await response.json().catch(() => ({ detail: 'Pull model failed' }));
    throw new Error(error.detail || `Failed to pull model (HTTP ${response.status})`);
  }

  return response.body;  // Return raw stream
}
```

**Key differences:**
- Manual fetch instead of `request<T>()` wrapper
- Custom error recovery with `.catch()` fallback
- Returns `response.body` (ReadableStream) not parsed JSON
- Allows consumer to handle stream processing

### Pattern 3: Type-Safe Discriminated Union
Used for responses with multiple outcomes:

```typescript
export type ToolResponse =
  | { success: true; result: string; error?: never }
  | { success: false; result?: never; error: string };
```

**Benefits:**
- Prevents invalid state (can't have both success=true and error)
- TypeScript enforces correct field access:
  ```typescript
  if (response.success) {
    // Only result is available here
    console.log(response.result);
  } else {
    // Only error is available here
    console.log(response.error);
  }
  ```

### Error Recovery Strategy
```
Request fails (non-2xx)
    ↓
Try parse JSON error response
    ↓
Fallback to generic message if parse fails
    ↓
Throw Error with formatted message including HTTP status
    ↓
Caller can catch and handle
```

---

## 3. Type Safety Approach

### Approach: Generic Type Parameters
All request functions use TypeScript generics:

```typescript
request<ResponseType>(endpoint, options)
```

**Benefits:**
- Compiler enforces correct return types
- IDE autocomplete works for response objects
- No need for manual type casting

### Response Type Hierarchy

```
Single Request Response
├── AIProvidersResponse { providers, active_provider, default_model }
├── AllModelsResponse { models, active_provider, default_model }
├── AISystemInfo { total_ram_gb, available_ram_gb, platform, has_gpu }
├── WarmupResponse { status, model, provider, message }
└── CustomAgentsResponse { agents: CustomAgent[] }

Container Responses
├── { message: string }              (simple status)
├── { agents: AgentInfo[] }          (array wrapper)
├── { tools: Tool[] }                (array wrapper)
├── { id: string; prompt: string }   (object wrapper)
└── ToolResponse (discriminated union)

Data Models
├── LLMProvider { id, name, type, description, configured, base_url? }
├── LLMModel { id, name, provider, description?, size?, family? }
├── AgentInfo { id, name, description, prompt, category }
├── Tool { name, description, input_schema, source }
├── CustomAgent { id, user_id, name, display_name, system_prompt, ... }
└── RecommendedModel { name, description, ram_required, speed, quality }
```

### Optional Fields Pattern
All interfaces use optional fields (`?`) for optional data:

```typescript
interface CustomAgent {
  id: string;                    // Required
  user_id: string;              // Required
  description?: string;         // Optional
  avatar?: string;              // Optional
  temperature?: number;         // Optional
  capabilities?: string[];      // Optional array
}
```

### Request Options Type
```typescript
interface RequestOptions {
  method?: string;           // Defaults to 'GET'
  body?: unknown;           // Auto-serialized to JSON
  headers?: Record<string, string>;  // Custom headers
}
```

---

## 4. Request/Response Structure

### Standard Request Flow

#### Simple GET Request
```typescript
export async function getAIProviders() {
  return request<AIProvidersResponse>('/ai/providers');
}

// Produces:
// GET /api/ai/providers
// Headers: { credentials: 'include' }
```

#### POST with Payload
```typescript
export async function saveAPIKey(provider: string, apiKey: string) {
  return request<{ message: string }>('/ai/api-key', {
    method: 'POST',
    body: { provider, api_key: apiKey }  // Snake_case for backend
  });
}

// Produces:
// POST /api/ai/api-key
// Content-Type: application/json
// { "provider": "...", "api_key": "..." }
```

#### PUT Request
```typescript
export async function updateAIProvider(provider: string) {
  return request<{ message: string }>('/ai/provider', {
    method: 'PUT',
    body: { provider }
  });
}
```

### Request Headers
Always included automatically:
- `Content-Type: application/json` (added if body present)
- `credentials: 'include'` (enable cookies)

### Response Structure

#### Success Response
```typescript
// Status: 2xx
{
  providers: [ ... ],
  active_provider: "anthropic",
  default_model: "claude-3-opus"
}
```

#### Error Response
```typescript
// Status: 4xx/5xx
{
  detail: "Provider not found",  // Extracted and thrown
  message?: "Alternative error field"
}

// Thrown as:
throw new Error("Provider not found (HTTP 404)")
```

### Special Cases

#### Streaming Response (No JSON Parsing)
```typescript
export async function pullModel(model: string): Promise<ReadableStream<Uint8Array> | null> {
  // Return raw stream, not parsed JSON
  return response.body;  // ReadableStream<Uint8Array>
}
```

#### Query Parameters
```typescript
export async function getCustomAgents(includeInactive = false) {
  const params = includeInactive ? '?include_inactive=true' : '';
  return request<CustomAgentsResponse>(`/ai/custom-agents${params}`);
}

// GET /api/ai/custom-agents?include_inactive=true
```

---

## 5. API Endpoint Structure

### Endpoint Organization
```
/ai/               AI/LLM endpoints
├── /providers           GET all providers
├── /provider            PUT update active provider
├── /models              GET all models
├── /models/local        GET local models only
├── /models/pull         POST pull new model (streaming)
├── /models/warmup       POST initialize model
├── /system              GET system information
├── /api-key             POST save API key
├── /agents              GET all agents
├── /agents/:id          GET specific agent
├── /custom-agents       GET custom agents
└── /custom-agents?include_inactive=true

/mcp/              MCP tool endpoints
├── /tools               GET available tools
└── /execute             POST run a tool
```

### Endpoint Patterns
- **Collection endpoints**: `/ai/providers` (GET returns array)
- **Resource endpoints**: `/ai/agents/:id` (GET returns single item)
- **Action endpoints**: `/ai/models/pull` (POST triggers operation)
- **Filter parameters**: `?include_inactive=true`

---

## 6. Template for Adding New Custom Agent Endpoints

### Complete Template

```typescript
// ═══════════════════════════════════════════════════════════════════════════════
// CUSTOM AGENT ENDPOINTS - Add new functions here
// ═══════════════════════════════════════════════════════════════════════════════

// ────────────────────────────────────────────────────────────────────────────────
// CREATE - Save new custom agent
// ────────────────────────────────────────────────────────────────────────────────
export async function createCustomAgent(agent: Omit<CustomAgent, 'id' | 'user_id' | 'created_at' | 'updated_at'>) {
  return request<CustomAgent>('/ai/custom-agents', {
    method: 'POST',
    body: agent
  });
}

// ────────────────────────────────────────────────────────────────────────────────
// READ - Get specific custom agent
// ────────────────────────────────────────────────────────────────────────────────
export async function getCustomAgent(id: string) {
  return request<CustomAgent>(`/ai/custom-agents/${id}`);
}

// ────────────────────────────────────────────────────────────────────────────────
// UPDATE - Modify existing custom agent
// ────────────────────────────────────────────────────────────────────────────────
export async function updateCustomAgent(
  id: string,
  updates: Partial<Omit<CustomAgent, 'id' | 'user_id' | 'created_at' | 'updated_at'>>
) {
  return request<CustomAgent>(`/ai/custom-agents/${id}`, {
    method: 'PUT',
    body: updates
  });
}

// ────────────────────────────────────────────────────────────────────────────────
// DELETE - Remove custom agent
// ────────────────────────────────────────────────────────────────────────────────
export async function deleteCustomAgent(id: string) {
  return request<{ message: string }>(`/ai/custom-agents/${id}`, {
    method: 'DELETE'
  });
}

// ────────────────────────────────────────────────────────────────────────────────
// ACTION - Execute custom agent
// ────────────────────────────────────────────────────────────────────────────────
export async function executeCustomAgent(
  id: string,
  input: { prompt: string; context?: Record<string, unknown> }
) {
  return request<{ response: string; thinking?: string }>(`/ai/custom-agents/${id}/execute`, {
    method: 'POST',
    body: input
  });
}

// ────────────────────────────────────────────────────────────────────────────────
// STREAMING - Stream custom agent response (for long-running operations)
// ────────────────────────────────────────────────────────────────────────────────
export async function executeCustomAgentStream(
  id: string,
  input: { prompt: string; context?: Record<string, unknown> }
): Promise<ReadableStream<Uint8Array> | null> {
  const response = await fetch(`${getApiBaseUrl()}/ai/custom-agents/${id}/execute/stream`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    credentials: 'include',
    body: JSON.stringify(input)
  });

  if (!response.ok) {
    const error = await response.json().catch(() => ({ detail: 'Stream execution failed' }));
    throw new Error(error.detail || `Failed to execute agent stream (HTTP ${response.status})`);
  }

  return response.body;
}

// ────────────────────────────────────────────────────────────────────────────────
// BATCH - Execute multiple custom agents in parallel
// ────────────────────────────────────────────────────────────────────────────────
export async function executeCustomAgentsBatch(
  agentIds: string[],
  input: { prompt: string; context?: Record<string, unknown> }
) {
  return request<
    Array<{ agent_id: string; response: string; error?: string }>
  >('/ai/custom-agents/execute-batch', {
    method: 'POST',
    body: { agent_ids: agentIds, input }
  });
}

// ────────────────────────────────────────────────────────────────────────────────
// CLONE - Duplicate an existing custom agent
// ────────────────────────────────────────────────────────────────────────────────
export async function cloneCustomAgent(id: string, name: string) {
  return request<CustomAgent>(`/ai/custom-agents/${id}/clone`, {
    method: 'POST',
    body: { name }
  });
}

// ────────────────────────────────────────────────────────────────────────────────
// TEST - Validate agent configuration before saving
// ────────────────────────────────────────────────────────────────────────────────
export async function testCustomAgent(
  agent: Omit<CustomAgent, 'id' | 'user_id' | 'created_at' | 'updated_at'>
) {
  return request<
    { valid: true; warnings?: string[] } | { valid: false; errors: string[] }
  >('/ai/custom-agents/test', {
    method: 'POST',
    body: agent
  });
}
```

### Type Extensions for Custom Agent Operations

Add to `types.ts`:

```typescript
// Custom Agent Request/Response Types
export interface CustomAgentInput {
  prompt: string;
  context?: Record<string, unknown>;
}

export interface CustomAgentExecutionResponse {
  response: string;
  thinking?: string;
  tokens_used?: number;
  execution_time_ms?: number;
}

export interface CustomAgentBatchRequest {
  agent_ids: string[];
  input: CustomAgentInput;
}

export interface BatchExecutionResult {
  agent_id: string;
  response: string;
  error?: string;
  execution_time_ms?: number;
}

export interface CustomAgentValidation {
  valid: boolean;
  errors?: string[];
  warnings?: string[];
  suggestions?: string[];
}
```

---

## 7. Usage Examples

### Example 1: Creating a Custom Agent
```typescript
import { createCustomAgent } from '$lib/api/ai/ai';

try {
  const agent = await createCustomAgent({
    name: 'code-reviewer',
    display_name: 'Code Reviewer',
    description: 'Reviews code for quality and best practices',
    system_prompt: 'You are an expert code reviewer...',
    model_preference: 'claude-3-opus',
    temperature: 0.2,
    tools_enabled: ['git', 'linter'],
    thinking_enabled: true,
    streaming_enabled: false
  });

  console.log('Created agent:', agent.id);
} catch (error) {
  console.error('Failed to create agent:', error.message);
}
```

### Example 2: Executing a Custom Agent
```typescript
import { executeCustomAgent } from '$lib/api/ai/ai';

try {
  const result = await executeCustomAgent('code-reviewer-123', {
    prompt: 'Review this code for security issues',
    context: { file_path: 'auth.ts', language: 'typescript' }
  });

  console.log('Agent response:', result.response);
} catch (error) {
  console.error('Execution failed:', error.message);
}
```

### Example 3: Streaming Custom Agent Response
```typescript
import { executeCustomAgentStream } from '$lib/api/ai/ai';

try {
  const stream = await executeCustomAgentStream('debugger-456', {
    prompt: 'Debug this error: Connection timeout on port 5432'
  });

  if (stream) {
    const reader = stream.getReader();
    const decoder = new TextDecoder();

    while (true) {
      const { done, value } = await reader.read();
      if (done) break;

      const chunk = decoder.decode(value);
      process.stdout.write(chunk);  // Or update UI
    }
  }
} catch (error) {
  console.error('Stream failed:', error.message);
}
```

### Example 4: Validating Before Save
```typescript
import { testCustomAgent, createCustomAgent } from '$lib/api/ai/ai';

const agentConfig = {
  name: 'documentation-generator',
  display_name: 'Doc Generator',
  system_prompt: 'Generate comprehensive documentation...',
  model_preference: 'claude-3-sonnet'
};

try {
  // First validate
  const validation = await testCustomAgent(agentConfig);

  if (!validation.valid) {
    console.error('Validation errors:', validation.errors);
    return;
  }

  if (validation.warnings) {
    console.warn('Warnings:', validation.warnings);
  }

  // Then create
  const agent = await createCustomAgent(agentConfig);
  console.log('Agent created:', agent.id);
} catch (error) {
  console.error('Failed:', error.message);
}
```

---

## 8. Best Practices Summary

### Naming
- ✅ Use verb-noun pattern: `get*`, `save*`, `update*`, `execute*`
- ✅ Be specific: `getCustomAgents` not `getAgents` (confusing with `getAgentPrompts`)
- ❌ Avoid: Generic names like `fetch()`, `call()`, `run()`

### Error Handling
- ✅ Let `request<T>()` handle standard errors
- ✅ Use custom try/catch for streaming endpoints
- ✅ Provide user-friendly error messages with HTTP status
- ❌ Avoid: Swallowing errors, vague messages

### Type Safety
- ✅ Always use generic types: `request<ResponseType>()`
- ✅ Use discriminated unions for multiple outcomes
- ✅ Mark optional fields with `?`
- ❌ Avoid: `any` types, implicit `unknown`

### Request Structure
- ✅ Include all necessary context in request body
- ✅ Use snake_case for backend field names (`api_key`, `user_id`)
- ✅ Use optional parameters for filters
- ❌ Avoid: Over-complicating request payloads

### Response Handling
- ✅ Parse JSON responses automatically with `request<T>()`
- ✅ Return raw streams only for streaming endpoints
- ✅ Include metadata in responses (timestamps, counts, etc.)
- ❌ Avoid: Mixing JSON and stream responses in same function

---

## 9. Common Patterns Quick Reference

| Pattern | Use Case | Example |
|---------|----------|---------|
| `request<T>(url)` | GET single item or list | `getAIProviders()` |
| `request<T>(url, { method: 'POST', body })` | Create/Save | `saveAPIKey(...)` |
| `request<T>(url, { method: 'PUT', body })` | Update | `updateAIProvider(...)` |
| `request<T>(url, { method: 'DELETE' })` | Delete | `deleteCustomAgent(id)` |
| Custom fetch + stream | Long operations | `pullModel(...)` |
| Discriminated union | Safe multi-state | `ToolResponse` type |
| Optional params | Filters/options | `getCustomAgents(includeInactive)` |
| `Omit<T, ...>` | Request without IDs | `createCustomAgent(agentData)` |

