# BusinessOS Frontend API Documentation Index

## Overview
This index documents the API patterns used in BusinessOS frontend for AI/agent endpoints. All patterns are extracted from production code in `frontend/src/lib/api/ai/ai.ts`.

---

## Documentation Files

### 1. **API_PATTERNS_ANALYSIS.md** (Comprehensive Reference)
Complete deep-dive analysis of API patterns used in BusinessOS.

**Contents:**
- Function naming conventions with categories
- Error handling patterns (3 approaches)
- Type safety strategies with hierarchy
- Request/response structure details
- Complete template for custom agent endpoints
- Usage examples for all common scenarios
- Best practices summary
- Quick reference table

**Use this when:** You need to understand the reasoning behind patterns, implement new endpoints, or learn best practices.

**Key sections:**
- Section 1: Function naming conventions
- Section 2: Error handling patterns
- Section 3: Type safety approach
- Section 4: Request/response structure
- Section 5: API endpoint structure
- Section 6: Template for custom agents
- Section 7: Usage examples
- Section 8: Best practices

---

### 2. **API_TEMPLATE_CUSTOM_AGENTS.ts** (Copy-Paste Ready)
Production-ready TypeScript template for implementing custom agent endpoints.

**Contents:**
- Type definitions for custom agent operations
- Complete CRUD functions (Create, Read, Update, Delete)
- Action endpoints (Execute, Validate, Clone)
- Batch operations (Execute multiple, Delete multiple)
- Utility functions (Stats, Export, Import)
- Reference error handling patterns
- Full JSDoc documentation with examples
- Usage examples for 3 common scenarios

**Use this when:** You need to add new custom agent endpoints to the API layer.

**Key sections:**
- CRUD operations with full documentation
- Action endpoints (execute, test, clone)
- Batch operations
- Utility functions
- Error handling patterns (reference)
- Working examples

---

### 3. **API_CHEATSHEET.md** (Quick Lookup)
One-page quick reference guide for API patterns.

**Contents:**
- Function naming quick cheat sheet
- 8 copy-paste request patterns
- Type safety quick reference
- Error handling quick reference
- 5 common use cases
- Common mistakes with fixes
- Endpoint structure reference
- Quick method reference table

**Use this when:** You need a quick lookup, forgot a pattern, or need to copy-paste code.

**Key sections:**
- Function naming cheat sheet
- Request pattern cookbook
- Type safety quick reference
- Error handling patterns
- Common use cases
- Common mistakes to avoid
- Endpoint structure
- Method reference table

---

## Quick Start by Use Case

### I want to understand API patterns
1. Read: **API_PATTERNS_ANALYSIS.md** sections 1-4
2. Reference: **API_CHEATSHEET.md** for quick lookup

### I want to implement new custom agent endpoints
1. Copy template from: **API_TEMPLATE_CUSTOM_AGENTS.ts**
2. Reference pattern details in: **API_PATTERNS_ANALYSIS.md** section 6
3. Check examples in: **API_PATTERNS_ANALYSIS.md** section 7

### I need a specific code pattern
1. Check: **API_CHEATSHEET.md** "Request Pattern Cookbook"
2. Adapt from: **API_TEMPLATE_CUSTOM_AGENTS.ts** examples
3. Reference: **API_PATTERNS_ANALYSIS.md** for context

### I'm debugging an API issue
1. Check error handling: **API_CHEATSHEET.md** "Error Handling Quick Reference"
2. Review patterns: **API_PATTERNS_ANALYSIS.md** section 2
3. Check common mistakes: **API_CHEATSHEET.md** "Common Mistakes to Avoid"

### I want to avoid common mistakes
1. Read: **API_CHEATSHEET.md** "Common Mistakes to Avoid"
2. Review: **API_PATTERNS_ANALYSIS.md** section 8

---

## Key Concepts Summary

### Function Naming Pattern
```
[verb][Resource]()
```
Examples: `getAIProviders()`, `createCustomAgent()`, `executeCustomAgent()`

### Request Pattern
```typescript
request<ReturnType>(endpoint, {
  method: 'GET|POST|PUT|DELETE',
  body: { /* payload */ }
})
```

### Error Handling
```typescript
// Automatic (standard endpoints)
try {
  const result = await createCustomAgent(...);
} catch (error) {
  console.error(error.message);  // Pre-formatted
}

// Custom (streaming endpoints)
const response = await fetch(...);
if (!response.ok) throw new Error(...);
return response.body;
```

### Type Safety
```typescript
// Always specify generic type
const agents = await request<CustomAgentsResponse>('/endpoint');
// agents is properly typed

// Discriminated unions prevent invalid state
type Result = { success: true; data: string } | { success: false; error: string };
```

---

## File Structure

```
frontend/src/lib/api/
├── ai/
│   ├── ai.ts          <- Main API functions (patterns demonstrated here)
│   └── types.ts       <- Type definitions
└── base.ts            <- request<T>() wrapper and fetch utilities
```

### Current API Functions (ai.ts)
- `getAIProviders()`
- `updateAIProvider()`
- `getAllModels()`
- `getLocalModels()`
- `pullModel()` - Streaming example
- `warmupModel()`
- `getAISystemInfo()`
- `saveAPIKey()`
- `getAgentPrompts()`
- `getAgentPrompt()`
- `getTools()`
- `executeTool()`
- `getCustomAgents()`

### Types (types.ts)
- `LLMProvider`, `LLMModel`, `AIProvidersResponse`, `AllModelsResponse`
- `AISystemInfo`, `AgentInfo`, `WarmupResponse`
- `Tool`, `ToolResponse` (discriminated union)
- `CustomAgent`, `CustomAgentsResponse`

