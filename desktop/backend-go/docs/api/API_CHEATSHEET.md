# BusinessOS API Patterns - Quick Reference Cheat Sheet

## Function Naming Cheat Sheet

```typescript
// GET Operations
get[Resource]()           // Get single item or list
get[Resource]s()          // Get multiple items
get[Resource]Filtered()   // Get with filters/options

// POST Operations (Create)
create[Resource]()        // Create new item
save[APIKey]()           // Save configuration
import[Resource]()       // Import from external

// POST Operations (Action)
execute[Agent]()         // Run/execute something
warmup[Model]()          // Initialize something
pull[Model]()            // Download/fetch large data
test[Agent]()            // Validate before saving
clone[Agent]()           // Duplicate existing item

// PUT Operations (Update)
update[Resource]()       // Modify existing item

// DELETE Operations
delete[Resource]()       // Delete single item
delete[Resource]s()      // Delete multiple items (batch)

// Batch Operations
execute[Resource]Batch() // Execute multiple
delete[Resource]Batch()  // Delete multiple
```

## Request Pattern Cookbook

### Pattern 1: Simple GET
```typescript
export async function getAIProviders() {
  return request<AIProvidersResponse>('/ai/providers');
}
```

### Pattern 2: GET with ID
```typescript
export async function getCustomAgent(id: string) {
  return request<CustomAgent>(`/ai/custom-agents/${id}`);
}
```

### Pattern 3: GET with Query Parameters
```typescript
export async function getCustomAgents(includeInactive = false) {
  const params = includeInactive ? '?include_inactive=true' : '';
  return request<CustomAgentsResponse>(`/ai/custom-agents${params}`);
}
```

### Pattern 4: POST with Body
```typescript
export async function saveAPIKey(provider: string, apiKey: string) {
  return request<{ message: string }>('/ai/api-key', {
    method: 'POST',
    body: { provider, api_key: apiKey }  // snake_case for backend
  });
}
```

### Pattern 5: PUT with Body
```typescript
export async function updateCustomAgent(id: string, updates: Partial<CustomAgent>) {
  return request<CustomAgent>(`/ai/custom-agents/${id}`, {
    method: 'PUT',
    body: updates
  });
}
```

### Pattern 6: DELETE
```typescript
export async function deleteCustomAgent(id: string) {
  return request<{ message: string }>(`/ai/custom-agents/${id}`, {
    method: 'DELETE'
  });
}
```

### Pattern 7: Streaming (Special Case)
```typescript
export async function pullModel(model: string): Promise<ReadableStream<Uint8Array> | null> {
  const response = await fetch(`${getApiBaseUrl()}/ai/models/pull`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    credentials: 'include',
    body: JSON.stringify({ model })
  });

  if (!response.ok) {
    const error = await response.json().catch(() => ({ detail: 'Failed' }));
    throw new Error(error.detail || `Failed (HTTP ${response.status})`);
  }

  return response.body;  // Return raw stream
}
```

### Pattern 8: Batch Operation
```typescript
export async function executeCustomAgentsBatch(agentIds: string[], input: CustomAgentInput) {
  return request<BatchExecutionResult[]>('/ai/custom-agents/execute-batch', {
    method: 'POST',
    body: { agent_ids: agentIds, input }
  });
}
```

## Type Safety Quick Reference

### Generic Type Parameters
```typescript
// Always specify return type
request<ResponseType>(endpoint, options)

// Type inference works perfectly
const agents = await getCustomAgents();  // agents: CustomAgentsResponse
const agent = await getCustomAgent(id);  // agent: CustomAgent
const result = await executeCustomAgent(id, input);  // result: CustomAgentExecutionResponse
```

### Common Response Types
```typescript
// Single item
CustomAgent

// Array of items
{ agents: CustomAgent[] }

// Success/Failure union
{ success: true; result: string; error?: never } |
{ success: false; result?: never; error: string }

// Status message
{ message: string }

// Complex response
{
  agents: CustomAgent[];
  total: number;
  page: number;
}
```

### Request Body Types
```typescript
// Creating new item - Omit auto-generated fields
Omit<CustomAgent, 'id' | 'user_id' | 'created_at' | 'updated_at'>

// Partial updates - Only fields being changed
Partial<Omit<CustomAgent, 'id' | 'user_id' | 'created_at' | 'updated_at'>>

// Custom input types
{ prompt: string; context?: Record<string, unknown> }

// Batch operations
{ agent_ids: string[]; input: CustomAgentInput }
```

## Error Handling Quick Reference

### Standard Error Flow (Using request<T>)
```typescript
// Automatic error handling
try {
  const agent = await createCustomAgent({ ... });
} catch (error) {
  console.error(error.message);  // Already formatted with HTTP status
}

// Error format: "Error message (HTTP 404)"
```

### Streaming Error Flow (Custom fetch)
```typescript
// Manual error handling required
try {
  const stream = await executeCustomAgentStream(id, input);
  // ... process stream
} catch (error) {
  console.error(error.message);
}
```

### Validation Error Pattern
```typescript
// Check discriminated union
const validation = await testCustomAgent(config);

if (!validation.valid) {
  console.error('Errors:', validation.errors);
} else {
  console.log('Warnings:', validation.warnings);
  // Safe to create
}
```

## Common Use Cases

### Use Case 1: Create and Execute Agent
```typescript
const agent = await createCustomAgent({
  name: 'my-agent',
  display_name: 'My Agent',
  system_prompt: 'You are...',
  model_preference: 'claude-3-opus'
});

const result = await executeCustomAgent(agent.id, {
  prompt: 'Do something'
});

await deleteCustomAgent(agent.id);
```