---

## Implementation Checklist

When adding new custom agent endpoints, follow this checklist:

- [ ] Define types in `types.ts` (or use existing)
- [ ] Write function with proper naming convention
- [ ] Choose request pattern (standard or streaming)
- [ ] Add JSDoc documentation with example
- [ ] Handle errors appropriately
- [ ] Use type-safe generics
- [ ] Test with actual backend endpoint
- [ ] Add to exports if needed

### Example: Implement `createCustomAgent()`

```typescript
// 1. Types already exist (CustomAgent)
// 2. Choose pattern: POST with body (standard request<T>)
// 3. Write function
export async function createCustomAgent(
  agent: Omit<CustomAgent, 'id' | 'user_id' | 'created_at' | 'updated_at'>
) {
  return request<CustomAgent>('/ai/custom-agents', {
    method: 'POST',
    body: agent
  });
}

// 4. Add JSDoc
/**
 * Create a new custom agent
 * @param agent - Agent configuration
 * @returns Promise<CustomAgent> - Created agent with id
 * @example
 * const agent = await createCustomAgent({...});
 */

// 5. Error handling: request<T>() handles automatically
try {
  const agent = await createCustomAgent({...});
} catch (error) {
  // Error already formatted with HTTP status
}
```

---

## Backend Integration Notes

### Field Name Convention
Frontend uses camelCase in TypeScript, backend expects snake_case in JSON:

```typescript
// Frontend
{ apiKey: '...' }

// Backend receives
{ "api_key": "..." }
```

### API Base URL
Determined by `getApiBaseUrl()` in `base.ts`:
- Electron app: Configurable local or cloud URL
- Web app: Environment variable or auto-detect

### Request Credentials
All requests include `credentials: 'include'` for cookie-based auth.

### Content-Type
Automatically added as `application/json` when body is present.

---

## Type Definition Hierarchy

```
Response Types
├── Single Item: CustomAgent
├── Array Wrapper: { agents: CustomAgent[] }
├── Status: { message: string }
├── Complex: { agents: CustomAgent[]; total: number }
└── Union: { success: boolean; result?: string; error?: string }

Request Types
├── Create: Omit<T, 'id' | 'user_id' | 'created_at' | 'updated_at'>
├── Update: Partial<Omit<T, ...>>
├── Batch: { ids: string[]; ...payload }
└── Custom: { prompt: string; context?: Record<string, unknown> }

Data Models
├── LLMProvider: { id, name, type, configured, ... }
├── CustomAgent: { id, user_id, name, system_prompt, ... }
├── AgentInfo: { id, name, description, prompt, category }
└── Tool: { name, description, input_schema, source }
```

---

## Common Patterns Reference

| Pattern | Use Case | Location |
|---------|----------|----------|
| Standard GET | Fetch data | `API_CHEATSHEET.md` Pattern 1-3 |
| Standard POST | Create/Save | `API_CHEATSHEET.md` Pattern 4 |
| Standard PUT | Update | `API_CHEATSHEET.md` Pattern 5 |
| Standard DELETE | Remove | `API_CHEATSHEET.md` Pattern 6 |
| Streaming | Large responses | `API_CHEATSHEET.md` Pattern 7 |
| Batch | Multiple items | `API_CHEATSHEET.md` Pattern 8 |
| Validation | Check before save | `API_TEMPLATE_CUSTOM_AGENTS.ts` |
| Union types | Safe multiple outcomes | `API_PATTERNS_ANALYSIS.md` sec 3 |

---

## Error Message Format

All errors thrown from API functions use consistent format:

```
{ErrorMessage} (HTTP {StatusCode})
```

Examples:
- `"Provider not found (HTTP 404)"`
- `"Unauthorized (HTTP 401)"`
- `"Failed to pull model (HTTP 500)"`

---

## Testing Endpoints

When testing new endpoints:

1. **Unit test the function:**
   ```typescript
   const agent = await createCustomAgent({
     name: 'test-agent',
     display_name: 'Test',
     system_prompt: 'Test prompt'
   });
   expect(agent.id).toBeDefined();
   ```

2. **Test error handling:**
   ```typescript
   try {
     await getCustomAgent('nonexistent-id');
     fail('Should have thrown');
   } catch (error) {
     expect(error.message).toContain('HTTP');
   }
   ```

3. **Test streaming (if applicable):**
   ```typescript
   const stream = await executeCustomAgentStream(id, input);
   const reader = stream?.getReader();
   const { value } = await reader?.read() ?? {};
   expect(value).toBeDefined();
   ```

---

## When to Reference Each Document

| Situation | Document |
|-----------|----------|
| Need complete understanding | `API_PATTERNS_ANALYSIS.md` |
| Implementing new endpoints | `API_TEMPLATE_CUSTOM_AGENTS.ts` |
| Forgot syntax/pattern | `API_CHEATSHEET.md` |
| Want copy-paste template | `API_TEMPLATE_CUSTOM_AGENTS.ts` |
| Checking error handling | `API_PATTERNS_ANALYSIS.md` sec 2 |
| Learning type safety | `API_PATTERNS_ANALYSIS.md` sec 3 |
| Avoiding mistakes | `API_CHEATSHEET.md` Common Mistakes |

---

## Document Maintenance

These documents are manually maintained reference materials. Update them when:
- Adding new endpoint patterns
- Discovering better practices
- Backend API changes
- Type structure changes

Last updated: 2026-01-08
Source files:
- `frontend/src/lib/api/ai/ai.ts`
- `frontend/src/lib/api/ai/types.ts`
- `frontend/src/lib/api/base.ts`