### Use Case 2: Stream Long-Running Operation
```typescript
const stream = await executeCustomAgentStream(agentId, { prompt: 'Long task' });

if (stream) {
  const reader = stream.getReader();
  const decoder = new TextDecoder();

  while (true) {
    const { done, value } = await reader.read();
    if (done) break;
    process.stdout.write(decoder.decode(value));
  }
}
```

### Use Case 3: Parallel Execution
```typescript
const results = await executeCustomAgentsBatch(
  ['agent-1', 'agent-2', 'agent-3'],
  { prompt: 'Analyze this' },
  true  // Execute in parallel
);

results.forEach(r => {
  if (r.error) {
    console.error(`${r.agent_id}: ${r.error}`);
  } else {
    console.log(`${r.agent_id}: ${r.response}`);
  }
});
```

### Use Case 4: Validate Before Save
```typescript
const validation = await testCustomAgent(config);

if (!validation.valid) {
  console.error('Cannot save:', validation.errors);
  return;
}

if (validation.warnings) {
  console.warn('Warnings:', validation.warnings);
}

// Safe to create
const agent = await createCustomAgent(config);
```

### Use Case 5: Update Partially
```typescript
// Only update specific fields
const updated = await updateCustomAgent(agentId, {
  display_name: 'New Name',
  temperature: 0.5
  // Other fields unchanged
});
```

## Common Mistakes to Avoid

### ❌ Wrong: Using `any` type
```typescript
// Don't do this
const result: any = await createCustomAgent(...);
result.display_name.toLowerCase();  // No type checking!
```

### ✅ Correct: Using proper types
```typescript
// Do this
const agent: CustomAgent = await createCustomAgent(...);
agent.display_name.toLowerCase();  // Fully type-safe
```

---

### ❌ Wrong: Not specifying generic type
```typescript
// Don't do this
const response = request('/ai/custom-agents');
// response is unknown
```

### ✅ Correct: Specify generic type
```typescript
// Do this
const response = request<CustomAgentsResponse>('/ai/custom-agents');
// response is properly typed
```

---

### ❌ Wrong: Trying to use request<T>() for streams
```typescript
// Don't do this - will fail on binary stream data
const stream = await request<ReadableStream>('/ai/models/pull', {
  method: 'POST',
  body: { model }
});
```

### ✅ Correct: Use custom fetch for streams
```typescript
// Do this - returns raw stream
const response = await fetch(`${getApiBaseUrl()}/ai/models/pull`, {
  method: 'POST',
  body: JSON.stringify({ model }),
  credentials: 'include',
  headers: { 'Content-Type': 'application/json' }
});

if (!response.ok) throw new Error('Failed');
return response.body;
```

---

### ❌ Wrong: camelCase for backend fields
```typescript
// Don't do this - backend expects snake_case
body: { apiKey: '...' }
```

### ✅ Correct: snake_case for backend
```typescript
// Do this
body: { api_key: '...' }
```

---

### ❌ Wrong: Not handling discriminated union correctly
```typescript
// Don't do this - assumes valid is always present
if (validation.valid) {
  console.log(validation.errors);  // Wrong! errors might not exist
}
```

### ✅ Correct: Check discriminant first
```typescript
// Do this
if (!validation.valid) {
  console.log(validation.errors);  // TypeScript ensures errors exists
}
```

## Endpoint Structure Reference

```
/ai/
├── /providers           GET all providers
├── /provider            PUT change active provider
├── /models              GET all models
├── /models/local        GET local models
├── /models/pull         POST download model (stream)
├── /models/warmup       POST initialize model
├── /system              GET system info
├── /api-key             POST save API key
├── /agents              GET all agents
├── /agents/:id          GET specific agent
├── /custom-agents       GET all custom agents
├── /custom-agents?...   GET with filters
├── /custom-agents       POST create agent
├── /custom-agents/:id   GET agent details
├── /custom-agents/:id   PUT update agent
├── /custom-agents/:id   DELETE agent
├── /custom-agents/:id/execute          POST run agent
├── /custom-agents/:id/execute/stream   POST run with stream
├── /custom-agents/:id/clone            POST duplicate agent
├── /custom-agents/:id/stats            GET usage stats
├── /custom-agents/test                 POST validate config
├── /custom-agents/execute-batch        POST run multiple
└── /custom-agents/delete-batch         POST delete multiple

/mcp/
├── /tools               GET available tools
└── /execute             POST run a tool
```

## Quick Method Reference

| Method | Use | Pattern |
|--------|-----|---------|
| `get*()` | Fetch data | `request<T>()` |
| `create*()` | Create new | `request<T>(..., { method: 'POST', body })` |
| `save*()` | Persist config | `request<T>(..., { method: 'POST', body })` |
| `update*()` | Modify | `request<T>(..., { method: 'PUT', body })` |
| `delete*()` | Remove | `request<T>(..., { method: 'DELETE' })` |
| `execute*()` | Run/action | `request<T>(..., { method: 'POST', body })` |
| `execute*Stream()` | Stream response | Custom fetch + raw stream |
| `test*()` | Validate | `request<T>(..., { method: 'POST', body })` |
| `*Batch()` | Multiple items | `request<T[]>(..., { method: 'POST', body })` |

## Quick Command Reference

```bash
# Add new endpoint to ai.ts
1. Define return type (add to types.ts if needed)
2. Create function with proper name
3. Use request<T>() or custom fetch
4. Add JSDoc comments
5. Test with example

# Example function template
export async function [verb][Resource](params) {
  return request<ReturnType>('[/endpoint]', {
    method: '[METHOD]',
    body: { /* payload */ }
  });
}
```

